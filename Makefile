VERSION=0.1.0
DEFAULT_ARGS := --allow-file-uploads

LDFLAGS=-ldflags "-X main.version=${VERSION}"

ROLLUP := node_modules/.bin/rollup

TEMPLATES_IN := $(wildcard src/templates/*.tmpl)
TEMPLATES_OUT := $(patsubst src/%,build/%,$(TEMPLATES_IN))
CSS_IN := $(wildcard src/css/*.css)
CSS_OUT := $(patsubst src/css/%,build/static/%,$(CSS_IN))
JS_IN := $(wildcard src/js/*)
JS_OUT := build/static/notes.bundle.js
GO_IN := $(wildcard src/go/**/*.go)
GO_OUT := build/notes
OUTPUT_FILES := $(GO_OUT) $(JS_OUT) $(TEMPLATES_OUT) $(CSS_OUT)

all: $(OUTPUT_FILES)
	@rsync -a lib/ build/static/lib/
	@rsync -a node_modules/\@fontsource/roboto/files/roboto-all* build/static/lib/
	@rsync -a node_modules/mathjax/es5 build/static/lib/mathjax
	@rsync -a node_modules/mermaid/dist/mermaid.min.js build/static/lib/
	@rsync -a node_modules/jquery/dist/jquery.min.js build/static/lib/
	@rsync -a node_modules/dropzone/dist/min/dropzone* build/static/lib/
	@rsync -a node_modules/github-markdown-css/*css build/static/lib/

$(GO_OUT): $(GO_IN)
	cd src/go && go build ${LDFLAGS} -o ../../build/notes

$(JS_OUT): $(JS_IN)
	$(ROLLUP) $(JS_IN) --file $(JS_OUT) --format iife

build/static/%.css: src/css/%.css
	@mkdir -p `dirname $@`
	cp $< $@

build/templates/%.tmpl: src/templates/%.tmpl
	@mkdir -p `dirname $@`
	cp $< $@

.PHONY: clean
clean:
	rm -rfv build

.PHONY: docker-build
docker-build:
	docker build . --tag isoron/notes:$(VERSION)
	docker tag isoron/notes:$(VERSION) isoron/notes:latest

.PHONY: docker-push
docker-push:
	docker push isoron/notes:$(VERSION)
	docker push isoron/notes:latest

.PHONY: docker-run
docker-run:
	docker run \
	    --userns host \
	    -it \
	    --volume `pwd`/data:/data \
	    --publish 8050:8050 \
	    isoron/notes:$(VERSION)

.PHONY: install-deps
install-deps:
	npm install

.PHONY: install-test-deps
install-test-deps:
	pip install -r src/python/requirements.txt

.PHONY: run
run: all
	cd build && ./notes $(DEFAULT_ARGS) --data ../data

.PHONY: test
test: all
	@cd build                                     ;\
	 ./notes $(DEFAULT_ARGS) --port=8040          &\
	 NOTES_PID=$$!                                ;\
	 cd ..                                        ;\
	 pytest                                       ;\
	 PYTEST_RESULT=$$?                            ;\
  	 kill $$NOTES_PID                             ;\
  	 exit $$PYTEST_RESULT
