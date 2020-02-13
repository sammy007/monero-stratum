package stratum

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/sammy007/monero-stratum/pool"
	"github.com/sammy007/monero-stratum/rpc"
	"github.com/sammy007/monero-stratum/util"
	"io"
	"log"
	"math/big"
	"net"
	"path"
	"sync"
	"sync/atomic"
	"time"
)

type StratumServer struct {
	luckWindow       int64
	luckLargeWindow  int64
	roundShares      int64
	blockStats       map[int64]blockEntry
	config           *pool.Config
	miners           MinersMap
	blockTemplate    atomic.Value
	upstream         int32
	upstreams        []*rpc.RPCClient
	timeout          time.Duration
	estimationWindow time.Duration
	blocksMu         sync.RWMutex
	sessionsMu       sync.RWMutex
	sessions         map[*Session]struct{}
}

type blockEntry struct {
	height   int64
	variance float64
	hash     string
}

type Endpoint struct {
	jobSequence uint64
	config      *pool.Port
	difficulty  *big.Int
	instanceId  []byte
	extraNonce  uint32
	targetHex   string
}

type Session struct {
	lastBlockHeight int64
	sync.Mutex
	conn            *net.TCPConn
	enc             *json.Encoder
	ip              string
	endpoint        *Endpoint
	validJobs       []*Job
}

const (
	MaxReqSize = 10 * 1024
)

const (
	DefaultStatsInterval = 5
)

func NewStratum(cfg *pool.Config) *StratumServer {
	stratum := &StratumServer{config: cfg, blockStats: make(map[int64]blockEntry)}

	// flush stats periodically if configured
	if cfg.StatsDir != "" {
		go func() {
			statsJson := path.Join(cfg.StatsDir, "stats.json")
			importMinerStatsMap(stratum, statsJson)

			// ...work out how often we should save statistics
			statsInterval := cfg.StatsInterval
			if statsInterval == 0 {
				log.Printf("statsInterval not found in config file, defaulting to %d seconds", DefaultStatsInterval)
				statsInterval = 5
			}

			// ...and start a periodic flush process
			for {
				time.Sleep(time.Duration(statsInterval) * time.Second)
				util.SaveJson(statsJson, collectMinerStatsMap(stratum))
			}
		}()
	}

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

	refreshIntv, _ := time.ParseDuration(cfg.BlockRefreshInterval)
	refreshTimer := time.NewTimer(refreshIntv)
	log.Printf("Set block refresh every %v", refreshIntv)

	checkIntv, _ := time.ParseDuration(cfg.UpstreamCheckInterval)
	checkTimer := time.NewTimer(checkIntv)

	infoIntv, _ := time.ParseDuration(cfg.UpstreamCheckInterval)
	infoTimer := time.NewTimer(infoIntv)

	// Init block template
	go stratum.refreshBlockTemplate(false)

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

	go func() {
		for {
			select {
			case <-infoTimer.C:
				poll := func(v *rpc.RPCClient) {
					_, err := v.UpdateInfo()
					if err != nil {
						log.Printf("Unable to update info on upstream %s: %v", v.Name, err)
					}
				}
				current := stratum.rpc()
				poll(current)

				// Async rpc call to not block on rpc timeout, ignoring current
				go func() {
					for _, v := range stratum.upstreams {
						if v != current {
							poll(v)
						}
					}
				}()
				infoTimer.Reset(infoIntv)
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
	e.targetHex = util.GetTargetHex(e.config.Difficulty)
	e.difficulty = big.NewInt(e.config.Difficulty)
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
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	server, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
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
			s.handleClient(cs, e)
			<-accept
		}()
	}
}

func (s *StratumServer) handleClient(cs *Session, e *Endpoint) {
	connbuff := bufio.NewReaderSize(cs.conn, MaxReqSize)
	s.setDeadline(cs.conn)

	for {
		data, isPrefix, err := connbuff.ReadLine()
		if isPrefix {
			log.Println("Socket flood detected from", cs.ip)
			break
		} else if err == io.EOF {
			log.Println("Client disconnected", cs.ip)
			break
		} else if err != nil {
			log.Println("Error reading:", err)
			break
		}

		// NOTICE: cpuminer-multi sends junk newlines, so we demand at least 1 byte for decode
		// NOTICE: Ns*CNMiner.exe will send malformed JSON on very low diff, not sure we should handle this
		if len(data) > 1 {
			var req JSONRpcReq
			err = json.Unmarshal(data, &req)
			if err != nil {
				log.Printf("Malformed request from %s: %v", cs.ip, err)
				break
			}
			s.setDeadline(cs.conn)
			err = cs.handleMessage(s, e, &req)
			if err != nil {
				break
			}
		}
	}
	s.removeSession(cs)
	cs.conn.Close()
}

func (cs *Session) handleMessage(s *StratumServer, e *Endpoint, req *JSONRpcReq) error {
	if req.Id == nil {
		err := fmt.Errorf("Server disconnect request")
		log.Println(err)
		return err
	} else if req.Params == nil {
		err := fmt.Errorf("Server RPC request params")
		log.Println(err)
		return err
	}

	// Handle RPC methods
	switch req.Method {

	case "login":
		var params LoginParams
		err := json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Unable to parse params")
			return err
		}
		reply, errReply := s.handleLoginRPC(cs, &params)
		if errReply != nil {
			return cs.sendError(req.Id, errReply, true)
		}
		return cs.sendResult(req.Id, &reply)
	case "getjob":
		var params GetJobParams
		err := json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Unable to parse params")
			return err
		}
		reply, errReply := s.handleGetJobRPC(cs, &params)
		if errReply != nil {
			return cs.sendError(req.Id, errReply, true)
		}
		return cs.sendResult(req.Id, &reply)
	case "submit":
		var params SubmitParams
		err := json.Unmarshal(*req.Params, &params)
		if err != nil {
			log.Println("Unable to parse params")
			return err
		}
		reply, errReply := s.handleSubmitRPC(cs, &params)
		if errReply != nil {
			return cs.sendError(req.Id, errReply, false)
		}
		return cs.sendResult(req.Id, &reply)
	case "keepalived":
		return cs.sendResult(req.Id, &StatusReply{Status: "KEEPALIVED"})
	default:
		errReply := s.handleUnknownRPC(req)
		return cs.sendError(req.Id, errReply, true)
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

func (cs *Session) sendError(id *json.RawMessage, reply *ErrorReply, drop bool) error {
	cs.Lock()
	defer cs.Unlock()
	message := JSONRpcResp{Id: id, Version: "2.0", Error: reply}
	err := cs.enc.Encode(&message)
	if err != nil {
		return err
	}
	if drop {
		return fmt.Errorf("Server disconnect request")
	}
	return nil
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
	s.miners.Set(miner.id, miner)
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

func (s *StratumServer) collectMinersStats() (float64, float64, int, []interface{}) {
	now := util.MakeTimestamp()
	var result []interface{}
	totalhashrate := float64(0)
	totalhashrate24h := float64(0)
	totalOnline := 0
	window24h := 24 * time.Hour

	for m := range s.miners.Iter() {
		stats := make(map[string]interface{})
		lastBeat := m.Val.getLastBeat()
		hashrate := m.Val.hashrate(s.estimationWindow)
		hashrate24h := m.Val.hashrate(window24h)
		totalhashrate += hashrate
		totalhashrate24h += hashrate24h
		stats["name"] = m.Key
		stats["hashrate"] = hashrate
		stats["hashrate24h"] = hashrate24h
		stats["lastBeat"] = lastBeat
		stats["validShares"] = atomic.LoadInt64(&m.Val.validShares)
		stats["staleShares"] = atomic.LoadInt64(&m.Val.staleShares)
		stats["invalidShares"] = atomic.LoadInt64(&m.Val.invalidShares)
		stats["accepts"] = atomic.LoadInt64(&m.Val.accepts)
		stats["rejects"] = atomic.LoadInt64(&m.Val.rejects)
		if !s.config.Frontend.HideIP {
			stats["ip"] = m.Val.ip
		}

		if now-lastBeat > (int64(s.timeout/2) / 1000000) {
			stats["warning"] = true
		}
		if now-lastBeat > (int64(s.timeout) / 1000000) {
			stats["timeout"] = true
		} else {
			totalOnline++
		}
		result = append(result, stats)
	}
	return totalhashrate, totalhashrate24h, totalOnline, result
}

func collectMinerStatsMap(s *StratumServer) map[string]interface{} {
	hashrate, hashrate24h, totalOnline, miners := s.collectMinersStats()
	stats := map[string]interface{}{
		"miners":      miners,
		"hashrate":    hashrate,
		"hashrate24h": hashrate24h,
		"totalMiners": len(miners),
		"totalOnline": totalOnline,
		"timedOut":    len(miners) - totalOnline,
		"now":         util.MakeTimestamp(),
	}

	var upstreams []interface{}
	current := atomic.LoadInt32(&s.upstream)

	for i, u := range s.upstreams {
		upstream := convertUpstream(u)
		upstream["current"] = current == int32(i)
		upstreams = append(upstreams, upstream)
	}
	stats["upstreams"] = upstreams
	stats["current"] = convertUpstream(s.rpc())
	stats["luck"] = s.getLuckStats()
	stats["blocks"] = s.getBlocksStats()

	if t := s.currentBlockTemplate(); t != nil {
		stats["height"] = t.height
		stats["diff"] = t.diffInt64
		roundShares := atomic.LoadInt64(&s.roundShares)
		stats["variance"] = float64(roundShares) / float64(t.diffInt64)
		stats["prevHash"] = t.prevHash[0:8]
		stats["template"] = true

		// the overall effort towards the current block "variance" is an OUTPUT of the calculation of
		// roundShares/networkDifficulty - therefore it makes no sense to save it in JSON as its a calculated output
		// we must persist the value of roundShares instead, as an extra field
		stats["roundShares"] = roundShares
	}

	return stats
}


func importMinerStatsMap(stratumServer *StratumServer, statsJson string)  {
	log.Println("importing previous statistics...")
	if parsed := util.LoadJson(statsJson) ; parsed != nil {
		if stats, ok := parsed.(map[string]interface{}); ok {
			// mined blocks
			setBlockStats(stratumServer, stats["blocks"])

			// progress
			if v, ok := stats["roundShares"].(json.Number); ok {
				stratumServer.roundShares, _ = v.Int64()
			}

		} else {
			log.Println("Parsed JSON from saved stats but its unreadable", parsed)
		}
	}
}