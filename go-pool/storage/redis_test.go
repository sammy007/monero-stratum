package storage

import (
	"gopkg.in/redis.v3"
	"os"
	"reflect"
	"strings"
	"testing"
)

var r *RedisClient

func TestMain(m *testing.M) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	r = &RedisClient{client: client, prefix: "test"}
	r.client.FlushAll()

	os.Exit(m.Run())
}

func TestWriteBlock(t *testing.T) {
	r.client.FlushAll()
	r.WriteBlock("addy", 1000, 999, 0, "abcdef")

	sharesRes := r.client.HGetAllMap("test:shares:round0").Val()
	expectedRound := map[string]string{"addy": "1000"}
	if !reflect.DeepEqual(sharesRes, expectedRound) {
		t.Errorf("Invalid round data: %v", sharesRes)
	}

	blockRes := r.client.ZRevRangeWithScores("test:blocks:candidates", 0, 99999).Val()
	blockRes = stripTimestampsFromZs(blockRes)
	expectedCandidates := []redis.Z{redis.Z{0, "abcdef:*:999:1000"}}
	if !reflect.DeepEqual(blockRes, expectedCandidates) {
		t.Errorf("Invalid candidates data: %v, expected: %v", blockRes, expectedCandidates)
	}
}

func TestWriteBlockAtSameHeight(t *testing.T) {
	r.client.FlushAll()
	r.WriteBlock("addy", 1000, 999, 1, "00000000")
	r.WriteBlock("addy", 2000, 999, 1, "00000001")
	r.WriteBlock("addy", 3000, 999, 1, "00000002")

	sharesRes := r.client.HGetAllMap("test:shares:round1").Val()
	expectedRound := map[string]string{"addy": "3000"}
	if !reflect.DeepEqual(sharesRes, expectedRound) {
		t.Errorf("Invalid round data: %v", sharesRes)
	}

	blockRes := r.client.ZRevRangeWithScores("test:blocks:candidates", 0, 99999).Val()
	blockRes = stripTimestampsFromZs(blockRes)
	expectedBlocks := []redis.Z{redis.Z{1, "00000002:*:999:3000"}, redis.Z{1, "00000001:*:999:2000"}, redis.Z{1, "00000000:*:999:1000"}}
	if len(blockRes) != 3 {
		t.Errorf("Invalid number of candidates: %v, expected: %v", len(blockRes), 3)
	}
	if !reflect.DeepEqual(blockRes, expectedBlocks) {
		t.Errorf("Invalid candidates data: %v, expected: %v", blockRes, expectedBlocks)
	}
}

func stripTimestampFromZ(z redis.Z) redis.Z {
	k := strings.Split(z.Member.(string), ":")
	res := []string{k[0], "*", k[2], k[3]}
	return redis.Z{Score: z.Score, Member: strings.Join(res, ":")}
}

func stripTimestampsFromZs(zs []redis.Z) []redis.Z {
	var res []redis.Z
	for _, z := range zs {
		res = append(res, stripTimestampFromZ(z))
	}
	return res
}
