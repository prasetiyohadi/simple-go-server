FROM golang:1.16-alpine3.13 as builder
WORKDIR /app
COPY . /app
RUN go mod download\
 && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/bin/app main.go

FROM alpine:3.13
COPY --from=builder /usr/local/bin/app /usr/local/bin/app
EXPOSE 8080
CMD ["app"]
