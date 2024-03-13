# Reframed

> ...

## What's included
- The REST API with JSON responses.
- The usage of [golang-migrate](https://github.com/golang-migrate/migrate) for the database migrations
- The usage of [log/slog](https://pkg.go.dev/log/slog) as the centralized Syslog logger.
- The usage of [go-chi](https://github.com/go-chi/chi) as the HTTP router.
- The usage of [validator](https://github.com/go-playground/validator) as the form validator.
- The usage of [ksuid](https://github.com/segmentio/ksuid) as the unique identifier.
- The usage of [sqlc](https://github.com/sqlc-dev/sqlc) as the generator type-safe code from SQL.
- The usage of [golangci-lint](https://github.com/golangci/golangci-lint) as a Go linters runner.

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


## Reframed App UI design

![alt text](https://github.com/rshelekhov/reframed/blob/main/pkg/img/browser-1.png)

![alt text](https://github.com/rshelekhov/reframed/blob/main/pkg/img/browser-2.png)

![alt text](https://github.com/rshelekhov/reframed/blob/main/pkg/img/browser-3.png)

![alt text](https://github.com/rshelekhov/reframed/blob/main/pkg/img/browser-4.png)