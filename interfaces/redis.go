package interfaces

import (
	"time"

	"github.com/stretchr/testify/mock"
	"gopkg.in/redis.v3"
)

type Rediser interface {
	Get(string) (string, error)
	Set(string, string, time.Duration) error
}

type GoRedisAdapter struct {
	redisClient *redis.Client
}

func NewGoRedis(redisClient *redis.Client) Rediser {
	gr := new(GoRedisAdapter)
	gr.redisClient = redisClient
	return gr
}

func (gr *GoRedisAdapter) Get(key string) (string, error) {
	return gr.redisClient.Get(key).Result()
}

func (gr *GoRedisAdapter) Set(key, value string, ttl time.Duration) error {
	return gr.redisClient.Set(key, value, ttl).Err()
}

type MockRediser struct {
	mock.Mock
}

func (m *MockRediser) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockRediser) Set(key, value string, ttl time.Duration) error {
	// Set the value to "" because it's used to set JSON, which serializes
	// in an unpredictable order.
	args := m.Called(key, "", ttl)
	return args.Error(0)
}
