package stratum

import (
	"encoding/json"
	"fmt"
	"github.com/sammy007/monero-stratum/rpc"
	"github.com/sammy007/monero-stratum/util"
	"log"
	"net/http"
	"sync/atomic"
)

func (s *StratumServer) StatsIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(collectMinerStatsMap(s))
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

func setBlockStats(s *StratumServer, importBlocksBlob interface{})  {
	s.blocksMu.Lock()
	defer s.blocksMu.Unlock()

	if importBlocks, ok := importBlocksBlob.([]interface{}) ; ok {
		// blocks are stored in JSON as array of hash but in the our struct as timestamp -> blockEntry
		for _, element := range importBlocks {
			if importBlock, ok := element.(map[string]interface{}) ; ok{
				block := blockEntry{}
				if d, ok := importBlock["height"].(json.Number); ok {
					block.height, _ = d.Int64()
				}
				if d, ok := importBlock["hash"].(string); ok {
					block.hash = d
				}
				if d, ok := importBlock["variance"].(json.Number); ok {
					block.variance, _ = d.Float64()
				}
				if d, ok := importBlock["timestamp"].(json.Number); ok {
					timestamp, _ := d.Int64()
					s.blockStats[timestamp] = block
					log.Printf("Imported block %d OK!", block.height)
				} else {
					log.Printf("Skipped importing a block... timestamp dectected as %T but should be int64! got value: '%s'", importBlock["timestamp"], importBlock["timestamp"])
				}
			}
		}
	} else {
		log.Println("Unable to import any blocks... *ALL* of the JSON is invalid!", importBlocksBlob)
		log.Println(fmt.Sprintf("detected type: %T", importBlocksBlob))

		//[]interface {}

	}

}