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
	"github.com/ksheedlo/ghviz/simulate"
)

func ListStarCounts(gh github.ListStarEventser) func(http.ResponseWriter, *http.Request) {
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
