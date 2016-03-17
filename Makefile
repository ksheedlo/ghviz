.PHONY: all clean test
all: js go

dashboard/bundle.min.js: dashboard/index.js dashboard/cache.js dashboard/helpers.js dashboard/ops.js dashboard/components/*.js
	cd dashboard; NODE_ENV=production ./node_modules/.bin/webpack --devtool source-map

js: dashboard/bundle.min.js

services/web/web: errors/*.go github/*.go interfaces/*.go middleware/*.go models/*.go services/web/*.go simulate/*.go
	cd services/web; go build

highscores/highscores: highscores/main.go github/*.go simulate/*.go
	cd highscores; go build

services/prewarm/prewarm: prewarm/*.go github/*.go interfaces/*.go services/prewarm/*.go simulate/*.go
	cd services/prewarm; go build

go: highscores/highscores services/prewarm/prewarm services/web/web

clean:
	rm dashboard/bundle.min.js dashboard/*.js.map highscores/highscores services/prewarm/prewarm services/web/web

test:
	go test github.com/ksheedlo/ghviz/github \
		github.com/ksheedlo/ghviz/models \
		github.com/ksheedlo/ghviz/simulate && \
	cd dashboard && \
	npm run lint
