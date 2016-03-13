.PHONY: all clean test
all: js go

dashboard/bundle.js: dashboard/index.js dashboard/cache.js dashboard/helpers.js dashboard/ops.js dashboard/components/*.js
	cd dashboard; NODE_ENV=production ./node_modules/.bin/browserify index.js -o bundle.js -t [ babelify --presets [ es2015 react ] ]

dashboard/bundle.min.js: dashboard/bundle.js
	java -jar compiler.jar --js dashboard/bundle.js --js_output_file dashboard/bundle.min.js -O SIMPLE -W QUIET

js: dashboard/bundle.min.js

ghviz: main.go errors/*.go github/*.go middleware/*.go models/*.go simulate/*.go
	go build

go: ghviz

clean:
	rm ghviz dashboard/bundle.js dashboard/bundle.min.js

test:
	go test github.com/ksheedlo/ghviz/github github.com/ksheedlo/ghviz/models github.com/ksheedlo/ghviz/simulate
