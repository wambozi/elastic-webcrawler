![Elasticsearch Search API](docs/img/elastic-webcrawler.png)

[![Sonarcloud Status](https://sonarcloud.io/api/project_badges/measure?project=wambozi_elastic-webcrawler&metric=coverage)](https://sonarcloud.io/dashboard?id=wambozi_elastic-webcrawler)

[![Release](https://github.com/wambozi/elastic-webcrawler/workflows/Release/badge.svg)](https://github.com/wambozi/elastic-webcrawler/)


## Description

Golang API that indexes web pages in Elasticsearch. Accepts POST requests and runs a crawl in the background.

## Dependencies

- `go 1.13.5^`
- `Elasticsearch v7.5.1^`

## Configuration

Requires an config yaml in `conf`.

For instance:

Path: `/conf/local.yml`

```YAML
elasticsearch:
  endpoint: http://localhost:9200
  password: changeme
  username: elastic

appsearch:
  endpoint: http://localhost:3002
  api: /api/as/v1/
  token: private-pq7aaoSDFapSADosdnfns

server:
  port: 8081
  readHeaderTimeoutMillis: 3000
```

## Usage

### Local

To run the binary locally:

- install vendor dependencies: `go mod vendor`
- create an env config in `/conf` (example above)
- compile (required to run the binary locally): `GO_ENABLED=0 go build -mod vendor -o ./bin/elastic-webcrawler ./cmd/elastic-webcrawler/main.go`
- run the compiled binary: `./bin/elastic-webcrawler`

### Docker

This project builds and publishes a container with two tags, `latest` and `commit_hash`, to Docker Hub on merge to master. Currently, you have to mount a volume with your config to run it.

Docker Hub: [https://hub.docker.com/repository/docker/wambozi/elastic-webcrawler](https://hub.docker.com/repository/docker/wambozi/elastic-webcrawler)

To run:

```shell
docker pull wambozi/elastic-webcrawler:latest
docker run --rm -it -e "ENV_ID=local" -v "/some/path/conf:/conf" -p 8081:8081 wambozi/elastic-webcrawler:latest 
```

### `POST /crawl`

Example POST body for an Elasticsearch crawl:

```JSON
{
    "index": "demo",
    "url": "http://www.example.com",
    "type": "elasticsearch"
}
```

Example POST body for an AppSearch crawl:

```JSON
{
    "engine": "demo",
    "url": "http://www.example.com",
    "type": "app-search"
}
```

Example response:

```JSON
{
    "status": 201,
    "url": "http://www.example.com"
}
```

## Contributors

- [Adam Bemiller](https://github.com/adambemiller)
  - Adam provided most of the high level project and server/routes framework for this project. Huge thanks to him!

## License

MIT License

