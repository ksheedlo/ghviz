package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"gopkg.in/redis.v3"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
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

func HighScores(redis interfaces.Rediser) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		owner := vars["owner"]
		repo := vars["repo"]
		yearString := vars["year"]
		monthString := vars["month"]
		startYear, parseErr := strconv.Atoi(yearString)
		if parseErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(
				`{"type":"error","code":400,"message":"%s is not a valid year"}`,
				yearString,
			)))
			return
		}
		startMonth, parseErr := strconv.Atoi(monthString)
		if parseErr != nil || startMonth < 1 || 12 < startMonth {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(
				`{"type":"error","code":400,"message":"%s is not a valid month between 01-12"}`,
				monthString,
			)))
			return
		}
		endYear := startYear
		endMonth := startMonth + 1
		if endMonth == 13 {
			endYear = startYear + 1
			startMonth = 1
		}
		startDate := strconv.FormatInt(
			time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC).Unix(),
			10,
		)
		endDate := strconv.FormatInt(
			time.Date(endYear, time.Month(endMonth), 1, 0, 0, 0, 0, time.UTC).Unix(),
			10,
		)
		eventSetId, err := redis.Get(
			fmt.Sprintf("gh:repos:%s:%s:issue_event_setid", owner, repo),
		)
		if err != nil || eventSetId == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(
				`{"type":"error","code":404,"message":"Scores for %s/%s were not found."}`,
				owner,
				repo,
			)))
			return
		}

		scoringEventJsons, redisErr := redis.ZRangeByScore(
			fmt.Sprintf("gh:repos:%s:%s:issue_events:%s", owner, repo, eventSetId),
			&interfaces.ZRangeByScoreOpts{Min: startDate, Max: endDate},
		)
		if redisErr != nil {
			logger.Printf("ERROR: %s\n", redisErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"type":"error","code":500,"message":"Internal Server Error"}`))
			return
		}
		var eventsToScore []simulate.ScoringEvent
		for _, scoringEventJson := range scoringEventJsons {
			scoringEvent := simulate.ScoringEvent{}
			if jsonErr := json.Unmarshal([]byte(scoringEventJson), &scoringEvent); jsonErr != nil {
				logger.Printf("error: %s\n", jsonErr.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"type":"error","code":500,"message":"Internal Server Error"}`))
				return
			}
			eventsToScore = append(eventsToScore, scoringEvent)
		}
		highScores := simulate.ScoreEvents(eventsToScore)
		sort.Sort(sort.Reverse(simulate.ByScore(highScores)))
		top := 8
		if len(highScores) < top {
			top = len(highScores)
		}
		jsonBlob, jsonErr := json.Marshal(highScores[:top])
		if jsonErr != nil {
			logger.Printf("error: %s\n", jsonErr.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"type":"error","code":500,"message":"Internal Server Error"}`))
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

	var redisClient interfaces.Rediser
	if redisHost := os.Getenv("GHVIZ_REDIS_HOST"); redisHost != "" {
		redisPort := withDefaultStr(os.Getenv("GHVIZ_REDIS_PORT"), "6379")
		redisClient = interfaces.NewGoRedis(redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password: os.Getenv("GHVIZ_REDIS_PASSWORD"),
			DB:       0,
		}))
	}

	gh := github.NewClient(&github.Options{
		MaxStaleness: 5,
		RedisClient:  redisClient,
		Token:        os.Getenv("GITHUB_TOKEN"),
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
	r.HandleFunc("/gh/{owner}/{repo}/highscores/{year:[0-9]+}/{month:(0[1-9]|1[012])}", withMiddleware(HighScores(redisClient)))
	http.ListenAndServe(":4000", r)
}
