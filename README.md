# Reframed

> ...

## What's included
- The idiomatic structure based on the resource-oriented design.
- The REST API with JSON responses.
- The usage of [golang-migrate](https://github.com/golang-migrate/migrate) for the database migrations
- The usage of [log/slog](https://pkg.go.dev/log/slog) as the centralized Syslog logger.
- The usage of [go-chi](https://github.com/go-chi/chi) as the HTTP router.
- The usage of [validator](https://github.com/go-playground/validator) as the form validator.
- The usage of [ksuid](https://github.com/segmentio/ksuid) as the unique identifier.

## Endpoints

| Name        | HTTP Method | Route         |
|-------------|-------------|---------------|
| Health      | GET         | /health       |
|             |             |               |
| Create user | POST        | /users        |
| Get user    | GET         | /users/{id}   |
| Get users   | GET         | /users/       |
| Update user | PUT         | /users/{id}   |
| Delete user | DELETE      | /users/{id}   |
| Get roles   | GET         | /users/roles  |


## Database design

| Column Name    | Datatype | Not Null | Primary Key |
|----------------|----------|----------|-------------|
| id             | ksuid    | ✅       | ✅          |

## Project structure

```shell
remedi
├── cmd
│  ├── migrate
│  │  └── main.go
│  └── remedi
│     └── main.go
│
├── internal
│  ├── config
│  │  └── config.go
│  │
│  ├── http-server
│  │  ├── middleware
│  │  ├── router
│  │  └── server
│  │
│  ├── lib
│  │  ├── api
│  │  │  ├── model
│  │  │  ├── parser
│  │  │  └── response
│  │  │
│  │  └── logger
│  │
│  ├── resource
│  │  ├── common
│  │  │  ├── err
│  │  │  ├── helper
│  │  │  └── model
│  │  │
│  │  ├── health
│  │  │
│  │  └── resourceName
│  │     ├── handlers.go
│  │     ├── model.go
│  │     ├── service.go
│  │     └── storage.go
│  │
│  └── storage
│     ├── postgres
│     │  └── postgres.go
│     └── storage.go
│
├── migrations
│  ├── 000001_init.down.sql
│  ├── 000001_init.up.sql
│  ├── 000002_zero_dump.down.sql
│  └── 000002_zero_dump.up.sql
│
├── test
│
├── .env.example
│
├── .gitignore
│
├── go.mod
│
├── Makefile

```