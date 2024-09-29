# Reframed

> Reframed is like your own personal assistant for organizing your day, keeping track of your projects, and moving closer to your goals.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

Backend is written in GO, so I suggest you have installed [Golang](https://golang.org).

Also, you need to install [golang-migrate](https://github.com/golang-migrate/migrate) tool for running database migrations.

And the last one, you need to have `psql` (PostgreSQL interactive terminal), because this tool  is used in the commands described in the makefile.

### Installing

Add config file to `./config/.env` (see an example in the `./config/.env.example`).

Set path to config:
```bash
export CONFIG_PATH=./config/.env
```

Set URL for PostgresQL:
```bash
export POSTGRESQL_URL='postgres://login:password@host:port/db_name?sslmode=disable'
```

By default, this app use port `8082`. Set `SERVER_PORT` environment variable with this or your value:
```bash
export SERVER_PORT=8082
```

Run migrations using `make migrate` command.

Then run the app using make `run-server` command.

Make sure you deployed and run the SSO [gRPC server](https://github.com/rshelekhov/sso)

## Running the tests

For testing the functionality of the application, both unit tests for individual functions and end-to-end tests for checking the entire application are used.

### Break down into functional tests

Functional tests allow to verify the application by sending real requests to the HTTP server and checking the received responses.

Run tests — `make test-all-app` or `make test-api`. You'll run database migrations (or check if you did it before), insert test-app into database, run the server and then run tests.

For more details you can see other commands in the `Makefile`.

### And coding style tests

This project uses linters to ensure code quality, consistency, and to catch potential issues such as code smells, bugs, or performance concerns early in the development process. Linters help maintain clean and efficient code by automatically checking it against predefined rules and best practices.

Linters are managed using the library https://github.com/golangci/golangci-lint.

To run the linters manually, use the command `make lint`.

## Deployment

To deploy the application on your server, you can download the latest image from Docker Hub directly to your server. To do this, you need to connect to the server via SSH and ensure that Docker is installed. Then, execute the following command:
```
docker pull rshelekhov/reframed-app:latest
```

To run the container, you need to place a config file in a volume and set values for the `CONFIG_PATH` and `POSTGRESQL_URL` variables. Additionally, you need to specify the port in the docker run command parameters. Here’s an example command to run the container:
```
docker run -d \
    -v ${PWD}/config/reframed-app:/src/config \
    -e CONFIG_PATH=/src/config/.env \
    -e POSTGRESQL_URL=postgres://user:password@127.0.0.1:5432/reframed_dev?sslmode=disable \
    -p 8082:8082 \
    --name reframed-app \
    reframed-app:latest
```

You can also check the settings for GitHub Actions in the `.github/workflows` folder to see how the application is deployed on dev server.

## What's included
- The REST API with JSON responses
- PostgresQL as a main database
- [viper](https://github.com/spf13/viper) as a complete configuration solution for Go applications including [12-Factor apps](https://12factor.net/#the_twelve_factors)
- [golang-migrate](https://github.com/golang-migrate/migrate) for the database migrations
- [sqlc](https://github.com/sqlc-dev/sqlc) as the generator type-safe code from SQL
- [log/slog](https://pkg.go.dev/log/slog) as the centralized Syslog logger
- [go-chi](https://github.com/go-chi/chi) as the HTTP router
- [validator](https://github.com/go-playground/validator) as the form validator
- [ksuid](https://github.com/segmentio/ksuid) as the unique identifier
- [golangci-lint](https://github.com/golangci/golangci-lint) as a Go linters runner


## Documentation

You can see more details about API in [the documentation](https://www.postman.com/warped-crater-962061/workspace/reframed) in Postman.

## Reframed App UI design

![alt text](https://github.com/rshelekhov/reframed/blob/main/internal/lib/img/browser-1.png)
![alt text](https://github.com/rshelekhov/reframed/blob/main/internal/lib/img/browser-2.png)
![alt text](https://github.com/rshelekhov/reframed/blob/main/internal/lib/img/browser-3.png)
![alt text](https://github.com/rshelekhov/reframed/blob/main/internal/lib/img/browser-4.png)