![Elasticsearch Search API](docs/img/elastic-webcrawler.png)

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

### `POST /crawl`

Example POST body:

```JSON
{
    "index": "demo",
    "url": "http://www.google.com"
}
```

Example response:

```JSON
{
    "status": 201,
    "url": "http://www.google.com"
}
```

## Contributors

- [Adam Bemiller](https://github.com/adambemiller)
  - Adam provided most of the high level project and server/routes framework for this project. Huge thanks to him!

## License

MIT License

