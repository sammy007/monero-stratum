package stratum

import (
	"log"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"../util"
)

var noncePattern *regexp.Regexp

const defaultWorkerId = "0"

func init() {
	noncePattern, _ = regexp.Compile("^[0-9a-f]{8}$")
}

func (s *StratumServer) handleLoginRPC(cs *Session, params *LoginParams) (*JobReply, *ErrorReply) {
	address, id := extractWorkerId(params.Login)
	if !s.config.BypassAddressValidation && !util.ValidateAddress(address, s.config.Address) {
		log.Printf("Invalid address %s used for login by %s", address, cs.ip)
		return nil, &ErrorReply{Code: -1, Message: "Invalid address used for login"}
	}

	t := s.currentBlockTemplate()
	if t == nil {
		return nil, &ErrorReply{Code: -1, Message: "Job not ready"}
	}

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

func (s *StratumServer) handleGetJobRPC(cs *Session, params *GetJobParams) (*JobReplyData, *ErrorReply) {
	miner, ok := s.miners.Get(params.Id)
	if !ok {
		return nil, &ErrorReply{Code: -1, Message: "Unauthenticated"}
	}
	t := s.currentBlockTemplate()
	if t == nil {
		return nil, &ErrorReply{Code: -1, Message: "Job not ready"}
	}
	miner.heartbeat()
	return cs.getJob(t), nil
}

func (s *StratumServer) handleSubmitRPC(cs *Session, params *SubmitParams) (*SubmitReply, *ErrorReply) {
	miner, ok := s.miners.Get(params.Id)
	if !ok {
		return nil, &ErrorReply{Code: -1, Message: "Unauthenticated"}
	}
	miner.heartbeat()

	job := cs.findJob(params.JobId)
	if job == nil {
		return nil, &ErrorReply{Code: -1, Message: "Invalid job id"}
	}

	if !noncePattern.MatchString(params.Nonce) {
		return nil, &ErrorReply{Code: -1, Message: "Malformed nonce"}
	}
	nonce := strings.ToLower(params.Nonce)
	exist := job.submit(nonce)
	if exist {
		atomic.AddUint64(&miner.invalidShares, 1)
		return nil, &ErrorReply{Code: -1, Message: "Duplicate share"}
	}

	t := s.currentBlockTemplate()
	if job.height != t.height {
		log.Printf("Stale share for height %d from %s@%s", job.height, miner.id, cs.ip)
		atomic.AddUint64(&miner.staleShares, 1)
		return nil, &ErrorReply{Code: -1, Message: "Block expired"}
	}

	validShare := miner.processShare(s, cs, job, t, nonce, params.Result)
	if !validShare {
		return nil, &ErrorReply{Code: -1, Message: "Low difficulty share"}
	}
	return &SubmitReply{Status: "OK"}, nil
}

func (s *StratumServer) handleUnknownRPC(req *JSONRpcReq) *ErrorReply {
	log.Printf("Unknown RPC method: %v", req)
	return &ErrorReply{Code: -1, Message: "Invalid method"}
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
	start := time.Now()
	slots := make(chan bool, 1024*16)
	var ok, fails int64
	var wg sync.WaitGroup

	for m := range s.sessions {
		wg.Add(1)
		slots <- true

		go func(cs *Session) {
			reply := cs.getJob(t)
			err := cs.pushMessage("job", &reply)
			<-slots

			if err != nil {
				log.Printf("Job transmit error to %s: %v", cs.ip, err)
				atomic.AddInt64(&fails, 1)
				wg.Done()
				s.removeSession(cs)
			} else {
				atomic.AddInt64(&ok, 1)
				s.setDeadline(cs.conn)
				wg.Done()
			}
		}(m)
	}
	wg.Wait()
	log.Printf("Done jobs broadcast in %s for %d/%d/%d miners", time.Since(start), count, ok, fails)
}

func (s *StratumServer) refreshBlockTemplate(bcast bool) {
	newBlock := s.fetchBlockTemplate()
	if newBlock && bcast {
		s.broadcastNewJobs()
	}
}

func extractWorkerId(loginWorkerPair string) (string, string) {
	parts := strings.SplitN(loginWorkerPair, ".", 2)
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return loginWorkerPair, defaultWorkerId
}
