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


## Database design

| Column Name    | Datatype | Not Null | Primary Key |
|----------------|----------|----------|-------------|
| id             | ksuid    | ✅       | ✅          |

## Project structure

```shell
reframed
├── cmd
│  ├── migrate
│  │  └── main.go
│  └── reframed
│     └── main.go
│
├── config
│  └── config.go
│
├── internal
│  ├── api
│  │  │── controller
│  │  │   ├── health.go
│  │  │   ├── helpers.go
│  │  │   └── entityName.go
│  │  └── route
│  │      ├── route.go
│  │      └── entityName.go
│  │
│  ├── entity
│  │  └── entityName.go
│  │
│  ├── lib
│  │  ├── api
│  │  │  ├── parser
│  │  │  └── response
│  │  │
│  │  └── logger
│  │
│  └── usecase
│     ├── storage
│     │  ├── storage.go
│     │  └── entityName.go
│     │
│     │── interfaces.go
│     └── entityName.go
│   
├── migrations
│  ├── 000001_init.down.sql
│  ├── 000001_init.up.sql
│  ├── 000002_zero_dump.down.sql
│  └── 000002_zero_dump.up.sql
│
├── pkg
│  ├── http-server
│  │  ├── middleware
│  │  └── server.go
│  │
│  ├── logger
│  │  └── logger.go
│  │
│  └── storage
│     └── postgres
│        └── postgres.go
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