package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"text/template"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"gopkg.in/redis.v3"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/middleware"
	"github.com/ksheedlo/ghviz/models"
	"github.com/ksheedlo/ghviz/simulate"
)

type IndexParams struct {
	Owner string
	Repo  string
}

func ListStarCounts(gh *github.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		allStargazers, err := gh.ListStargazers(logger, vars["owner"], vars["repo"])
		if err != nil {
			w.WriteHeader(err.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
			return
		}
		starEvents, decodeErr := models.StarEventsFromApi(allStargazers)
		if decodeErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		jsonBlob, jsonErr := json.Marshal(simulate.StarCounts(starEvents))
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

func ListOpenIssuesAndPrs(gh *github.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		allIssues, err := gh.ListIssues(logger, vars["owner"], vars["repo"])
		if err != nil {
			w.WriteHeader(err.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", err.Message)))
			return
		}
		events := models.IssueEventsFromApi(allIssues)
		jsonBlob, jsonErr := json.Marshal(simulate.OpenIssueAndPrCounts(events))
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

var IndexTpl *template.Template = template.Must(template.ParseFiles("index.tpl.html"))

func ServeIndex(params *IndexParams) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		IndexTpl.Execute(w, params)
	}
}

func ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.ServeFile(w, r, path.Join("dashboard", vars["path"]))
}

func TopIssues(gh *github.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		allItems, httpErr := gh.ListTopIssues(logger, vars["owner"], vars["repo"], 5)
		if httpErr != nil {
			log.Fatal(httpErr)
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", httpErr.Message)))
			return
		}
		jsonBlob, jsonErr := json.Marshal(allItems)
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

func TopPrs(gh *github.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		allItems, httpErr := gh.ListTopPrs(logger, vars["owner"], vars["repo"], 5)
		if httpErr != nil {
			log.Fatal(httpErr)
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", httpErr.Message)))
			return
		}
		jsonBlob, jsonErr := json.Marshal(allItems)
		if jsonErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server Error\n"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

func withDefaultStr(config, default_ string) string {
	if config == "" {
		return default_
	}
	return config
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	r := mux.NewRouter()

	var redisClient *redis.Client
	if redisHost := os.Getenv("GHVIZ_REDIS_HOST"); redisHost != "" {
		redisPort := withDefaultStr(os.Getenv("GHVIZ_REDIS_PORT"), "6379")
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password: os.Getenv("GHVIZ_REDIS_PASSWORD"),
			DB:       0,
		})
	}

	gh := github.NewClient(&github.Options{
		RedisClient: redisClient,
		Token:       os.Getenv("GITHUB_TOKEN"),
	})
	withMiddleware := middleware.Compose(
		middleware.AddResponseId,
		middleware.AddLogger,
		middleware.LogRequest,
		middleware.Gzip,
	)
	r.HandleFunc("/", withMiddleware(ServeIndex(&IndexParams{
		Owner: os.Getenv("GHVIZ_OWNER"),
		Repo:  os.Getenv("GHVIZ_REPO"),
	})))
	r.HandleFunc("/dashboard/{path:.*}", withMiddleware(ServeStaticFile))
	r.HandleFunc("/gh/{owner}/{repo}/star_counts", withMiddleware(ListStarCounts(gh)))
	r.HandleFunc("/gh/{owner}/{repo}/issue_counts", withMiddleware(ListOpenIssuesAndPrs(gh)))
	r.HandleFunc("/gh/{owner}/{repo}/top_issues", withMiddleware(TopIssues(gh)))
	r.HandleFunc("/gh/{owner}/{repo}/top_prs", withMiddleware(TopPrs(gh)))
	http.ListenAndServe(":4000", r)
}
