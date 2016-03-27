package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"text/template"

	"github.com/gorilla/mux"
	"gopkg.in/redis.v3"

	"github.com/ksheedlo/ghviz/github"
	"github.com/ksheedlo/ghviz/interfaces"
	"github.com/ksheedlo/ghviz/middleware"
	"github.com/ksheedlo/ghviz/routes"
)

var indexTpl *template.Template = template.Must(template.ParseFiles("index.tpl.html"))

func ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.ServeFile(w, r, path.Join("dashboard", vars["path"]))
}

func withDefaultStr(config, default_ string) string {
	if config == "" {
		return default_
	}
	return config
}

func main() {
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
		middleware.AddResponseId(interfaces.RandomTag),
		middleware.AddLogger(os.Stdout),
		middleware.LogRequest,
		middleware.Gzip,
	)
	r.HandleFunc("/", withMiddleware(routes.ServeIndex(&routes.IndexParams{
		Owner: os.Getenv("GHVIZ_OWNER"),
		Repo:  os.Getenv("GHVIZ_REPO"),
	}, indexTpl)))
	r.HandleFunc("/dashboard/{path:.*}", withMiddleware(ServeStaticFile))
	r.HandleFunc(
		"/gh/{owner}/{repo}/star_counts",
		withMiddleware(routes.ListStarCounts(gh)),
	)
	r.HandleFunc(
		"/gh/{owner}/{repo}/issue_counts",
		withMiddleware(routes.ListOpenIssuesAndPrs(gh)),
	)
	r.HandleFunc("/gh/{owner}/{repo}/top_issues", withMiddleware(routes.TopIssues(gh)))
	r.HandleFunc("/gh/{owner}/{repo}/top_prs", withMiddleware(routes.TopPrs(gh)))
	r.HandleFunc(
		"/gh/{owner}/{repo}/highscores/{year:[0-9]+}/{month:(0[1-9]|1[012])}",
		withMiddleware(routes.HighScores(redisClient)),
	)
	http.ListenAndServe(":4000", r)
}
