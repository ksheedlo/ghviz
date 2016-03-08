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
)

var LINK_NEXT_REGEX *regexp.Regexp = regexp.MustCompile("<([^>]+)>; rel=\"next\"")

type Client struct {
	baseUrl  string
	client   *http.Client
	username string
	password string
}

func NewClientWithBaseUrl(username, password, baseUrl string) *Client {
	httpClient := &http.Client{}
	githubClient := &Client{}
	githubClient.client = httpClient
	githubClient.username = username
	githubClient.password = password
	githubClient.baseUrl = baseUrl
	return githubClient
}

func NewClient(username, password string) *Client {
	return NewClientWithBaseUrl(username, password, "https://api.github.com")
}

func (gh *Client) sendGithubRequest(logger *log.Logger, url, mediaType string) (*http.Response, *errors.HttpError) {
	rr, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Fatal(err)
		return nil, &errors.HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
	}
	rr.SetBasicAuth(gh.username, gh.password)
	rr.Header.Add("Accept", mediaType)
	startTime := time.Now()
	resp, err := gh.client.Do(rr)
	logger.Printf("send GET %s %s\n", url, time.Since(startTime).String())
	if err != nil {
		logger.Fatal(err)
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
			logger.Fatal(httpErr)
			return nil, httpErr
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Fatal(err)
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

func (gh *Client) ListStargazers(logger *log.Logger, owner, repo string) ([]map[string]interface{}, *errors.HttpError) {
	return gh.paginateGithub(
		logger,
		fmt.Sprintf("/repos/%s/%s/stargazers?per_page=100", owner, repo),
		"application/vnd.github.v3.star+json",
	)
}

func (gh *Client) ListIssues(logger *log.Logger, owner, repo string) ([]map[string]interface{}, *errors.HttpError) {
	return gh.paginateGithub(
		logger,
		fmt.Sprintf("/repos/%s/%s/issues?per_page=100&state=all&sort=created&direction=asc", owner, repo),
		"application/vnd.github.v3+json",
	)
}

func (gh *Client) ListTopIssues(logger *log.Logger, owner, repo string, limit int) ([]map[string]interface{}, *errors.HttpError) {
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
			logger.Fatal(err)
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

	return allItems, nil
}

func (gh *Client) ListTopPrs(logger *log.Logger, owner, repo string, limit int) ([]map[string]interface{}, *errors.HttpError) {
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
			logger.Fatal(err)
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

	return allItems, nil
}
