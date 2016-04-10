package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/middleware"
	"github.com/ksheedlo/ghviz/models"
	"github.com/ksheedlo/ghviz/simulate"
)

func ListStarCounts(gh github.ListStarEventser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		starEvents, err := gh.ListStarEvents(logger, vars["owner"], vars["repo"])
		if err != nil {
			w.WriteHeader(err.Status)
			fmt.Fprintf(w, "%s\n", err.Message)
			return
		}
		// Suppress JSON marshaling errors because we know we can always
		// marshal `simulate.StarCount`s.
		jsonBlob, _ := json.Marshal(simulate.StarCounts(starEvents))
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

func ListOpenIssuesAndPrs(gh github.ListIssueser) http.HandlerFunc {
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
		// Suppress JSON marshaling errors because we know we can always
		// marshal `simulate.OpenIssueAndPrCount`s.
		jsonBlob, _ := json.Marshal(simulate.OpenIssueAndPrCounts(events))
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

func TopIssues(gh github.ListTopIssueser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		allItems, httpErr := gh.ListTopIssues(logger, vars["owner"], vars["repo"], 5)
		if httpErr != nil {
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", httpErr.Message)))
			return
		}
		// Suppress JSON marshaling errors because we know we can always
		// marshal `github.Issue`s.
		jsonBlob, _ := json.Marshal(allItems)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

func TopPrs(gh github.ListTopPrser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := context.Get(r, middleware.CtxLog).(*log.Logger)
		vars := mux.Vars(r)
		allItems, httpErr := gh.ListTopPrs(logger, vars["owner"], vars["repo"], 5)
		if httpErr != nil {
			w.WriteHeader(httpErr.Status)
			w.Write([]byte(fmt.Sprintf("%s\n", httpErr.Message)))
			return
		}
		// Suppress JSON marshaling errors because we know we can always
		// marshal `github.Issue`s.
		jsonBlob, _ := json.Marshal(allItems)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}

func HighScores(redis interfaces.Rediser) http.HandlerFunc {
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
			endMonth = 1
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
		top := 5
		if len(highScores) < top {
			top = len(highScores)
		}
		// Suppress JSON marshaling errors because we know we can always
		// marshal `simulate.ActorScore`s.
		jsonBlob, _ := json.Marshal(highScores[:top])
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBlob)
	}
}
