package stratum

import (
	"log"
	"regexp"
	"strings"
	"sync/atomic"

	"../util"
)

var noncePattern *regexp.Regexp

const defaultWorkerId = "0"

func init() {
	noncePattern, _ = regexp.Compile("^[0-9a-f]{8}$")
}

func (s *StratumServer) handleLoginRPC(cs *Session, e *Endpoint, params *LoginParams) (*JobReply, *ErrorReply) {
	if !s.config.BypassAddressValidation && !util.ValidateAddress(params.Login, s.config.Address) {
		return nil, &ErrorReply{Code: -1, Message: "Invalid address used for login", Close: true}
	}

	t := s.currentBlockTemplate()
	if t == nil {
		return nil, &ErrorReply{Code: -1, Message: "Job not ready", Close: true}
	}

	id := extractWorkerId(params.Login)
	miner, ok := s.miners.Get(id)
	if !ok {
		miner = NewMiner(id, cs.ip)
		s.registerMiner(miner)
	}

	log.Printf("Miner connected %s@%s", id, cs.ip)

	s.registerSession(cs)
	miner.heartbeat()

	return &JobReply{Id: id, Job: cs.getJob(t), Status: "OK"}, nil
}

func (s *StratumServer) handleGetJobRPC(cs *Session, e *Endpoint, params *GetJobParams) (reply *JobReplyData, errorReply *ErrorReply) {
	miner, ok := s.miners.Get(params.Id)
	if !ok {
		errorReply = &ErrorReply{Code: -1, Message: "Unauthenticated", Close: true}
		return
	}
	t := s.currentBlockTemplate()
	if t == nil {
		errorReply = &ErrorReply{Code: -1, Message: "Job not ready", Close: true}
		return
	}
	miner.heartbeat()
	reply = cs.getJob(t)
	return
}

func (s *StratumServer) handleSubmitRPC(cs *Session, e *Endpoint, params *SubmitParams) (reply *SubmitReply, errorReply *ErrorReply) {
	miner, ok := s.miners.Get(params.Id)
	if !ok {
		errorReply = &ErrorReply{Code: -1, Message: "Unauthenticated", Close: true}
		return
	}
	miner.heartbeat()

	job := cs.findJob(params.JobId)
	if job == nil {
		errorReply = &ErrorReply{Code: -1, Message: "Invalid job id", Close: true}
		atomic.AddUint64(&miner.invalidShares, 1)
		return
	}

	if !noncePattern.MatchString(params.Nonce) {
		errorReply = &ErrorReply{Code: -1, Message: "Malformed nonce", Close: true}
		atomic.AddUint64(&miner.invalidShares, 1)
		return
	}
	nonce := strings.ToLower(params.Nonce)
	exist := job.submit(nonce)
	if exist {
		errorReply = &ErrorReply{Code: -1, Message: "Duplicate share", Close: true}
		atomic.AddUint64(&miner.invalidShares, 1)
		return
	}

	t := s.currentBlockTemplate()
	if job.height != t.Height {
		log.Printf("Block expired for height %v %s@%s", job.height, miner.Id, miner.IP)
		errorReply = &ErrorReply{Code: -1, Message: "Block expired", Close: false}
		atomic.AddUint64(&miner.staleShares, 1)
		return
	}

	validShare := miner.processShare(s, e, job, t, nonce, params.Result)
	if !validShare {
		errorReply = &ErrorReply{Code: -1, Message: "Low difficulty share", Close: !ok}
		return
	}

	reply = &SubmitReply{Status: "OK"}
	return
}

func (s *StratumServer) handleUnknownRPC(cs *Session, req *JSONRpcReq) *ErrorReply {
	log.Printf("Unknown RPC method: %v", req)
	return &ErrorReply{Code: -1, Message: "Invalid method", Close: true}
}

func (s *StratumServer) broadcastNewJobs() {
	t := s.currentBlockTemplate()
	if t == nil {
		return
	}
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()
	count := len(s.sessions)
	log.Printf("Broadcasting new jobs to %d miners", count)
	bcast := make(chan int, 1024*16)
	n := 0

	for m := range s.sessions {
		n++
		bcast <- n
		go func(cs *Session) {
			reply := cs.getJob(t)
			err := cs.pushMessage("job", &reply)
			<-bcast
			if err != nil {
				log.Printf("Job transmit error to %s: %v", cs.ip, err)
				s.removeSession(cs)
			} else {
				s.setDeadline(cs.conn)
			}
		}(m)
	}
}

func (s *StratumServer) refreshBlockTemplate(bcast bool) {
	newBlock := s.fetchBlockTemplate()
	if newBlock && bcast {
		s.broadcastNewJobs()
	}
}

func extractWorkerId(loginWorkerPair string) string {
	parts := strings.SplitN(loginWorkerPair, ".", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return defaultWorkerId
}
