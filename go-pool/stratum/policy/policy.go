package policy

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"../../pool"
	"../../storage"
	"../../util"
)

type Stats struct {
	sync.Mutex
	ValidShares   uint32
	InvalidShares uint32
	Malformed     uint32
	ConnLimit     int32
	FailsCount    uint32
	LastBeat      int64
	Banned        uint32
	BannedAt      int64
}

type PolicyServer struct {
	sync.RWMutex
	config     *pool.Policy
	stats      StatsMap
	banChannel chan string
	startedAt  int64
	grace      int64
	timeout    int64
	blacklist  []string
	whitelist  []string
	storage    *storage.RedisClient
}

func Start(cfg *pool.Config, storage *storage.RedisClient) *PolicyServer {
	s := &PolicyServer{config: &cfg.Policy, startedAt: util.MakeTimestamp()}
	grace, _ := time.ParseDuration(cfg.Policy.Limits.Grace)
	s.grace = int64(grace / time.Millisecond)
	s.banChannel = make(chan string, 64)
	s.stats = NewStatsMap()
	s.storage = storage
	s.refreshState()

	timeout, _ := time.ParseDuration(s.config.ResetInterval)
	s.timeout = int64(timeout / time.Millisecond)

	resetIntv, _ := time.ParseDuration(s.config.ResetInterval)
	resetTimer := time.NewTimer(resetIntv)
	log.Printf("Set policy stats reset every %v", resetIntv)

	refreshIntv, _ := time.ParseDuration(s.config.RefreshInterval)
	refreshTimer := time.NewTimer(refreshIntv)
	log.Printf("Set policy state refresh every %v", refreshIntv)

	go func() {
		for {
			select {
			case <-resetTimer.C:
				s.resetStats()
				resetTimer.Reset(resetIntv)
			case <-refreshTimer.C:
				s.refreshState()
				refreshTimer.Reset(refreshIntv)
			}
		}
	}()

	for i := 0; i < s.config.Workers; i++ {
		s.startPolicyWorker()
	}
	log.Printf("Running with %v policy workers", s.config.Workers)
	return s
}

func (s *PolicyServer) startPolicyWorker() {
	go func() {
		for {
			select {
			case ip := <-s.banChannel:
				s.doBan(ip)
			}
		}
	}()
}

func (s *PolicyServer) resetStats() {
	now := util.MakeTimestamp()
	banningTimeout := s.config.Banning.Timeout * 1000
	total := 0

	for m := range s.stats.IterBuffered() {
		lastBeat := atomic.LoadInt64(&m.Val.LastBeat)
		bannedAt := atomic.LoadInt64(&m.Val.BannedAt)

		if now-bannedAt >= banningTimeout {
			atomic.StoreInt64(&m.Val.BannedAt, 0)
			if atomic.CompareAndSwapUint32(&m.Val.Banned, 1, 0) {
				log.Printf("Ban dropped for %v", m.Key)
			}
		}
		if now-lastBeat >= s.timeout {
			s.stats.Remove(m.Key)
			total++
		}
	}
	log.Printf("Flushed stats for %v IP addresses", total)
}

func (s *PolicyServer) refreshState() {
	s.Lock()
	defer s.Unlock()

	s.blacklist = s.storage.GetBlacklist()
	s.whitelist = s.storage.GetWhitelist()
	log.Println("Policy state refresh complete")
}

func (s *PolicyServer) NewStats() *Stats {
	x := &Stats{
		ConnLimit: s.config.Limits.Limit,
		Malformed: s.config.Banning.MalformedLimit,
	}
	x.heartbeat()
	return x
}

func (s *PolicyServer) Get(ip string) *Stats {
	if x, ok := s.stats.Get(ip); ok {
		x.heartbeat()
		return x
	}
	x := s.NewStats()
	s.stats.Set(ip, x)
	return x
}

func (s *PolicyServer) ApplyLimitPolicy(ip string) bool {
	if !s.config.Limits.Enabled {
		return true
	}
	now := util.MakeTimestamp()
	if now-s.startedAt > s.grace {
		return s.Get(ip).decrLimit() > 0
	}
	return true
}

func (s *PolicyServer) ApplyLoginPolicy(addy, ip string) bool {
	if s.InBlackList(addy) {
		x := s.Get(ip)
		s.forceBan(x, ip)
		return false
	}
	return true
}

func (s *PolicyServer) ApplyMalformedPolicy(ip string) {
	x := s.Get(ip)
	n := x.incrMalformed()
	if n >= s.config.Banning.MalformedLimit {
		s.forceBan(x, ip)
	}
}

func (s *PolicyServer) ApplySharePolicy(ip string, validShare bool) bool {
	x := s.Get(ip)
	if validShare && s.config.Limits.Enabled {
		s.Get(ip).incrLimit(s.config.Limits.LimitJump)
	}
	x.Lock()

	if validShare {
		x.ValidShares++
		if s.config.Limits.Enabled {
			x.incrLimit(s.config.Limits.LimitJump)
		}
	} else {
		x.InvalidShares++
	}

	totalShares := x.ValidShares + x.InvalidShares
	if totalShares < s.config.Banning.CheckThreshold {
		x.Unlock()
		return true
	}
	validShares := float32(x.ValidShares)
	invalidShares := float32(x.InvalidShares)
	x.resetShares()
	x.Unlock()

	if invalidShares == 0 {
		return true
	}

	// Can be +Inf or value, previous check prevents NaN
	ratio := invalidShares / validShares

	if ratio >= s.config.Banning.InvalidPercent/100.0 {
		s.forceBan(x, ip)
		return false
	}
	return true
}

func (x *Stats) resetShares() {
	x.ValidShares = 0
	x.InvalidShares = 0
}

func (s *PolicyServer) forceBan(x *Stats, ip string) {
	if !s.config.Banning.Enabled || s.InWhiteList(ip) {
		return
	}

	if atomic.CompareAndSwapUint32(&x.Banned, 0, 1) {
		if len(s.config.Banning.IPSet) > 0 {
			s.banChannel <- ip
		}
	}
}

func (x *Stats) incrLimit(n int32) {
	atomic.AddInt32(&x.ConnLimit, n)
}

func (x *Stats) incrMalformed() uint32 {
	return atomic.AddUint32(&x.Malformed, 1)
}

func (x *Stats) decrLimit() int32 {
	return atomic.AddInt32(&x.ConnLimit, -1)
}

func (s *PolicyServer) InBlackList(addy string) bool {
	s.RLock()
	defer s.RUnlock()
	return util.StringInSlice(addy, s.blacklist)
}

func (s *PolicyServer) InWhiteList(ip string) bool {
	s.RLock()
	defer s.RUnlock()
	return util.StringInSlice(ip, s.whitelist)
}

func (s *PolicyServer) doBan(ip string) {
	set, timeout := s.config.Banning.IPSet, s.config.Banning.Timeout
	cmd := fmt.Sprintf("sudo ipset add %s %s timeout %v -!", set, ip, timeout)
	args := strings.Fields(cmd)
	head := args[0]
	args = args[1:]

	log.Printf("Banned %v with timeout %v", ip, timeout)

	_, err := exec.Command(head, args...).Output()
	if err != nil {
		log.Printf("CMD Error: %s", err)
	}
}

func (x *Stats) heartbeat() {
	now := util.MakeTimestamp()
	atomic.StoreInt64(&x.LastBeat, now)
}
