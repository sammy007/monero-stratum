package stratum

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/sammy007/monero-stratum/rpc"
	"github.com/sammy007/monero-stratum/util"
)

func (s *StratumServer) StatsIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

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
	}
	json.NewEncoder(w).Encode(stats)
}

func convertUpstream(u *rpc.RPCClient) map[string]interface{} {
	upstream := map[string]interface{}{
		"name":             u.Name,
		"url":              u.Url.String(),
		"sick":             u.Sick(),
		"accepts":          atomic.LoadInt64(&u.Accepts),
		"rejects":          atomic.LoadInt64(&u.Rejects),
		"lastSubmissionAt": atomic.LoadInt64(&u.LastSubmissionAt),
		"failsCount":       atomic.LoadInt64(&u.FailsCount),
		"info":             u.Info(),
	}
	return upstream
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

func (s *StratumServer) getLuckStats() map[string]interface{} {
	now := util.MakeTimestamp()
	var variance float64
	var totalVariance float64
	var blocksCount int
	var totalBlocksCount int

	s.blocksMu.Lock()
	defer s.blocksMu.Unlock()

	for k, v := range s.blockStats {
		if k >= now-int64(s.luckWindow) {
			blocksCount++
			variance += v.variance
		}
		if k >= now-int64(s.luckLargeWindow) {
			totalBlocksCount++
			totalVariance += v.variance
		} else {
			delete(s.blockStats, k)
		}
	}
	if blocksCount != 0 {
		variance = variance / float64(blocksCount)
	}
	if totalBlocksCount != 0 {
		totalVariance = totalVariance / float64(totalBlocksCount)
	}
	result := make(map[string]interface{})
	result["variance"] = variance
	result["blocksCount"] = blocksCount
	result["window"] = s.config.LuckWindow
	result["totalVariance"] = totalVariance
	result["totalBlocksCount"] = totalBlocksCount
	result["largeWindow"] = s.config.LargeLuckWindow
	return result
}

func (s *StratumServer) getBlocksStats() []interface{} {
	now := util.MakeTimestamp()
	var result []interface{}

	s.blocksMu.Lock()
	defer s.blocksMu.Unlock()

	for k, v := range s.blockStats {
		if k >= now-int64(s.luckLargeWindow) {
			block := map[string]interface{}{
				"height":    v.height,
				"hash":      v.hash,
				"variance":  v.variance,
				"timestamp": k,
			}
			result = append(result, block)
		} else {
			delete(s.blockStats, k)
		}
	}
	return result
}
