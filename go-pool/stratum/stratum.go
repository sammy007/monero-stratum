package stratum

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"../pool"
	"../rpc"
	"../util"
)

type StratumServer struct {
	config           *pool.Config
	miners           MinersMap
	blockTemplate    atomic.Value
	upstream         int32
	upstreams        []*rpc.RPCClient
	timeout          time.Duration
	estimationWindow time.Duration
	blocksMu         sync.RWMutex
	blockStats       map[int64]float64
	luckWindow       int64
	luckLargeWindow  int64
	roundShares      int64
	sessionsMu       sync.RWMutex
	sessions         map[*Session]struct{}
}

type Endpoint struct {
	config     *pool.Port
	instanceId []byte
	extraNonce uint32
}

type Session struct {
	sync.Mutex
	conn            *net.TCPConn
	enc             *json.Encoder
	ip              string
	endpoint        *Endpoint
	difficulty      int64
	validJobs       []*Job
	lastBlockHeight int64
	target          uint32
	targetHex       string
}

const (
	MaxReqSize = 10 * 1024
)

func NewStratum(cfg *pool.Config) *StratumServer {
	stratum := &StratumServer{config: cfg}

	stratum.upstreams = make([]*rpc.RPCClient, len(cfg.Upstream))
	for i, v := range cfg.Upstream {
		client, err := rpc.NewRPCClient(&v)
		if err != nil {
			log.Fatal(err)
		} else {
			stratum.upstreams[i] = client
			log.Printf("Upstream: %s => %s", client.Name, client.Url)
		}
	}
	log.Printf("Default upstream: %s => %s", stratum.rpc().Name, stratum.rpc().Url)

	stratum.miners = NewMinersMap()
	stratum.sessions = make(map[*Session]struct{})

	timeout, _ := time.ParseDuration(cfg.Stratum.Timeout)
	stratum.timeout = timeout

	estimationWindow, _ := time.ParseDuration(cfg.EstimationWindow)
	stratum.estimationWindow = estimationWindow

	luckWindow, _ := time.ParseDuration(cfg.LuckWindow)
	stratum.luckWindow = int64(luckWindow / time.Millisecond)
	luckLargeWindow, _ := time.ParseDuration(cfg.LargeLuckWindow)
	stratum.luckLargeWindow = int64(luckLargeWindow / time.Millisecond)

	// Init block template
	stratum.refreshBlockTemplate(false)

	refreshIntv, _ := time.ParseDuration(cfg.BlockRefreshInterval)
	refreshTimer := time.NewTimer(refreshIntv)
	log.Printf("Set block refresh every %v", refreshIntv)

	checkIntv, _ := time.ParseDuration(cfg.UpstreamCheckInterval)
	checkTimer := time.NewTimer(checkIntv)

	go func() {
		for {
			select {
			case <-refreshTimer.C:
				stratum.refreshBlockTemplate(true)
				refreshTimer.Reset(refreshIntv)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-checkTimer.C:
				stratum.checkUpstreams()
				checkTimer.Reset(checkIntv)
			}
		}
	}()

	return stratum
}

func NewEndpoint(cfg *pool.Port) *Endpoint {
	e := &Endpoint{config: cfg}
	e.instanceId = make([]byte, 4)
	_, err := rand.Read(e.instanceId)
	if err != nil {
		log.Fatalf("Can't seed with random bytes: %v", err)
	}
	return e
}

func (s *StratumServer) Listen() {
	quit := make(chan bool)
	for _, port := range s.config.Stratum.Ports {
		go func(cfg pool.Port) {
			e := NewEndpoint(&cfg)
			e.Listen(s)
		}(port)
	}
	<-quit
}

func (e *Endpoint) Listen(s *StratumServer) {
	bindAddr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
	addr, err := net.ResolveTCPAddr("tcp", bindAddr)
	checkError(err)
	server, err := net.ListenTCP("tcp", addr)
	checkError(err)
	defer server.Close()

	log.Printf("Stratum listening on %s", bindAddr)
	accept := make(chan int, e.config.MaxConn)
	n := 0

	for {
		conn, err := server.AcceptTCP()
		if err != nil {
			continue
		}
		conn.SetKeepAlive(true)
		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		cs := &Session{conn: conn, ip: ip, enc: json.NewEncoder(conn), endpoint: e}
		n += 1

		accept <- n
		go func() {
			err = s.handleClient(cs, e)
			if err != nil {
				s.removeSession(cs)
				conn.Close()
			}
			<-accept
		}()
	}
}

func (s *StratumServer) handleClient(cs *Session, e *Endpoint) error {
	_, targetHex := util.GetTargetHex(e.config.Difficulty)
	cs.targetHex = targetHex

	connbuff := bufio.NewReaderSize(cs.conn, MaxReqSize)
	s.setDeadline(cs.conn)

	for {
		data, isPrefix, err := connbuff.ReadLine()
		if isPrefix {
			log.Printf("Socket flood detected")
			return errors.New("Socket flood")
		} else if err == io.EOF {
			log.Println("Client disconnected", cs.ip)
			s.removeSession(cs)
			break
		} else if err != nil {
			log.Printf("Error reading: %v", err)
			return err
		}

		// NOTICE: cpuminer-multi sends junk newlines, so we demand at least 1 byte for decode
		// NOTICE: Ns*CNMiner.exe will send malformed JSON on very low diff, not sure we should handle this
		if len(data) > 1 {
			var req JSONRpcReq
			err = json.Unmarshal(data, &req)
			if err != nil {
				log.Printf("Malformed request: %v", err)
				return err
			}
			s.setDeadline(cs.conn)
			cs.handleMessage(s, e, &req)
		}
	}
	return nil
}

func (cs *Session) handleMessage(s *StratumServer, e *Endpoint, req *JSONRpcReq) {
	if req.Id == nil {
		log.Println("Missing RPC id")
		cs.conn.Close()
		return
	} else if req.Params == nil {
		log.Println("Missing RPC params")
		cs.conn.Close()
		return
	}

	var err error

	// Handle RPC methods
	switch req.Method {
	case "login":
		var params LoginParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Unable to parse params")
			break
		}
		reply, errReply := s.handleLoginRPC(cs, e, &params)
		if errReply != nil {
			err = cs.sendError(req.Id, errReply)
			break
		}
		err = cs.sendResult(req.Id, &reply)
	case "getjob":
		var params GetJobParams
		err = json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Unable to parse params")
			break
		}
		reply, errReply := s.handleGetJobRPC(cs, e, &params)
		if errReply != nil {
			err = cs.sendError(req.Id, errReply)
			break
		}
		err = cs.sendResult(req.Id, &reply)
	case "submit":
		var params SubmitParams
		err := json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Unable to parse params")
			break
		}
		reply, errReply := s.handleSubmitRPC(cs, e, &params)
		if errReply != nil {
			err = cs.sendError(req.Id, errReply)
			break
		}
		err = cs.sendResult(req.Id, &reply)
	default:
		errReply := s.handleUnknownRPC(cs, req)
		err = cs.sendError(req.Id, errReply)
	}

	if err != nil {
		cs.conn.Close()
	}
}

func (cs *Session) sendResult(id *json.RawMessage, result interface{}) error {
	cs.Lock()
	defer cs.Unlock()
	message := JSONRpcResp{Id: id, Version: "2.0", Error: nil, Result: result}
	return cs.enc.Encode(&message)
}

func (cs *Session) pushMessage(method string, params interface{}) error {
	cs.Lock()
	defer cs.Unlock()
	message := JSONPushMessage{Version: "2.0", Method: method, Params: params}
	return cs.enc.Encode(&message)
}

func (cs *Session) sendError(id *json.RawMessage, reply *ErrorReply) error {
	cs.Lock()
	defer cs.Unlock()
	message := JSONRpcResp{Id: id, Version: "2.0", Error: reply}
	err := cs.enc.Encode(&message)
	if reply.Close {
		return errors.New("Force close")
	}
	return err
}

func (s *StratumServer) setDeadline(conn *net.TCPConn) {
	conn.SetDeadline(time.Now().Add(s.timeout))
}

func (s *StratumServer) registerSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	s.sessions[cs] = struct{}{}
}

func (s *StratumServer) removeSession(cs *Session) {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()
	delete(s.sessions, cs)
}

func (s *StratumServer) isActive(cs *Session) bool {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	_, exist := s.sessions[cs]
	return exist
}

func (s *StratumServer) registerMiner(miner *Miner) {
	s.miners.Set(miner.Id, miner)
}

func (s *StratumServer) removeMiner(id string) {
	s.miners.Remove(id)
}

func (s *StratumServer) currentBlockTemplate() *BlockTemplate {
	if t := s.blockTemplate.Load(); t != nil {
		return t.(*BlockTemplate)
	}
	return nil
}

func (s *StratumServer) checkUpstreams() {
	candidate := int32(0)
	backup := false

	for i, v := range s.upstreams {
		ok, err := v.Check(8, s.config.Address)
		if err != nil {
			log.Printf("Upstream %v didn't pass check: %v", v.Name, err)
		}
		if ok && !backup {
			candidate = int32(i)
			backup = true
		}
	}

	if s.upstream != candidate {
		log.Printf("Switching to %v upstream", s.upstreams[candidate].Name)
		atomic.StoreInt32(&s.upstream, candidate)
	}
}

func (s *StratumServer) rpc() *rpc.RPCClient {
	i := atomic.LoadInt32(&s.upstream)
	return s.upstreams[i]
}

func checkError(err error) {
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
