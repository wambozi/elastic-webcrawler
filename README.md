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
  token: private-xxxxxxxxxxxxxxxxx

server:
  port: 8081
  readHeaderTimeoutMillis: 3000
```

## Usage

### Running on Local

Steps:

1. Launch App Search and Elasticsearch
2. Create a local config file:

```yaml
elasticsearch:
  endpoint: http://docker.for.mac.localhost:9200
appsearch:
  endpoint: http://docker.for.mac.localhost:3002
  api: /api/as/v1/
  token: private-xxxxxxxxxxxxxxx
server:
  port: 8081
  readHeaderTimeoutMillis: 3000
```

3. Install vendor dependencies: `go mod vendor`
4. Export env ID: `export ENV_ID=local`
5. Create an env config in `/conf` (example above). The name of this config should match the value of the env ID exported.
6. Compile (required to run the binary locally): `GO_ENABLED=0 go build -mod vendor -o ./bin/elastic-webcrawler ./cmd/elastic-webcrawler/main.go`
7. Run the compiled binary: `./bin/elastic-webcrawler`
8. If using App Search, [create the engine](https://swiftype.com/documentation/app-search/getting-started#engine) in App Search (API doesn't create it for you).
9. Launch a crawl:

```shell
curl -XPOST localhost:8081/crawl -d '{
    "engine": "swiftype-website",
    "url": "https://swiftype.com/",
    "type": "app-search"
}'
```

### Running with Docker

This project builds and publishes a container with two tags, `latest` and `commit_hash`, to Docker Hub on merge to master.

Docker Hub: [https://hub.docker.com/repository/docker/wambozi/elastic-webcrawler](https://hub.docker.com/repository/docker/wambozi/elastic-webcrawler)

Steps:

1. Launch App Search and Elasticsearch
2. Create a local config file:

```yaml
elasticsearch:
  endpoint: http://docker.for.mac.localhost:9200
appsearch:
  endpoint: http://docker.for.mac.localhost:3002
  api: /api/as/v1/
  token: private-xxxxxxxxxxxxxxx
server:
  port: 8081
  readHeaderTimeoutMillis: 3000
```

3. Run the container.

```shell
docker pull wambozi/elastic-webcrawler:latest
docker run --rm -it -e "ENV_ID=local" -v "/some/path/to/conf:/conf" -p 8081:8081 wambozi/elastic-webcrawler:latest 
```

- `-v` : Mount the current dir into /conf dir of the container (so it makes local.yml accessible here). [Using bind mounts in docker](https://docs.docker.com/storage/bind-mounts/)
- `-e`: Required to specify the name of the env file. If `ENV_ID=local` isn't passed into the container, the container will exit with: `error: Error reading config file: Config File "no-config-set" Not Found in "[/conf /opt/bin/conf /opt/bin]"`
- `-p` expose the webserver port. This port should correspond to the value for `server.port` in your config.

4. If using App Search, [create the engine](https://swiftype.com/documentation/app-search/getting-started#engine) in App Search (API doesn't create it for you).
5. Launch a crawl:

```shell
curl -XPOST localhost:8081/crawl -d '{
    "engine": "swiftype-website",
    "url": "https://swiftype.com/",
    "type": "app-search"
}'
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

