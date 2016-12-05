package stratum

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"
	"sync"
	"sync/atomic"

	"../../cnutil"
	"../../hashing"
	"../util"
)

type Job struct {
	sync.RWMutex
	Id          string
	ExtraNonce  uint32
	Height      int64
	Difficulty  int64
	Submissions map[string]bool
}

type Miner struct {
	sync.RWMutex
	Id              string
	Login           string
	Pass            string
	IP              string
	Difficulty      int64
	ValidJobs       []*Job
	LastBlockHeight int64
	Target          uint32
	TargetHex       string
	LastBeat        int64
	Session         *Session
}

func (job *Job) submit(nonce string) bool {
	job.Lock()
	defer job.Unlock()
	_, exist := job.Submissions[nonce]
	if exist {
		return true
	}
	job.Submissions[nonce] = true
	return false
}

func NewMiner(login, pass string, diff int64, ip string) *Miner {
	id := util.Random()
	miner := &Miner{Id: id, Login: login, Pass: pass, Difficulty: diff, IP: ip}
	target, targetHex := util.GetTargetHex(diff)
	miner.Target = target
	miner.TargetHex = targetHex
	return miner
}

func (m *Miner) pushJob(job *Job) {
	m.Lock()
	defer m.Unlock()
	m.ValidJobs = append(m.ValidJobs, job)

	if len(m.ValidJobs) > 4 {
		m.ValidJobs = m.ValidJobs[1:]
	}
}

func (m *Miner) getJob(s *StratumServer) *JobReplyData {
	t := s.currentBlockTemplate()
	height := atomic.SwapInt64(&m.LastBlockHeight, t.Height)

	if height == t.Height {
		return &JobReplyData{}
	}

	blob, extraNonce := t.nextBlob()
	job := &Job{Id: util.Random(), ExtraNonce: extraNonce, Height: t.Height, Difficulty: m.Difficulty}
	job.Submissions = make(map[string]bool)
	m.pushJob(job)
	reply := &JobReplyData{JobId: job.Id, Blob: blob, Target: m.TargetHex}
	return reply
}

func (m *Miner) heartbeat() {
	now := util.MakeTimestamp()
	atomic.StoreInt64(&m.LastBeat, now)
}

func (m *Miner) findJob(id string) *Job {
	m.RLock()
	defer m.RUnlock()
	for _, job := range m.ValidJobs {
		if job.Id == id {
			return job
		}
	}
	return nil
}

func (m *Miner) processShare(s *StratumServer, job *Job, t *BlockTemplate, nonce string, result string) bool {
	shareBuff := make([]byte, len(t.Buffer))
	copy(shareBuff, t.Buffer)

	extraBuff := new(bytes.Buffer)
	binary.Write(extraBuff, binary.BigEndian, job.ExtraNonce)
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
		log.Printf("Bad hash from miner %v@%v", m.Login, m.IP)
		return false
	}

	hashDiff := util.GetHashDifficulty(hashBytes).Int64() // FIXME: Will return max int64 value if overflows
	block := hashDiff >= t.Difficulty
	if block {
		_, err := s.rpc.SubmitBlock(hex.EncodeToString(shareBuff))
		if err != nil {
			log.Printf("Block submission failure at height %v: %v", t.Height, err)
		} else {
			if len(convertedBlob) == 0 {
				convertedBlob = cnutil.ConvertBlob(shareBuff)
			}
			blockFastHash := hex.EncodeToString(hashing.FastHash(convertedBlob))
			// Immediately refresh current BT and send new jobs
			s.refreshBlockTemplate(true)
			log.Printf("Block %v found at height %v by miner %v@%v", blockFastHash[0:6], t.Height, m.Login, m.IP)
		}
	} else if hashDiff < job.Difficulty {
		log.Printf("Rejected low difficulty share of %v from %v@%v", hashDiff, m.Login, m.IP)
		return false
	}

	log.Printf("Valid share at difficulty %v/%v", s.port.Difficulty, hashDiff)
	return true
}
