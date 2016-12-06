package stratum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"../../cnutil"
	"../../hashing"
	"../util"
)

type Job struct {
	sync.RWMutex
	id          string
	extraNonce  uint32
	height      int64
	difficulty  int64
	submissions map[string]struct{}
}

type Miner struct {
	sync.RWMutex
	Id       string
	IP       string
	LastBeat int64
	Endpoint *Endpoint

	startedAt     int64
	validShares   uint64
	invalidShares uint64
	staleShares   uint64
	accepts       uint64
	rejects       uint64
	shares        map[int64]int64
}

func (job *Job) submit(nonce string) bool {
	job.Lock()
	defer job.Unlock()
	_, exist := job.submissions[nonce]
	if exist {
		return true
	}
	job.submissions[nonce] = struct{}{}
	return false
}

func NewMiner(id string, ip string) *Miner {
	shares := make(map[int64]int64)
	return &Miner{Id: id, IP: ip, shares: shares}
}

func (cs *Session) getJob(t *BlockTemplate) *JobReplyData {
	height := atomic.SwapInt64(&cs.lastBlockHeight, t.Height)

	if height == t.Height {
		return &JobReplyData{}
	}

	extraNonce := atomic.AddUint32(&cs.endpoint.extraNonce, 1)
	blob := t.nextBlob(extraNonce, cs.endpoint.instanceId)
	job := &Job{id: util.Random(), extraNonce: extraNonce, height: t.Height, difficulty: cs.difficulty}
	job.submissions = make(map[string]struct{})
	cs.pushJob(job)
	reply := &JobReplyData{JobId: job.id, Blob: blob, Target: cs.targetHex}
	return reply
}

func (cs *Session) pushJob(job *Job) {
	cs.Lock()
	defer cs.Unlock()
	cs.validJobs = append(cs.validJobs, job)

	if len(cs.validJobs) > 4 {
		cs.validJobs = cs.validJobs[1:]
	}
}

func (cs *Session) findJob(id string) *Job {
	cs.Lock()
	defer cs.Unlock()
	for _, job := range cs.validJobs {
		if job.id == id {
			return job
		}
	}
	return nil
}

func (m *Miner) heartbeat() {
	now := util.MakeTimestamp()
	atomic.StoreInt64(&m.LastBeat, now)
}

func (m *Miner) getLastBeat() int64 {
	return atomic.LoadInt64(&m.LastBeat)
}

func (m *Miner) storeShare(diff int64) {
	now := util.MakeTimestamp() / 1000
	m.Lock()
	m.shares[now] += diff
	m.Unlock()
}

func (m *Miner) hashrate(estimationWindow time.Duration) float64 {
	now := util.MakeTimestamp() / 1000
	totalShares := int64(0)
	window := int64(estimationWindow / time.Second)
	boundary := now - m.startedAt

	if boundary > window {
		boundary = window
	}

	m.Lock()
	for k, v := range m.shares {
		if k < now-86400 {
			delete(m.shares, k)
		} else if k >= now-window {
			totalShares += v
		}
	}
	m.Unlock()
	return float64(totalShares) / float64(boundary)
}

func (m *Miner) processShare(s *StratumServer, e *Endpoint, job *Job, t *BlockTemplate, nonce string, result string) bool {
	r := s.rpc()

	shareBuff := make([]byte, len(t.Buffer))
	copy(shareBuff, t.Buffer)
	copy(shareBuff[t.ReservedOffset+4:t.ReservedOffset+7], e.instanceId)

	extraBuff := new(bytes.Buffer)
	binary.Write(extraBuff, binary.BigEndian, job.extraNonce)
	copy(shareBuff[t.ReservedOffset:], extraBuff.Bytes())

	nonceBuff, _ := hex.DecodeString(nonce)
	copy(shareBuff[39:], nonceBuff)

	var hashBytes, convertedBlob []byte

	if s.config.BypassShareValidation {
		hashBytes, _ = hex.DecodeString(result)
	} else {
		convertedBlob = cnutil.ConvertBlob(shareBuff)
		hashBytes = hashing.Hash(convertedBlob, false)
	}

	if !s.config.BypassShareValidation && hex.EncodeToString(hashBytes) != result {
		log.Printf("Bad hash from miner %v@%v", m.Id, m.IP)
		atomic.AddUint64(&m.invalidShares, 1)
		return false
	}

	hashDiff := util.GetHashDifficulty(hashBytes).Int64() // FIXME: Will return max int64 value if overflows
	atomic.AddInt64(&s.roundShares, e.config.Difficulty)
	atomic.AddUint64(&m.validShares, 1)
	m.storeShare(e.config.Difficulty)

	block := hashDiff >= t.Difficulty

	if block {
		_, err := r.SubmitBlock(hex.EncodeToString(shareBuff))
		if err != nil {
			atomic.AddUint64(&m.rejects, 1)
			atomic.AddUint64(&r.Rejects, 1)
			log.Printf("Block submission failure at height %v: %v", t.Height, err)
		} else {
			if len(convertedBlob) == 0 {
				convertedBlob = cnutil.ConvertBlob(shareBuff)
			}
			blockFastHash := hex.EncodeToString(hashing.FastHash(convertedBlob))
			// Immediately refresh current BT and send new jobs
			s.refreshBlockTemplate(true)
			atomic.AddUint64(&m.accepts, 1)
			atomic.AddUint64(&r.Accepts, 1)
			atomic.StoreInt64(&r.LastSubmissionAt, util.MakeTimestamp())
			log.Printf("Block %v found at height %v by miner %v@%v", blockFastHash[0:6], t.Height, m.Id, m.IP)
		}
	} else if hashDiff < job.difficulty {
		log.Printf("Rejected low difficulty share of %v from %v@%v", hashDiff, m.Id, m.IP)
		atomic.AddUint64(&m.invalidShares, 1)
		return false
	}

	log.Printf("Valid share at difficulty %v/%v", e.config.Difficulty, hashDiff)
	return true
}
