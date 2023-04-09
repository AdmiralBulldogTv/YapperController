FROM golang:1.17.1-alpine3.12 AS build_base

RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /tmp/app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download && go get -u github.com/gobuffalo/packr/v2/packr2

COPY . .

# Build the Go app
RUN packr2 && go build -o yappercontroller

# Start fresh from a smaller image
FROM alpine:3.14

WORKDIR /app

COPY --from=build_base /tmp/app/yappercontroller /app/yappercontroller

RUN apk add --no-cache sox

# Run the binary program produced by `go install`
CMD ["/app/yappercontroller"]
