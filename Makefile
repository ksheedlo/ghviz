.PHONY: all
all: js go

dashboard/bundle.js: dashboard/index.js
	browserify dashboard/index.js -o dashboard/bundle.js

dashboard/bundle.min.js: dashboard/bundle.js
	java -jar compiler.jar --js dashboard/bundle.js --js_output_file dashboard/bundle.min.js -O SIMPLE -W QUIET

js: dashboard/bundle.min.js

github-visualizer: main.go
	go build

go: github-visualizer

.PHONY: clean
clean:
	rm github-visualizer dashboard/bundle.js dashboard/bundle.min.js
