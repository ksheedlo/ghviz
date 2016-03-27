package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/ksheedlo/ghviz/github"
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
