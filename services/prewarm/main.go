package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/prewarm"

	"github.com/jonboulle/clockwork"
	"gopkg.in/redis.v3"
)

func withDefaultStr(config, default_ string) string {
	if config == "" {
		return default_
	}
	return config
}

const HIGH_SCORES_USAGE string = `Prewarm the list of high scores (i.e., all time monthly top contributors)
        into the cache. Recommended, as this is an expensive operation.`

const ISSUES_USAGE string = `Prewarm Github issues into the cache.`

const STARGAZERS_USAGE string = `Prewarm stargazers into the cache.`

const TOP_ISSUES_USAGE string = `Specify the number of top issues to prewarm into the cache. Set to 0 or
        leave empty to not fetch top issues.`

const TOP_PRS_USAGE string = `Specify the number of top PRs to prewarm into the cache. Set to 0 or leave
        empty to not fetch top issues.`

func main() {
	redisPort := withDefaultStr(os.Getenv("GHVIZ_REDIS_PORT"), "6379")
	redisHost := os.Getenv("GHVIZ_REDIS_HOST")
	redisClient := interfaces.NewGoRedis(redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: os.Getenv("GHVIZ_REDIS_PASSWORD"),
		DB:       0,
	}))

	gh := github.NewClient(&github.Options{
		MaxStaleness: -1,
		RedisClient:  redisClient,
		Token:        os.Getenv("GITHUB_TOKEN"),
	})
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)
	owner := os.Getenv("GHVIZ_OWNER")
	repo := os.Getenv("GHVIZ_REPO")

	prewarmHighScores := flag.Bool("high-scores", false, HIGH_SCORES_USAGE)
	prewarmIssues := flag.Bool("issues", false, ISSUES_USAGE)
	prewarmStargazers := flag.Bool("stargazers", false, STARGAZERS_USAGE)
	prewarmTopIssues := flag.Int("top-issues", 0, TOP_ISSUES_USAGE)
	prewarmTopPrs := flag.Int("top-prs", 0, TOP_PRS_USAGE)

	flag.Parse()

	errChan := make(chan int)
	pendingTasks := 0
	if *prewarmHighScores {
		pendingTasks++
		go func() {
			if err := prewarm.PrewarmHighScores(
				logger,
				gh,
				redisClient,
				clockwork.NewRealClock(),
				interfaces.RandomTag,
				owner,
				repo,
			); err != nil {
				logger.Printf("ERROR: %s\n", err.Error())
				errChan <- 1
			} else {
				errChan <- 0
			}
		}()
	}
	if *prewarmIssues {
		pendingTasks++
		go func() {
			if _, err := gh.ListIssues(logger, owner, repo); err != nil {
				logger.Printf("ERROR: %s\n", err.Error())
				errChan <- 1
			} else {
				errChan <- 0
			}
		}()
	}
	if *prewarmStargazers {
		pendingTasks++
		go func() {
			if _, err := gh.ListStargazers(logger, owner, repo); err != nil {
				logger.Printf("ERROR: %s\n", err.Error())
				errChan <- 1
			} else {
				errChan <- 0
			}
		}()
	}
	if *prewarmTopIssues > 0 {
		pendingTasks++
		go func() {
			if _, err := gh.ListTopIssues(logger, owner, repo, *prewarmTopIssues); err != nil {
				logger.Printf("ERROR: %s\n", err.Error())
				errChan <- 1
			} else {
				errChan <- 0
			}
		}()
	}
	if *prewarmTopPrs > 0 {
		pendingTasks++
		go func() {
			if _, err := gh.ListTopPrs(logger, owner, repo, *prewarmTopPrs); err != nil {
				logger.Printf("ERROR: %s\n", err.Error())
				errChan <- 1
			} else {
				errChan <- 0
			}
		}()
	}

	if pendingTasks == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(2)
	}
	exitCode := 0
	for i := 0; i < pendingTasks; i++ {
		code := <-errChan
		if code == 1 {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}
