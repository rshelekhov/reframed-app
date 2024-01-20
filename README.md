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
- The usage of [mockery](https://github.com/vektra/mockery) as the mocks generator.

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
