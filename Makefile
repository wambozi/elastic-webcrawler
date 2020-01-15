OUT := ./bin/elastic-webcrawler
PKG := github.com/wambozi/elastic-webcrawler
VERSION := $(shell git describe --always --long --dirty)
ELASTIC_VERSION := 7.5.1

PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . --name '*.go'  | grep -v /vendor/)

.PHONY: clean
clean:
	-@rm -rf ${OUT} ${OUT}-v*
	for elasticRunner in $$(docker ps -a --filter=name=elastic -q); do \
		docker stop $$elasticRunner; \
		docker rm -f $$elasticRunner; \
	done
	for network in $$(docker network ls | grep testing | awk '{print $$1}'); do \
		docker network rm $$network; \
	done
	

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
test-runner: clean
	[ -d reports ] || mkdir reports
	docker network create testing --subnet=172.18.0.0/16 --gateway=172.18.0.1
	docker run -it --network testing --ip 172.18.0.2 -d --name elastic -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:${ELASTIC_VERSION}
	until $$(curl --output /dev/null --silent --head --fail $$ELASTICSEARCH_ENDPOINT); do \
		printf '.' ; \
		sleep 5 ; \
	done
	curl  -H "Content-Type:application/json" -XPUT $$ELASTICSEARCH_ENDPOINT/test/_doc/1234 -d '{ "title" : "test", "post_date" : "2009-11-15T14:12:12", "message" : "testing out Elasticsearch" }'
	go test --coverprofile=reports/cov.out $$(go list ./... | grep -v /vendor/)
	go tool cover -func=reports/cov.out

.PHONY: test-local
test-local: export ELASTICSEARCH_ENDPOINT=http://localhost:9200
test-local: clean
	[ -d reports ] || mkdir reports
	docker run -it -d --name elastic -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:${ELASTIC_VERSION}
	until $$(curl --output /dev/null --silent --head --fail $$ELASTICSEARCH_ENDPOINT); do \
		printf '.' ; \
		sleep 5 ; \
	done
	curl  -H "Content-Type:application/json" -XPUT $$ELASTICSEARCH_ENDPOINT/test/_doc/1234 -d '{ "title" : "test", "post_date" : "2009-11-15T14:12:12", "message" : "testing out Elasticsearch" }'
	go test --coverprofile=reports/cov.out $$(go list ./... | grep -v /vendor/)
	go tool cover -func=reports/cov.out

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

.PHONY: publish
publish: compile
	docker login --username wambozi --password ${DOCKER_TOKEN}
	docker build -t wambozi/elastic-search-api:${VERSION} .
	docker tag wambozi/elastic-search-api:${VERSION} wambozi/elastic-search-api:latest
	docker push wambozi/elastic-search-api:${VERSION}
	docker push wambozi/elastic-search-api:latest
