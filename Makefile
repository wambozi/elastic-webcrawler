OUT := ./bin/elastic-webcrawler
PKG := github.com/wambozi/elastic-webcrawler
VERSION := $(shell git describe --always --long --dirty)
ELASTIC_VERSION := 7.5.1

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . --name '*.go'  | grep -v /vendor/)

BRIDGE_CIDR := $(shell docker network inspect bridge | grep "Gateway" | sed -e 's/^.*Gateway\"\://g' | sed 's/\"//g' | sed 's/ //g')

.PHONY: clean-test
clean-test:
	-@rm -rf ${OUT} ${OUT}-v*
	for elasticRunner in $$(docker ps -a --filter=name=elastic-test -q); do \
		docker stop $$elasticRunner; \
		docker rm -f $$elasticRunner; \
	done
	for elasticRunner in $$(docker ps -a --filter=name=app-search-test -q); do \
		docker stop $$elasticRunner; \
		docker rm -f $$elasticRunner; \
	done
	for network in $$(docker network ls | grep testing | awk '{print $$1}'); do \
		docker network rm $$network; \
	done

.PHONY: clean-local
clean-local:
	-@rm -rf ${OUT} ${OUT}-v*
	for elasticRunner in $$(docker ps -a --filter=name=elastic-local -q); do \
		docker stop $$elasticRunner; \
		docker rm -f $$elasticRunner; \
	done
	for elasticRunner in $$(docker ps -a --filter=name=app-search-local -q); do \
		docker stop $$elasticRunner; \
		docker rm -f $$elasticRunner; \
	done
	for network in $$(docker network ls | grep elastic | awk '{print $$1}'); do \
		docker network rm $$network; \
	done

.PHONY: runner-deps
runner-deps:
	go get -v -t -d ./...
	if [ -f Gopkg.toml ]; then \
		curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh; \
		dep ensure; \
	fi

.PHONY: compile
compile:
	go env -w GOPRIVATE=github.com/wambozi/*
	go mod vendor
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -o ${OUT} -ldflags="-extldflags \"-static\"" ./cmd/elastic-webcrawler/main.go

.PHONY: format
format:
	@gofmt -w *.go $$(ls -d */ | grep -v /vendor/)

.PHONY: test-runner
test-runner: export ELASTICSEARCH_ENDPOINT=http://172.18.0.2:9200
test-runner: clean-test
	[ -d reports ] || mkdir reports
	docker network create testing --subnet=172.18.0.0/16 --gateway=172.18.0.1
	docker run -it -d --network testing --ip 172.18.0.2 --name elastic-test -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "action.auto_create_index=.app-search-*-logs-*,-.app-search-*,+*" docker.elastic.co/elasticsearch/elasticsearch:${ELASTIC_VERSION}
	until $$(curl --output /dev/null --silent --head --fail $$ELASTICSEARCH_ENDPOINT); do \
		printf '.' ; \
		sleep 5 ; \
	done
	docker run -it -d --network testing --ip 172.18.0.2 --name app-search-test -p 3002:3002 -e "elasticsearch.host=$$ELASTICSEARCH_ENDPOINT" docker.elastic.co/app-search/app-search:${ELASTIC_VERSION}
	curl  -H "Content-Type:application/json" -XPUT $$ELASTICSEARCH_ENDPOINT/test/_doc/1234 -d '{ "title" : "test", "post_date" : "2009-11-15T14:12:12", "message" : "testing out Elasticsearch" }'
	go test --coverprofile=reports/cov.out $$(go list ./... | grep -v /vendor/)
	go tool cover -func=reports/cov.out

.PHONY: test-local
test-local: export ELASTICSEARCH_ENDPOINT=http://localhost:9200
test-local: clean-local clean-test
	[ -d reports ] || mkdir reports
	@echo "Bridge CIDR: ${BRIDGE_CIDR}";
	docker run -it -d --name elastic-test -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "action.auto_create_index=.app-search-*-logs-*,-.app-search-*,+*" docker.elastic.co/elasticsearch/elasticsearch:${ELASTIC_VERSION}
	until $$(curl --output /dev/null --silent --head --fail $$ELASTICSEARCH_ENDPOINT); do \
		printf '.' ; \
		sleep 5 ; \
	done
	docker run -it -d --name app-search-test -p 3002:3002 -e "elasticsearch.host=http://${BRIDGE_CIDR}:9200" docker.elastic.co/app-search/app-search:${ELASTIC_VERSION}
	curl  -H "Content-Type:application/json" -XPUT $$ELASTICSEARCH_ENDPOINT/test/_doc/1234 -d '{ "title" : "test", "post_date" : "2009-11-15T14:12:12", "message" : "testing out Elasticsearch" }'
	go test --coverprofile=reports/cov.out $$(go list ./... | grep -v /vendor/)
	go tool cover -func=reports/cov.out

.PHONY: run-local
run-local: clean-local
	@echo "Bridge CIDR: ${BRIDGE_CIDR}";
	docker run -d --name elastic-local -it -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "action.auto_create_index=.app-search-*-logs-*,-.app-search-*,+*" docker.elastic.co/elasticsearch/elasticsearch:${ELASTIC_VERSION}
	until $$(curl --output /dev/null --silent --head --fail http://localhost:9200); do \
		printf '.' ; \
		sleep 5 ; \
	done
	docker run -d --name app-search-local -it -p 3002:3002 -e "elasticsearch.host=http://${BRIDGE_CIDR}:9200" docker.elastic.co/app-search/app-search:${ELASTIC_VERSION}
	docker run --rm --name webcrawler -it -e "ENV_ID=local" -v "$$(pwd)/conf:/conf" -p 8081:8081 wambozi/elastic-webcrawler:latest

.PHONY: vet
vet:
	@go vet .

.PHONY: lint
lint:
	@for file in ${GO_FILES}; do \
		golint $$file; \
	done

.PHONY: sonar
sonar:
	gitlab-sonar-scanner -Dsonar.login=${SONAR_USER_TOKEN}

.PHONY: build
build: compile
	docker build -t wambozi/elastic-webcrawler:${VERSION} .
	docker tag wambozi/elastic-webcrawler:${VERSION} wambozi/elastic-webcrawler:latest

.PHONY: publish
publish: compile
	docker login --username wambozi --password ${DOCKER_TOKEN}
	docker push wambozi/elastic-webcrawler:${VERSION}
	docker push wambozi/elastic-webcrawler:latest
