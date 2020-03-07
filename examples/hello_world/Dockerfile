FROM golang:alpine
RUN apk update
RUN apk add --no-cache git
RUN go get github.com/film42/volley/...
ADD hello_world.go /app/
RUN mkdir -p /app
RUN go build -o /app/hello_world /app/hello_world.go
CMD ["volleyd", "run", "/app/hello_world"]
