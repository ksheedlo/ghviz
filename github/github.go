package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/ksheedlo/ghviz/errors"
	"github.com/ksheedlo/ghviz/interfaces"
)

var LINK_NEXT_REGEX *regexp.Regexp = regexp.MustCompile("<([^>]+)>; rel=\"next\"")

type Client struct {
	baseUrl     string
	httpClient  *http.Client
	redisClient interfaces.Rediser
	token       string
}

type Options struct {
	BaseUrl     string
	RedisClient interfaces.Rediser
	Token       string
}

type Issue struct {
	ClosedAt  time.Time
	CreatedAt time.Time
	EventsUrl string
	HtmlUrl   string
	IsClosed  bool
	IsPr      bool
	Title     string
}

func (issue *Issue) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"closed_at":  issue.ClosedAt,
		"created_at": issue.CreatedAt,
		"events_url": issue.EventsUrl,
		"html_url":   issue.HtmlUrl,
		"is_closed":  issue.IsClosed,
		"is_pr":      issue.IsPr,
		"title":      issue.Title,
	})
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

func (gh *Client) paginateGithub(
	logger *log.Logger,
	path, mediaType string,
) ([]map[string]interface{}, *errors.HttpError) {
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
		cachedItems, err := gh.redisClient.Get(cacheKey)
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
	duration := (time.Duration(10) * time.Minute)
	if redisErr := gh.redisClient.Set(cacheKey, string(jsonBlob), duration); redisErr != nil {
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

func cleanIssueJsons(issues []map[string]interface{}) {
	for _, issue := range issues {
		for key, _ := range issue {
			if key != "closed_at" &&
				key != "created_at" &&
				key != "events_url" &&
				key != "html_url" &&
				key != "pull_request" &&
				key != "title" {
				delete(issue, key)
			}
		}
	}
}

func parseIssues(logger *log.Logger, rawIssues []map[string]interface{}) ([]Issue, *errors.HttpError) {
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
		issues[i].HtmlUrl = rawIssue["html_url"].(string)
		issues[i].IsPr = isPr
		issues[i].Title = rawIssue["title"].(string)
	}

	return issues, nil
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
			cleanIssueJsons(issues)
			return issues, nil
		},
	)

	if err != nil {
		return nil, err
	}

	return parseIssues(logger, rawIssues)
}

func (gh *Client) filterTopIssues(
	logger *log.Logger,
	cacheKey, pluralType, owner, repo string,
	limit int,
	filterFn func(map[string]interface{}) bool,
) ([]Issue, *errors.HttpError) {
	rawIssues, err := gh.redisWrap(
		cacheKey,
		pluralType,
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
					if filterFn(items[i]) {
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

			cleanIssueJsons(allItems)
			return allItems, nil
		},
	)

	if err != nil {
		return nil, err
	}
	return parseIssues(logger, rawIssues)
}

func (gh *Client) ListTopIssues(logger *log.Logger, owner, repo string, limit int) ([]Issue, *errors.HttpError) {
	return gh.filterTopIssues(
		logger,
		fmt.Sprintf("github:repo:%s:%s:top_issues:%d", owner, repo, limit),
		"top issues",
		owner,
		repo,
		limit,
		func(rawIssue map[string]interface{}) bool {
			_, isPr := rawIssue["pull_request"]
			return !isPr
		},
	)
}

func (gh *Client) ListTopPrs(logger *log.Logger, owner, repo string, limit int) ([]Issue, *errors.HttpError) {
	return gh.filterTopIssues(
		logger,
		fmt.Sprintf("github:repo:%s:%s:top_prs:%d", owner, repo, limit),
		"top PRs",
		owner,
		repo,
		limit,
		func(rawIssue map[string]interface{}) bool {
			_, isPr := rawIssue["pull_request"]
			return isPr
		},
	)
}
