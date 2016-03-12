package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"gopkg.in/redis.v3"

	"github.com/ksheedlo/ghviz/errors"
)

var LINK_NEXT_REGEX *regexp.Regexp = regexp.MustCompile("<([^>]+)>; rel=\"next\"")

type Client struct {
	baseUrl     string
	httpClient  *http.Client
	redisClient *redis.Client
	token       string
}

type Options struct {
	BaseUrl     string
	RedisClient *redis.Client
	Token       string
}

type Issue struct {
	ClosedAt  time.Time
	CreatedAt time.Time
	EventsUrl string
	IsClosed  bool
	IsPr      bool
}

func withDefaultBaseUrl(baseUrl string) string {
	if baseUrl == "" {
		return "https://api.github.com"
	}
	return baseUrl
}

func NewClient(options *Options) *Client {
	httpClient := &http.Client{}
	client := &Client{}
	client.httpClient = httpClient
	client.baseUrl = withDefaultBaseUrl(options.BaseUrl)
	client.redisClient = options.RedisClient
	client.token = options.Token
	return client
}

func (gh *Client) sendGithubRequest(logger *log.Logger, url, mediaType string) (*http.Response, *errors.HttpError) {
	rr, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Printf("%s\n", err.Error())
		return nil, &errors.HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
	}
	rr.Header.Add("Authorization", fmt.Sprintf("token %s", gh.token))
	rr.Header.Add("Accept", mediaType)
	startTime := time.Now()
	resp, err := gh.httpClient.Do(rr)
	logger.Printf("send GET %s %s\n", url, time.Since(startTime).String())
	if err != nil {
		logger.Printf("ERROR: %s\n", err.Error())
		return nil, &errors.HttpError{Message: "Github Upstream Error", Status: http.StatusBadGateway}
	}
	return resp, nil
}

func (gh *Client) sendGithubV3Request(logger *log.Logger, url string) (*http.Response, *errors.HttpError) {
	return gh.sendGithubRequest(logger, url, "application/vnd.github.v3+json")
}

func (gh *Client) paginateGithub(logger *log.Logger, path, mediaType string) ([]map[string]interface{}, *errors.HttpError) {
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url := fmt.Sprintf("%s%s", gh.baseUrl, path); url != ""; {
		resp, httpErr := gh.sendGithubRequest(logger, url, mediaType)
		if httpErr != nil {
			return nil, httpErr
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, &errors.HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
		}
		json.Unmarshal(contents, &items)
		allItems = append(allItems, items...)
		for i := 0; i < len(items); i++ {
			items[i] = nil
		}
		match := LINK_NEXT_REGEX.FindStringSubmatch(resp.Header.Get("Link"))
		if match != nil {
			url = match[1]
		} else {
			url = ""
		}
	}

	return allItems, nil
}

func stargazersKey(owner, repo string) string {
	return fmt.Sprintf("github:repo:%s:%s:stargazers", owner, repo)
}

func (gh *Client) redisWrap(
	cacheKey string,
	pluralType string,
	logger *log.Logger,
	fallback func() ([]map[string]interface{}, *errors.HttpError),
) ([]map[string]interface{}, *errors.HttpError) {
	if gh.redisClient != nil {
		cachedItems, err := gh.redisClient.Get(cacheKey).Result()
		if err != nil || cachedItems == "" {
			logger.Printf(
				"Key %s was not found in Redis, attempting to fetch %s from Github.\n",
				cacheKey,
				pluralType,
			)
		} else {
			var items []map[string]interface{}
			err := json.Unmarshal([]byte(cachedItems), &items)
			if err != nil {
				logger.Printf(
					"Key %s was found in Redis, but a JSON decoding error occurred: %s\n",
					cacheKey,
					err.Error(),
				)
			} else {
				logger.Printf("Found %s in Redis.\n", cacheKey)
				return items, nil
			}
		}
	}

	items, err := fallback()
	if err != nil {
		return nil, err
	}

	if gh.redisClient == nil {
		return items, nil
	}
	jsonBlob, jsonErr := json.Marshal(items)
	if jsonErr != nil {
		logger.Printf("JSON encoding error occurred: %s\n", jsonErr.Error())
		return items, nil
	}
	duration, durationErr := time.ParseDuration("10m")
	if durationErr != nil {
		logger.Printf("Duration parsing error occurred: %s\n", durationErr.Error())
		return items, nil
	}
	if redisErr := gh.redisClient.Set(cacheKey, jsonBlob, duration).Err(); redisErr != nil {
		logger.Printf("Redis store error occurred: %s\n", redisErr.Error())
	}
	return items, nil
}

func (gh *Client) ListStargazers(logger *log.Logger, owner, repo string) ([]map[string]interface{}, *errors.HttpError) {
	return gh.redisWrap(
		stargazersKey(owner, repo),
		"stargazers",
		logger,
		func() ([]map[string]interface{}, *errors.HttpError) {
			stargazers, err := gh.paginateGithub(
				logger,
				fmt.Sprintf("/repos/%s/%s/stargazers?per_page=100", owner, repo),
				"application/vnd.github.v3.star+json",
			)
			if err != nil {
				return nil, err
			}
			for _, stargazer := range stargazers {
				for key, _ := range stargazer {
					if key != "starred_at" {
						delete(stargazer, key)
					}
				}
			}
			return stargazers, nil
		},
	)
}

func (gh *Client) ListIssues(logger *log.Logger, owner, repo string) ([]Issue, *errors.HttpError) {
	rawIssues, err := gh.redisWrap(
		fmt.Sprintf("github:repo:%s:%s:issues", owner, repo),
		"issues",
		logger,
		func() ([]map[string]interface{}, *errors.HttpError) {
			issues, err := gh.paginateGithub(
				logger,
				fmt.Sprintf("/repos/%s/%s/issues?per_page=100&state=all&sort=created&direction=asc", owner, repo),
				"application/vnd.github.v3+json",
			)
			if err != nil {
				return nil, err
			}
			for _, issue := range issues {
				for key, _ := range issue {
					if key != "closed_at" &&
						key != "created_at" &&
						key != "pull_request" &&
						key != "events_url" {
						delete(issue, key)
					}
				}
			}
			return issues, nil
		},
	)

	if err != nil {
		return nil, err
	}

	issues := make([]Issue, len(rawIssues))
	for i, rawIssue := range rawIssues {
		createdAt, parseErr := time.Parse(time.RFC3339, rawIssue["created_at"].(string))
		if parseErr != nil {
			logger.Printf("ERROR: %s\n", parseErr.Error())
			return nil, &errors.HttpError{
				Message: "Server Error", Status: http.StatusInternalServerError,
			}
		}
		issues[i].IsClosed = false
		var rawClosedAt interface{}
		var hasClosedAt bool
		if rawClosedAt, hasClosedAt = rawIssue["closed_at"]; hasClosedAt {
			issues[i].IsClosed = (rawClosedAt != nil)
		}
		if issues[i].IsClosed {
			closedAt, parseErr := time.Parse(time.RFC3339, rawClosedAt.(string))
			if parseErr != nil {
				logger.Printf("ERROR: %s\n", parseErr.Error())
				return nil, &errors.HttpError{
					Message: "Server Error",
					Status:  http.StatusInternalServerError,
				}
			}
			issues[i].ClosedAt = closedAt
		}
		_, isPr := rawIssue["pull_request"]

		issues[i].CreatedAt = createdAt
		issues[i].EventsUrl = rawIssue["events_url"].(string)
		issues[i].IsPr = isPr
	}

	return issues, nil
}

func (gh *Client) ListTopIssues(logger *log.Logger, owner, repo string, limit int) ([]map[string]interface{}, *errors.HttpError) {
	return gh.redisWrap(
		fmt.Sprintf("github:repo:%s:%s:top_issues:%d", owner, repo, limit),
		"top issues",
		logger,
		func() ([]map[string]interface{}, *errors.HttpError) {
			url := fmt.Sprintf(
				"%s/repos/%s/%s/issues?per_page=100&state=open&sort=created&direction=desc",
				gh.baseUrl,
				owner,
				repo,
			)
			items := make([]map[string]interface{}, 0)
			allItems := make([]map[string]interface{}, 0)

			for url != "" && len(allItems) < limit {
				resp, httpErr := gh.sendGithubV3Request(logger, url)
				if httpErr != nil {
					return nil, httpErr
				}
				defer resp.Body.Close()
				contents, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logger.Printf("ERROR: %s\n", err.Error())
					return nil, &errors.HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
				}
				json.Unmarshal(contents, &items)
				for i := 0; i < len(items) && len(allItems) < limit; i++ {
					if _, isPr := items[i]["pull_request"]; !isPr {
						allItems = append(allItems, items[i])
					}
				}
				for i := 0; i < len(items); i++ {
					items[i] = nil
				}
				match := LINK_NEXT_REGEX.FindStringSubmatch(resp.Header.Get("Link"))
				if match != nil {
					url = match[1]
				} else {
					url = ""
				}
			}

			for _, item := range allItems {
				for key, _ := range item {
					if key != "html_url" && key != "title" {
						delete(item, key)
					}
				}
			}

			return allItems, nil
		},
	)
}

func (gh *Client) ListTopPrs(logger *log.Logger, owner, repo string, limit int) ([]map[string]interface{}, *errors.HttpError) {
	return gh.redisWrap(
		fmt.Sprintf("github:repo:%s:%s:top_prs:%d", owner, repo, limit),
		"top PRs",
		logger,
		func() ([]map[string]interface{}, *errors.HttpError) {
			url := fmt.Sprintf(
				"%s/repos/%s/%s/issues?per_page=100&state=open&sort=created&direction=desc",
				gh.baseUrl,
				owner,
				repo,
			)
			items := make([]map[string]interface{}, 0)
			allItems := make([]map[string]interface{}, 0)

			for url != "" && len(allItems) < limit {
				resp, httpErr := gh.sendGithubV3Request(logger, url)
				if httpErr != nil {
					return nil, httpErr
				}
				defer resp.Body.Close()
				contents, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					logger.Printf("ERROR: %s\n", err.Error())
					return nil, &errors.HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
				}
				json.Unmarshal(contents, &items)
				for i := 0; i < len(items) && len(allItems) < limit; i++ {
					if _, isPr := items[i]["pull_request"]; isPr {
						allItems = append(allItems, items[i])
					}
				}
				for i := 0; i < len(items); i++ {
					items[i] = nil
				}
				match := LINK_NEXT_REGEX.FindStringSubmatch(resp.Header.Get("Link"))
				if match != nil {
					url = match[1]
				} else {
					url = ""
				}
			}

			for _, item := range allItems {
				for key, _ := range item {
					if key != "html_url" && key != "title" {
						delete(item, key)
					}
				}
			}

			return allItems, nil
		},
	)
}
