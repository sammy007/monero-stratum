package storage

import (
	"log"
	"strconv"
	"strings"
	"time"

	"gopkg.in/redis.v3"

	"../pool"
)

type RedisClient struct {
	client *redis.Client
	prefix string
}

func NewRedisClient(cfg *pool.Redis, prefix string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Endpoint,
		Password: cfg.Password,
		DB:       cfg.Database,
		PoolSize: cfg.PoolSize,
	})
	return &RedisClient{client: client, prefix: prefix}
}

func (r *RedisClient) Check() {
	pong, err := r.client.Ping().Result()
	if err != nil {
		log.Fatalf("Can't establish Redis connection: %v", err)
	}
	log.Printf("Redis PING command reply: %v", pong)
}

// Always returns list of addresses. If Redis fails it will return empty list.
func (r *RedisClient) GetBlacklist() []string {
	cmd := r.client.SMembers(r.formatKey("blacklist"))
	if cmd.Err() != nil {
		log.Printf("Failed to get blacklist from Redis: %v", cmd.Err())
		return []string{}
	}
	return cmd.Val()
}

// Always returns list of IPs. If Redis fails it will return empty list.
func (r *RedisClient) GetWhitelist() []string {
	cmd := r.client.SMembers(r.formatKey("whitelist"))
	if cmd.Err() != nil {
		log.Printf("Failed to get blacklist from Redis: %v", cmd.Err())
		return []string{}
	}
	return cmd.Val()
}

func (r *RedisClient) WriteShare(login string, diff int64) {
	tx := r.client.Multi()
	defer tx.Close()

	ms := time.Now().UnixNano() / 1000000
	ts := ms / 1000

	_, err := tx.Exec(func() error {
		r.writeShare(tx, ms, ts, login, diff)
		return nil
	})
	if err != nil {
		log.Printf("Failed to insert share data into Redis: %v", err)
	}
}

func (r *RedisClient) WriteBlock(login string, diff, roundDiff, height int64, hashHex string) {
	tx := r.client.Multi()
	defer tx.Close()

	ms := time.Now().UnixNano() / 1000000
	ts := ms / 1000

	cmds, err := tx.Exec(func() error {
		r.writeShare(tx, ms, ts, login, diff)
		tx.HSet(r.formatKey("stats"), "lastBlockFound", strconv.FormatInt(ms, 10))
		tx.ZIncrBy(r.formatKey("finders"), 1, login)
		tx.Rename(r.formatKey("shares", "roundCurrent"), r.formatKey("shares", formatRound(height)))
		tx.HGetAllMap(r.formatKey("shares", formatRound(height)))
		return nil
	})
	if err != nil {
		log.Printf("Failed to insert block candidate into Redis: %v", err)
	} else {
		sharesMap, _ := cmds[7].(*redis.StringStringMapCmd).Result()
		totalShares := int64(0)
		for _, v := range sharesMap {
			n, _ := strconv.ParseInt(v, 10, 64)
			totalShares += n
		}
		s := join(hashHex, ts, roundDiff, totalShares)
		cmd := r.client.ZAdd(r.formatKey("blocks", "candidates"), redis.Z{Score: float64(height), Member: s})
		if cmd.Err() != nil {
			log.Printf("Failed to insert block candidate shares into Redis: %v", cmd.Err())
		} else {
			log.Printf("Inserted block to Redis, height: %v, variance: %v/%v, %v", height, totalShares, roundDiff, cmd.Val())
		}
	}
}

func (r *RedisClient) writeShare(tx *redis.Multi, ms, ts int64, login string, diff int64) {
	tx.HIncrBy(r.formatKey("shares", "roundCurrent"), login, diff)
	tx.ZAdd(r.formatKey("hashrate"), redis.Z{Score: float64(ts), Member: join(diff, login, ms)})
	tx.HIncrBy(r.formatKey("workers", login), "hashes", diff)
	tx.HSet(r.formatKey("workers", login), "lastShare", strconv.FormatInt(ts, 10))
}

func (r *RedisClient) formatKey(args ...interface{}) string {
	return join(r.prefix, join(args...))
}

func formatRound(height int64) string {
	return "round" + strconv.FormatInt(height, 10)
}

func join(args ...interface{}) string {
	s := make([]string, len(args))
	for i, v := range args {
		switch v.(type) {
		case string:
			s[i] = v.(string)
		case int64:
			s[i] = strconv.FormatInt(v.(int64), 10)
		default:
			panic("Invalid type specified for conversion")
		}
	}
	return strings.Join(s, ":")
}
