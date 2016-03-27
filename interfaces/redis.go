package interfaces

import (
	"time"

	"gopkg.in/redis.v3"
)

type ZZ struct {
	Score  float64
	Member interface{}
}

type ZRangeByScoreOpts struct {
	Min, Max string
}

type Rediser interface {
	Del(string) (int64, error)
	Get(string) (string, error)
	Set(string, string, time.Duration) error
	ZAdd(string, ...ZZ) (int64, error)
	ZRangeByScore(string, *ZRangeByScoreOpts) ([]string, error)
}

type GoRedisAdapter struct {
	redisClient *redis.Client
}

func NewGoRedis(redisClient *redis.Client) Rediser {
	gr := new(GoRedisAdapter)
	gr.redisClient = redisClient
	return gr
}

func (gr *GoRedisAdapter) Del(key string) (int64, error) {
	return gr.redisClient.Del(key).Result()
}

func (gr *GoRedisAdapter) Get(key string) (string, error) {
	return gr.redisClient.Get(key).Result()
}

func (gr *GoRedisAdapter) Set(key, value string, ttl time.Duration) error {
	return gr.redisClient.Set(key, value, ttl).Err()
}

func (gr *GoRedisAdapter) ZAdd(key string, members ...ZZ) (int64, error) {

	var zmembers []redis.Z
	for _, member := range members {
		zmembers = append(zmembers, redis.Z{
			Score:  member.Score,
			Member: member.Member,
		})
	}
	return gr.redisClient.ZAdd(key, zmembers...).Result()
}

func (gr *GoRedisAdapter) ZRangeByScore(
	key string,
	opts *ZRangeByScoreOpts,
) ([]string, error) {
	return gr.redisClient.ZRangeByScore(key, redis.ZRangeByScore{
		Min: opts.Min,
		Max: opts.Max,
	}).Result()
}
