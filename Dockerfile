FROM golang:1.8.3-alpine as builder
COPY . /go/src/github.com/cjsaylor/boxmeup-go
WORKDIR /go/src/github.com/cjsaylor/boxmeup-go
RUN GOOS=linux GOARCH=386 go build -v

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN adduser -D -u 1000 appuser
USER appuser
WORKDIR /app
COPY --from=builder /go/src/github.com/cjsaylor/boxmeup-go/boxmeup-go app
EXPOSE 8080

CMD ["./app"]