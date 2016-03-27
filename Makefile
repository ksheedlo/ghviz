NODE_ENV?=production

.PHONY: all clean jsclean test
all: js go

dashboard/bundle.min.js: dashboard/index.js dashboard/api-client.js dashboard/cache.js dashboard/helpers.js dashboard/components/*.js
	cd dashboard; NODE_ENV=$(NODE_ENV) ./node_modules/.bin/webpack

js: dashboard/bundle.min.js

services/web/web: errors/*.go github/*.go interfaces/*.go middleware/*.go models/*.go routes/*.go services/web/*.go simulate/*.go
	cd services/web; go build

services/prewarm/prewarm: prewarm/*.go github/*.go interfaces/*.go services/prewarm/*.go simulate/*.go
	cd services/prewarm; go build

go: services/prewarm/prewarm services/web/web

clean:
	rm dashboard/bundle.min.js dashboard/*.js.map services/prewarm/prewarm services/web/web

jsclean:
	rm dashboard/bundle.min.js dashboard/*.js.map

test:
	go vet github.com/ksheedlo/ghviz/errors \
		github.com/ksheedlo/ghviz/github \
		github.com/ksheedlo/ghviz/interfaces \
		github.com/ksheedlo/ghviz/middleware \
		github.com/ksheedlo/ghviz/models \
		github.com/ksheedlo/ghviz/prewarm \
		github.com/ksheedlo/ghviz/services/prewarm \
		github.com/ksheedlo/ghviz/services/web \
		github.com/ksheedlo/ghviz/simulate && \
	go test -cover github.com/ksheedlo/ghviz/github \
		github.com/ksheedlo/ghviz/middleware \
		github.com/ksheedlo/ghviz/models \
		github.com/ksheedlo/ghviz/prewarm \
		github.com/ksheedlo/ghviz/routes \
		github.com/ksheedlo/ghviz/simulate && \
	cd dashboard && \
	NODE_ENV=development npm run test
