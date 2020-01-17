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

### Running Binary

Steps:

1. Launch App Search and Elasticsearch
2. Create a local config file:

```yaml
elasticsearch:
  endpoint: http://localhost:9200
appsearch:
  endpoint: http://localhost:3002
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

This project builds and publishes a container with two tags, `latest` and `commit_hash`, to Docker Hub on merge to master. If you're running the container locally with Elasticsearch and/or App Search running, make sure to run all of them on the same docker network. More about Docker networks can be found [here](https://docs.docker.com/network/).

Docker Hub: [https://hub.docker.com/repository/docker/wambozi/elastic-webcrawler](https://hub.docker.com/repository/docker/wambozi/elastic-webcrawler)

Steps:

1. Launch App Search and Elasticsearch
2. Inspect docker network(s) to get the subnet for your bridge.

```shell
$ docker network inspect bridge
[
  {
     "Name": "bridge",
     "Id": "b38c312777a0f3890034c9b396669842947b80c9051d10a283c9d43937910578",
     "Scope": "local",
     "Driver": "bridge",
     "IPAM": {
     "Driver": "default",
     "Options": null,
     "Config": [
      {
         "Subnet": "172.17.0.2/16" << CIDR for our bridge
      }
    ]
  },
...
]
```

3. Create a local config file in `conf` using the subnet for the network you plan to run the Elasticsearch and App Search containers on. In this case, 172.17.0.2, e.g: 

```yaml
elasticsearch:
  endpoint: http://172.18.0.2:9200
appsearch:
  endpoint: http://172.18.0.2:3002
  api: /api/as/v1/
  token: private-xxxxxxxxxxxxxxx
server:
  port: 8081
  readHeaderTimeoutMillis: 3000
```

4. Run the container. The `docker run` command doesn't need to specify the docker network, as long as we put the subnet for the bridge network in our `local.yml`

```shell
docker run -d --name elastic --network=bridge -it -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "action.auto_create_index=.app-search-*-logs-*,-.app-search-*,+*" docker.elastic.co/elasticsearch/elasticsearch:7.5.1
docker run -d --name app-search --network=bridge -it -p 3002:3002 -e "allow_es_settings_modification=true" docker.elastic.co/app-search/app-search:7.5.1
docker run --rm --name webcrawler -it -e "ENV_ID=local" -v "$(pwd)/conf:/conf" -p 8081:8081 wambozi/elastic-webcrawler:latest
```

- `--network=bridge` : The default network driver. If you don’t specify a driver, this is the type of network you are creating. Bridge networks are usually used when your applications run in standalone containers that need to communicate. I specify it here for transparency.
- `-t` : Allocate a pseudo-tty
- `-i` : Keep STDIN open even if not attached
- `-v` : Mount the current dir into /conf dir of the container (so it makes local.yml accessible here). [Using bind mounts in docker](https://docs.docker.com/storage/bind-mounts/)
- `-e`: Required to specify the name of the env file. If `ENV_ID=local` isn't passed into the container, the container will exit with: `ERRO[0000] stdErr: &{file:0xc0000980c0} , error: Error reading config file. env: nick error: Config File "prod" Not Found in "[/conf /opt/bin/conf /opt/bin]"`. The value of `ENV_ID` should be the name of the file being used.
- `-p` expose the webserver port. This port should correspond to the value for `server.port` in your config.

1. If using App Search, [create the engine](https://swiftype.com/documentation/app-search/getting-started#engine) in App Search (API doesn't create it for you).
2. Launch a crawl:

```shell
curl -XPOST localhost:8081/crawl -d '{
    "engine": "example-website",
    "url": "https://example.com/",
    "type": "app-search"
}'
```

### Run using Makefile

- To run the docker container locally with elasticsearch and app-search, using make: `make run-local`
  - This runs the `docker run` commands above and checks that Elasticsearch is healthy

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

