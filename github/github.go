package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ksheedlo/ghviz/errors"
	"github.com/ksheedlo/ghviz/interfaces"
)

var LINK_NEXT_REGEX *regexp.Regexp = regexp.MustCompile("<([^>]+)>; rel=\"next\"")

type Client struct {
	baseUrl      string
	httpClient   *http.Client
	maxStaleness int
	redisClient  interfaces.Rediser
	token        string
}

type Options struct {
	BaseUrl      string
	MaxStaleness int
	RedisClient  interfaces.Rediser
	Token        string
}

type Issue struct {
	ClosedAt  time.Time
	CreatedAt time.Time
	EventsUrl string
	HtmlUrl   string
	IsClosed  bool
	IsPr      bool
	Number    int
	Submitter string
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
		"number":     issue.Number,
		"submitter":  issue.Submitter,
		"title":      issue.Title,
	})
}

type DetailedIssueEventType int

const (
	IssueCreated DetailedIssueEventType = iota
	IssueClosed
	IssueMerged
	IssueLabeled
	IssueUnlabeled
)

var issueEventTypes map[string]DetailedIssueEventType = map[string]DetailedIssueEventType{
	"closed":    IssueClosed,
	"created":   IssueCreated,
	"merged":    IssueMerged,
	"labeled":   IssueLabeled,
	"unlabeled": IssueUnlabeled,
}

type DetailedIssueEvent struct {
	ActorId     string
	CreatedAt   time.Time
	Detail      interface{}
	EventType   DetailedIssueEventType
	Id          string
	IssueNumber int
}

type ByCreatedAt []DetailedIssueEvent

func (a ByCreatedAt) Len() int           { return len(a) }
func (a ByCreatedAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreatedAt) Less(i, j int) bool { return a[i].CreatedAt.Before(a[j].CreatedAt) }

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
	client.maxStaleness = options.MaxStaleness
	client.redisClient = options.RedisClient
	client.token = options.Token
	return client
}

func (gh *Client) sendGithubRequest(logger *log.Logger, url, mediaType string) (*http.Response, *errors.HttpError) {
	// Suppress errors from http.NewRequest. This can only error if the URL or
	// HTTP method are invalid, and we control both in this module.
	rr, _ := http.NewRequest("GET", url, nil)
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
	urll, mediaType string,
) ([]map[string]interface{}, *errors.HttpError) {
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url := urll; url != ""; {
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

func parseRedisValues(cacheKey, cachedValues string) (time.Time, []byte, error) {
	idx := strings.Index(cachedValues, "|")
	if idx == -1 {
		return time.Unix(0, 0), nil, &errors.BadRedisValues{CacheKey: cacheKey}
	}
	unixSeconds, err := strconv.ParseInt(cachedValues[:idx], 10, 64)
	if err != nil {
		return time.Unix(0, 0), nil, err
	}
	jsonBytes := []byte(cachedValues[(idx + 1):])
	return time.Unix(unixSeconds, 0), jsonBytes, nil
}

func isStale(gh *Client, timeSubmitted time.Time) bool {
	return time.Since(timeSubmitted) > time.Duration(gh.maxStaleness)*time.Minute
}

func redisWrap(
	gh *Client,
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
			timeSubmitted, jsonBytes, err := parseRedisValues(cacheKey, cachedItems)
			if err != nil {
				logger.Printf(
					"Failed to parse Redis values because of an error: %s, attempting to fetch from Github.\n",
					err.Error(),
				)
			} else if isStale(gh, timeSubmitted) {
				logger.Printf(
					"Key %s was found stale, attempting to fetch from Github.\n",
					cacheKey,
				)
			} else {
				var items []map[string]interface{}
				err := json.Unmarshal(jsonBytes, &items)
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
	if redisErr := gh.redisClient.Set(
		cacheKey,
		fmt.Sprintf("%d|%s", time.Now().Unix(), string(jsonBlob)),
		time.Duration(0),
	); redisErr != nil {
		logger.Printf("Redis store error occurred: %s\n", redisErr.Error())
	}
	return items, nil
}

func (gh *Client) ListStargazers(logger *log.Logger, owner, repo string) ([]map[string]interface{}, *errors.HttpError) {
	return redisWrap(
		gh,
		stargazersKey(owner, repo),
		"stargazers",
		logger,
		func() ([]map[string]interface{}, *errors.HttpError) {
			stargazers, err := gh.paginateGithub(
				logger,
				fmt.Sprintf("%s/repos/%s/%s/stargazers?per_page=100", gh.baseUrl, owner, repo),
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
				key != "number" &&
				key != "pull_request" &&
				key != "title" &&
				key != "user" {
				delete(issue, key)
			}
		}

		userJson := issue["user"].(map[string]interface{})
		for key, _ := range userJson {
			if key != "login" {
				delete(userJson, key)
			}
		}
	}
}

func parseIssue(
	logger *log.Logger,
	issue *Issue,
	rawIssue map[string]interface{},
) *errors.HttpError {
	createdAt, parseErr := time.Parse(time.RFC3339, rawIssue["created_at"].(string))
	if parseErr != nil {
		logger.Printf("ERROR: %s\n", parseErr.Error())
		return &errors.HttpError{
			Message: "Server Error", Status: http.StatusInternalServerError,
		}
	}
	issue.IsClosed = false
	var rawClosedAt interface{}
	var hasClosedAt bool
	if rawClosedAt, hasClosedAt = rawIssue["closed_at"]; hasClosedAt {
		issue.IsClosed = (rawClosedAt != nil)
	}
	if issue.IsClosed {
		closedAt, parseErr := time.Parse(time.RFC3339, rawClosedAt.(string))
		if parseErr != nil {
			logger.Printf("ERROR: %s\n", parseErr.Error())
			return &errors.HttpError{
				Message: "Server Error",
				Status:  http.StatusInternalServerError,
			}
		}
		issue.ClosedAt = closedAt
	}
	_, isPr := rawIssue["pull_request"]

	issue.CreatedAt = createdAt
	issue.EventsUrl = rawIssue["events_url"].(string)
	issue.HtmlUrl = rawIssue["html_url"].(string)
	issue.IsPr = isPr
	issue.Number = int(rawIssue["number"].(float64))
	issue.Submitter = (rawIssue["user"].(map[string]interface{}))["login"].(string)
	issue.Title = rawIssue["title"].(string)
	return nil
}

func parseIssues(logger *log.Logger, rawIssues []map[string]interface{}) ([]Issue, *errors.HttpError) {
	issues := make([]Issue, len(rawIssues))
	for i, rawIssue := range rawIssues {
		if err := parseIssue(logger, &issues[i], rawIssue); err != nil {
			return nil, err
		}
	}

	return issues, nil
}

func (gh *Client) ListIssues(logger *log.Logger, owner, repo string) ([]Issue, *errors.HttpError) {
	rawIssues, err := redisWrap(
		gh,
		fmt.Sprintf("github:repo:%s:%s:issues", owner, repo),
		"issues",
		logger,
		func() ([]map[string]interface{}, *errors.HttpError) {
			issues, err := gh.paginateGithub(
				logger,
				fmt.Sprintf(
					"%s/repos/%s/%s/issues?per_page=100&state=all&sort=created&direction=asc",
					gh.baseUrl,
					owner,
					repo,
				),
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

type ListAllPrEventser interface {
	ListAllPrEvents(*log.Logger, string, string) ([]DetailedIssueEvent, *errors.HttpError)
}

func (gh *Client) ListAllPrEvents(
	logger *log.Logger,
	owner, repo string,
) ([]DetailedIssueEvent, *errors.HttpError) {
	issueEvents, err := gh.paginateGithub(
		logger,
		fmt.Sprintf("%s/repos/%s/%s/issues/events?per_page=100", gh.baseUrl, owner, repo),
		"application/vnd.github.v3+json",
	)
	if err != nil {
		return nil, err
	}

	var detailedEvents []DetailedIssueEvent
	knownIssues := make(map[int]Issue)
	for _, event := range issueEvents {
		issueNumber := int((event["issue"].(map[string]interface{}))["number"].(float64))
		issue, issueIsKnown := knownIssues[issueNumber]
		if !issueIsKnown {
			issue = Issue{}
			parseIssue(logger, &issue, event["issue"].(map[string]interface{}))
			knownIssues[issue.Number] = issue
			if issue.IsPr {
				detailedEvents = append(detailedEvents, DetailedIssueEvent{
					ActorId:     issue.Submitter,
					CreatedAt:   issue.CreatedAt,
					EventType:   IssueCreated,
					Id:          fmt.Sprintf("cr%d", issue.Number),
					IssueNumber: issue.Number,
				})
			}
		}
		if eventType, eventIsKnown := issueEventTypes[event["event"].(string)]; eventIsKnown {
			actorId := (event["actor"].(map[string]interface{}))["login"].(string)
			if issue.IsPr {
				var detail interface{}
				switch eventType {
				case IssueClosed, IssueMerged:
					detail = event["commit_id"]
				case IssueLabeled, IssueUnlabeled:
					detail = event["label"]
				}
				createdAt, err := time.Parse(time.RFC3339, event["created_at"].(string))
				if err != nil {
					logger.Printf("ERROR: %s\n", err)
					return nil, &errors.HttpError{
						Message: "Server Error",
						Status:  http.StatusInternalServerError,
					}
				}
				detailedEvents = append(detailedEvents, DetailedIssueEvent{
					ActorId:     actorId,
					CreatedAt:   createdAt,
					Detail:      detail,
					EventType:   eventType,
					Id:          fmt.Sprintf("%d", int(event["id"].(float64))),
					IssueNumber: issue.Number,
				})
			}
		}
	}

	return detailedEvents, nil
}

func (gh *Client) filterTopIssues(
	logger *log.Logger,
	cacheKey, pluralType, owner, repo string,
	limit int,
	filterFn func(map[string]interface{}) bool,
) ([]Issue, *errors.HttpError) {
	rawIssues, err := redisWrap(
		gh,
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
