package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/ksheedlo/ghviz/errors"
)

var LINK_NEXT_REGEX *regexp.Regexp = regexp.MustCompile("<([^>]+)>; rel=\"next\"")

type GithubClient struct {
	baseUrl  string
	client   *http.Client
	username string
	password string
}

func NewClientWithBaseUrl(username, password, baseUrl string) *GithubClient {
	httpClient := &http.Client{}
	githubClient := &GithubClient{}
	githubClient.client = httpClient
	githubClient.username = username
	githubClient.password = password
	githubClient.baseUrl = baseUrl
	return githubClient
}

func NewClient(username, password string) *GithubClient {
	return NewClientWithBaseUrl(username, password, "https://api.github.com")
}

func (gh *GithubClient) sendGithubRequest(url, mediaType string) (*http.Response, *errors.HttpError) {
	rr, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		return nil, &errors.HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
	}
	rr.SetBasicAuth(gh.username, gh.password)
	rr.Header.Add("Accept", mediaType)
	log.Printf("GET %s\n", url)
	resp, err := gh.client.Do(rr)
	if err != nil {
		log.Fatal(err)
		return nil, &errors.HttpError{Message: "Github Upstream Error", Status: http.StatusBadGateway}
	}
	return resp, nil
}

func (gh *GithubClient) sendGithubV3Request(url string) (*http.Response, *errors.HttpError) {
	return gh.sendGithubRequest(url, "application/vnd.github.v3+json")
}

func (gh *GithubClient) paginateGithub(path, mediaType string) ([]map[string]interface{}, *errors.HttpError) {
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url := fmt.Sprintf("%s%s", gh.baseUrl, path); url != ""; {
		resp, httpErr := gh.sendGithubRequest(url, mediaType)
		if httpErr != nil {
			log.Fatal(httpErr)
			return nil, httpErr
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
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

func (gh *GithubClient) ListStargazers(owner, repo string) ([]map[string]interface{}, *errors.HttpError) {
	return gh.paginateGithub(
		fmt.Sprintf("/repos/%s/%s/stargazers?per_page=100", owner, repo),
		"application/vnd.github.v3.star+json",
	)
}

func (gh *GithubClient) ListIssues(owner, repo string) ([]map[string]interface{}, *errors.HttpError) {
	return gh.paginateGithub(
		fmt.Sprintf("/repos/%s/%s/issues?per_page=100&state=all&sort=created&direction=asc", owner, repo),
		"application/vnd.github.v3+json",
	)
}

func (gh *GithubClient) ListTopIssues(owner, repo string, limit int) ([]map[string]interface{}, *errors.HttpError) {
	url := fmt.Sprintf(
		"%s/repos/%s/%s/issues?per_page=100&state=open&sort=created&direction=desc",
		gh.baseUrl,
		owner,
		repo,
	)
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url != "" && len(allItems) < limit {
		resp, httpErr := gh.sendGithubV3Request(url)
		if httpErr != nil {
			return nil, httpErr
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
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

func (gh *GithubClient) ListTopPrs(owner, repo string, limit int) ([]map[string]interface{}, *errors.HttpError) {
	url := fmt.Sprintf(
		"%s/repos/%s/%s/issues?per_page=100&state=open&sort=created&direction=desc",
		gh.baseUrl,
		owner,
		repo,
	)
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url != "" && len(allItems) < limit {
		resp, httpErr := gh.sendGithubV3Request(url)
		if httpErr != nil {
			return nil, httpErr
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
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
