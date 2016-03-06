.PHONY: all
all: js go

dashboard/bundle.js: dashboard/index.js
	browserify dashboard/index.js -o dashboard/bundle.js

dashboard/bundle.min.js: dashboard/bundle.js
	java -jar compiler.jar --js dashboard/bundle.js --js_output_file dashboard/bundle.min.js -O SIMPLE -W QUIET

js: dashboard/bundle.min.js

ghviz: main.go
	go build

go: ghviz

.PHONY: clean
clean:
	rm ghviz dashboard/bundle.js dashboard/bundle.min.js
