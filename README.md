# Boxmeup Server

> This is a WIP [Go](https://go-lang.org) implementation of [Boxmeup](https://boxmeupapp.com).

Boxmeup is a web and mobile application to help users keep track of what they have in their containers and how to find items in specific containers.

## Requirements

* [Go >= 1.8.1](https://golang.org) - For local development
* [Docker 17.05.0-ce+](https://www.docker.com) - For building and running in docker containers

## Setup

```bash
docker-compose up -d
```

Bring your own mysql:
```bash
docker run -p 8080:8080 -e MYSQL_DSN=username:password@host:port/database cjsaylor/boxmeup-go
```

See `.env.sample` for available configurations.

## Development

Dependencies are committed into the repo via `godeps`, so no `go install` required.

To build: `go build -o server ./bin`

To add a dependency:

* `go get godep`
* `go get <pkg>`
* Use it somewhere in the code.
* `godep save`