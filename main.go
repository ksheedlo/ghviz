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
	"text/template"
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

type IssueAndPrEvent struct {
	EventType IssueEventType
	IsPr      bool
	Timestamp time.Time
}

type ByIprTimestamp []IssueAndPrEvent

func (a ByIprTimestamp) Len() int           { return len(a) }
func (a ByIprTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByIprTimestamp) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }

type OpenIssueAndPrCount struct {
	OpenIssues int
	OpenPrs    int
	Timestamp  time.Time
}

type IndexParams struct {
	Owner string
	Repo  string
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

func DecodeIssueAndPrEvents(apiObjects []map[string]interface{}) ([]IssueAndPrEvent, error) {
	var issueEvents []IssueAndPrEvent
	for i := 0; i < len(apiObjects); i++ {
		issueOpened := IssueAndPrEvent{EventType: IssueOpened}
		_, issueOpened.IsPr = apiObjects[i]["pull_request"]
		createdAt, err := time.Parse(time.RFC3339, apiObjects[i]["created_at"].(string))
		if err != nil {
			return nil, err
		}
		issueOpened.Timestamp = createdAt
		issueEvents = append(issueEvents, issueOpened)

		if closedAt := apiObjects[i]["closed_at"]; closedAt != nil {
			issueClosed := IssueAndPrEvent{EventType: IssueClosed}
			issueClosed.IsPr = issueOpened.IsPr
			closedAt, err := time.Parse(time.RFC3339, closedAt.(string))
			if err != nil {
				return nil, err
			}
			issueClosed.Timestamp = closedAt
			issueEvents = append(issueEvents, issueClosed)
		}
	}
	sort.Sort(ByIprTimestamp(issueEvents))
	return issueEvents, nil
}

func ComputeOpenIssueAndPrCounts(issueEvents []IssueAndPrEvent) []OpenIssueAndPrCount {
	issueCounts := make([]OpenIssueAndPrCount, len(issueEvents))
	openIssues := 0
	openPrs := 0
	for i := 0; i < len(issueEvents); i++ {
		switch {
		case issueEvents[i].EventType == IssueOpened && issueEvents[i].IsPr:
			openPrs++
		case issueEvents[i].EventType == IssueClosed && issueEvents[i].IsPr:
			openPrs--
		case issueEvents[i].EventType == IssueOpened && (!issueEvents[i].IsPr):
			openIssues++
		default:
			openIssues--
		}
		issueCounts[i].OpenIssues = openIssues
		issueCounts[i].OpenPrs = openPrs
		issueCounts[i].Timestamp = issueEvents[i].Timestamp
	}
	return issueCounts
}

func SendGithubRequest(client *http.Client, url, mediaType string) (*http.Response, *HttpError) {
	rr, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		return nil, &HttpError{Message: "Server Error", Status: http.StatusInternalServerError}
	}
	rr.SetBasicAuth(os.Getenv("GITHUB_USERNAME"), os.Getenv("GITHUB_PASSWORD"))
	rr.Header.Add("Accept", mediaType)
	log.Printf("GET %s\n", url)
	resp, err := client.Do(rr)
	if err != nil {
		log.Fatal(err)
		return nil, &HttpError{Message: "Github Upstream Error", Status: http.StatusBadGateway}
	}
	return resp, nil
}

func SendGithubV3Request(client *http.Client, url string) (*http.Response, *HttpError) {
	return SendGithubRequest(client, url, "application/vnd.github.v3+json")
}

func PaginateGithub(path, mediaType string) ([]map[string]interface{}, *HttpError) {
	client := &http.Client{}
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url := fmt.Sprintf("https://api.github.com%s", path); url != ""; {
		resp, httpErr := SendGithubRequest(client, url, mediaType)
		if httpErr != nil {
			log.Fatal(httpErr)
			return nil, httpErr
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

func ListOpenIssuesAndPrs(w http.ResponseWriter, r *http.Request) {
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
	events, decodeErr := DecodeIssueAndPrEvents(allIssues)
	if decodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	jsonBlob, jsonErr := json.Marshal(ComputeOpenIssueAndPrCounts(events))
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	w.Write(jsonBlob)
}

var IndexTpl *template.Template = template.Must(template.ParseFiles("index.tpl.html"))

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	indexParams := IndexParams{Owner: os.Getenv("GHVIZ_OWNER"), Repo: os.Getenv("GHVIZ_REPO")}
	IndexTpl.Execute(w, indexParams)
}

func ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.ServeFile(w, r, path.Join("dashboard", vars["path"]))
}

func TopIssues(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/issues?per_page=100&state=open&sort=created&direction=desc",
		vars["owner"],
		vars["repo"],
	)
	client := &http.Client{}
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url != "" && len(allItems) < 5 {
		resp, httpErr := SendGithubV3Request(client, url)
		if httpErr != nil {
			log.Fatal(httpErr)
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(httpErr.Message))
			return
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		json.Unmarshal(contents, &items)
		for i := 0; i < len(items) && len(allItems) < 5; i++ {
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
	jsonBlob, jsonErr := json.Marshal(allItems)
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	w.Write(jsonBlob)
}

func TopPrs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/issues?per_page=100&state=open&sort=created&direction=desc",
		vars["owner"],
		vars["repo"],
	)
	client := &http.Client{}
	items := make([]map[string]interface{}, 0)
	allItems := make([]map[string]interface{}, 0)

	for url != "" && len(allItems) < 5 {
		resp, httpErr := SendGithubV3Request(client, url)
		if httpErr != nil {
			log.Fatal(httpErr)
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(httpErr.Message))
			return
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		json.Unmarshal(contents, &items)
		for i := 0; i < len(items) && len(allItems) < 5; i++ {
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
	jsonBlob, jsonErr := json.Marshal(allItems)
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Server Error\n"))
		return
	}
	w.Write(jsonBlob)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	r := mux.NewRouter()
	r.HandleFunc("/", ServeIndex)
	r.HandleFunc("/dashboard/{path:.*}", ServeStaticFile)
	r.HandleFunc("/gh/{owner}/{repo}/star_counts", ListStarCounts)
	r.HandleFunc("/gh/{owner}/{repo}/issue_counts", ListOpenIssuesAndPrs)
	r.HandleFunc("/gh/{owner}/{repo}/top_issues", TopIssues)
	r.HandleFunc("/gh/{owner}/{repo}/top_prs", TopPrs)
	http.ListenAndServe(":4000", r)
}
