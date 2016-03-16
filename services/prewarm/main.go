package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/prewarm"
	"gopkg.in/redis.v3"
)

func withDefaultStr(config, default_ string) string {
	if config == "" {
		return default_
	}
	return config
}

func main() {
	redisPort := withDefaultStr(os.Getenv("GHVIZ_REDIS_PORT"), "6379")
	redisHost := os.Getenv("GHVIZ_REDIS_HOST")
	redisClient := interfaces.NewGoRedis(redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: os.Getenv("GHVIZ_REDIS_PASSWORD"),
		DB:       0,
	}))

	gh := github.NewClient(&github.Options{
		RedisClient: redisClient,
		Token:       os.Getenv("GITHUB_TOKEN"),
	})
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	owner := os.Getenv("GHVIZ_OWNER")
	repo := os.Getenv("GHVIZ_REPO")
	if err := prewarm.PrewarmHighScores(logger, gh, redisClient, owner, repo); err != nil {
		logger.Fatal(err)
	}
}
