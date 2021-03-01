FROM golang:1.16-alpine3.13 as builder
WORKDIR /app
COPY . /app
RUN go mod download\
 && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o /app/app main.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/app /app/app
EXPOSE 8080
CMD ["./app"]
