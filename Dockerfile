# Stage 1: Build the application
FROM golang:1.22-alpine3.19 AS builder

WORKDIR /src

# Setup base software for building an app
RUN apk update && apk add --no-cache ca-certificates git make

# Install golang-migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download -x && go mod verify

# Copy application source
COPY . .

# Build the application
RUN go build -o /app ./cmd/reframed

# Stage 2: Prepare the final runtime image
FROM golang:1.22-alpine3.19 AS runner

RUN apk update && apk add --no-cache ca-certificates make postgresql-client

WORKDIR /src

COPY --from=builder /app ./
COPY --from=builder /src/Makefile ./
COPY --from=builder /src/migrations ./migrations
COPY --from=builder /src/api_tests ./api_tests
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /src/go.mod ./
COPY --from=builder /src/go.sum ./

CMD ["./app"]