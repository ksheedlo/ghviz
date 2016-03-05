package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"sort"
	"time"

	"github.com/gorilla/mux"
)

var LINK_NEXT_REGEX *regexp.Regexp = regexp.MustCompile("<([^>]+)>; rel=\"next\"")

type HttpError struct {
	Message string
	Status  int
}

func (e HttpError) Error() string {
	return e.Message
}

type StarEvent struct {
	StarredAt time.Time
}

type StarCount struct {
	Stars     int
	Timestamp time.Time
	UnixTime  int64
}

type ByStarredAt []StarEvent

func (a ByStarredAt) Len() int           { return len(a) }
func (a ByStarredAt) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByStarredAt) Less(i, j int) bool { return a[i].StarredAt.Before(a[j].StarredAt) }

type IssueEventType int

const (
	IssueOpened IssueEventType = iota
	IssueClosed
)

type IssueEvent struct {
	EventType IssueEventType
	Timestamp time.Time
}

type ByTimestamp []IssueEvent

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }

type OpenIssueCount struct {
	OpenIssues int
	Timestamp  time.Time
	UnixTime   int64
}

func DecodeStarEvents(apiObjects []map[string]interface{}) ([]StarEvent, error) {
	starEvents := make([]StarEvent, len(apiObjects))
	for i := 0; i < len(apiObjects); i++ {
		starredAt, err := time.Parse(time.RFC3339, apiObjects[i]["starred_at"].(string))
		if err != nil {
			return nil, err
		}
		starEvents[i].StarredAt = starredAt
	}
	sort.Sort(ByStarredAt(starEvents))
	return starEvents, nil
}

func ComputeStarCounts(starEvents []StarEvent) []StarCount {
	starCounts := make([]StarCount, len(starEvents))
	for i := 0; i < len(starEvents); i++ {
		starCounts[i].Stars = i + 1
		starCounts[i].Timestamp = starEvents[i].StarredAt
		starCounts[i].UnixTime = starEvents[i].StarredAt.Unix()
	}
	return starCounts
}

func DecodeIssueEvents(apiObjects []map[string]interface{}) ([]IssueEvent, error) {
	var issueEvents []IssueEvent
	for i := 0; i < len(apiObjects); i++ {
		if _, isPullRequest := apiObjects[i]["pull_request"]; !isPullRequest {
			issueOpened := IssueEvent{EventType: IssueOpened}
			createdAt, err := time.Parse(time.RFC3339, apiObjects[i]["created_at"].(string))
			if err != nil {
				return nil, err
			}
			issueOpened.Timestamp = createdAt
			issueEvents = append(issueEvents, issueOpened)

			if closedAt := apiObjects[i]["closed_at"]; closedAt != nil {
				issueClosed := IssueEvent{EventType: IssueClosed}
				closedAt, err := time.Parse(time.RFC3339, closedAt.(string))
				if err != nil {
					return nil, err
				}
				issueClosed.Timestamp = closedAt
				issueEvents = append(issueEvents, issueClosed)
			}
		}
	}
	sort.Sort(ByTimestamp(issueEvents))
	return issueEvents, nil
}

func DecodePrEvents(apiObjects []map[string]interface{}) ([]IssueEvent, error) {
	var issueEvents []IssueEvent
	for i := 0; i < len(apiObjects); i++ {
		if _, isPullRequest := apiObjects[i]["pull_request"]; isPullRequest {
			issueOpened := IssueEvent{EventType: IssueOpened}
			createdAt, err := time.Parse(time.RFC3339, apiObjects[i]["created_at"].(string))
			if err != nil {
				return nil, err
			}
			issueOpened.Timestamp = createdAt
			issueEvents = append(issueEvents, issueOpened)

			if closedAt := apiObjects[i]["closed_at"]; closedAt != nil {
				issueClosed := IssueEvent{EventType: IssueClosed}
				closedAt, err := time.Parse(time.RFC3339, closedAt.(string))
				if err != nil {
					return nil, err
				}
				issueClosed.Timestamp = closedAt
				issueEvents = append(issueEvents, issueClosed)
			}
		}
	}
	sort.Sort(ByTimestamp(issueEvents))
	return issueEvents, nil
}

func ComputeOpenIssueCounts(issueEvents []IssueEvent) []OpenIssueCount {
	issueCounts := make([]OpenIssueCount, len(issueEvents))
	openIssues := 0
	for i := 0; i < len(issueEvents); i++ {
		if issueEvents[i].EventType == IssueOpened {
			openIssues++
		} else {
			openIssues--
		}
		issueCounts[i].OpenIssues = openIssues
		issueCounts[i].Timestamp = issueEvents[i].Timestamp
		issueCounts[i].UnixTime = issueEvents[i].Timestamp.Unix()
	}
	return issueCounts
}

func PaginateGithub(path, mediaType string) ([]map[string]interface{}, *HttpError) {
	url := fmt.Sprintf("https://api.github.com%s", path)
	client := &http.Client{}
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url != "" {
		rr, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
			return nil, &HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
		}
		rr.SetBasicAuth(os.Getenv("GITHUB_USERNAME"), os.Getenv("GITHUB_PASSWORD"))
		rr.Header.Add("Accept", mediaType)
		resp, err := client.Do(rr)
		if err != nil {
			log.Fatal(err)
			return nil, &HttpError{Message: "Github Upstream Error", Status: http.StatusBadGateway}
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			return nil, &HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
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

func PaginateStargazers(owner, repo string) ([]byte, *HttpError) {
	allItems, err := PaginateGithub(
		fmt.Sprintf("/repos/%s/%s/stargazers?per_page=100", owner, repo),
		"application/vnd.github.v3.star+json",
	)
	if err != nil {
		return nil, err
	}
	jsonBlob, jsonErr := json.Marshal(allItems)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return nil, &HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
	}
	return jsonBlob, nil
}

func PaginateEvents(owner, repo string) ([]byte, *HttpError) {
	allItems, err := PaginateGithub(
		fmt.Sprintf("/repos/%s/%s/events?per_page=100", owner, repo),
		"application/vnd.github.v3+json",
	)
	if err != nil {
		return nil, err
	}
	jsonBlob, jsonErr := json.Marshal(allItems)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return nil, &HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
	}
	return jsonBlob, nil
}

func PaginateIssues(owner, repo string) ([]byte, *HttpError) {
	allItems, err := PaginateGithub(
		fmt.Sprintf("/repos/%s/%s/issues?per_page=100&state=all&sort=created&direction=asc", owner, repo),
		"application/vnd.github.v3+json",
	)
	if err != nil {
		return nil, err
	}
	jsonBlob, jsonErr := json.Marshal(allItems)
	if jsonErr != nil {
		log.Fatal(jsonErr)
		return nil, &HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
	}
	return jsonBlob, nil
}

func ListStargazers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blob, err := PaginateStargazers(vars["owner"], vars["repo"])
	if err != nil {
		w.WriteHeader(err.Status)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
		return
	}
	w.Write(blob)
}

func ListStarCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allStargazers, err := PaginateGithub(
		fmt.Sprintf("/repos/%s/%s/stargazers?per_page=100", vars["owner"], vars["repo"]),
		"application/vnd.github.v3.star+json",
	)
	if err != nil {
		w.WriteHeader(err.Status)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
		return
	}
	starEvents, decodeErr := DecodeStarEvents(allStargazers)
	if decodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	jsonBlob, jsonErr := json.Marshal(ComputeStarCounts(starEvents))
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	w.Write(jsonBlob)
}

func ListEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blob, err := PaginateEvents(vars["owner"], vars["repo"])
	if err != nil {
		w.WriteHeader(err.Status)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
		return
	}
	w.Write(blob)
}

func ListIssues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	blob, err := PaginateIssues(vars["owner"], vars["repo"])
	if err != nil {
		w.WriteHeader(err.Status)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
		return
	}
	w.Write(blob)
}

func ListOpenIssueCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allIssues, err := PaginateGithub(
		fmt.Sprintf("/repos/%s/%s/issues?per_page=100&state=all&sort=created&direction=asc", vars["owner"], vars["repo"]),
		"application/vnd.github.v3+json",
	)
	if err != nil {
		w.WriteHeader(err.Status)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
		return
	}
	issueEvents, decodeErr := DecodeIssueEvents(allIssues)
	if decodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	jsonBlob, jsonErr := json.Marshal(ComputeOpenIssueCounts(issueEvents))
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	w.Write(jsonBlob)
}

func ListOpenPrCounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	allIssues, err := PaginateGithub(
		fmt.Sprintf("/repos/%s/%s/issues?per_page=100&state=all&sort=created&direction=asc", vars["owner"], vars["repo"]),
		"application/vnd.github.v3+json",
	)
	if err != nil {
		w.WriteHeader(err.Status)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
		return
	}
	prEvents, decodeErr := DecodePrEvents(allIssues)
	if decodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	jsonBlob, jsonErr := json.Marshal(ComputeOpenIssueCounts(prEvents))
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	w.Write(jsonBlob)
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.ServeFile(w, r, path.Join("dashboard", vars["path"]))
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	r := mux.NewRouter()
	r.HandleFunc("/", ServeIndex)
	r.HandleFunc("/dashboard/{path:.*}", ServeStaticFile)
	r.HandleFunc("/gh/{owner}/{repo}/events", ListEvents)
	r.HandleFunc("/gh/{owner}/{repo}/issues", ListIssues)
	r.HandleFunc("/gh/{owner}/{repo}/stargazers", ListStargazers)
	r.HandleFunc("/gh/{owner}/{repo}/star_counts", ListStarCounts)
	r.HandleFunc("/gh/{owner}/{repo}/open_issue_counts", ListOpenIssueCounts)
	r.HandleFunc("/gh/{owner}/{repo}/open_pr_counts", ListOpenPrCounts)
	http.ListenAndServe(":4000", r)
}
