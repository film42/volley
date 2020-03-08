# Using golang:apline since the other example uses it. Making things easier.
FROM golang:alpine
RUN apk update
RUN apk add --no-cache git
RUN go get github.com/film42/volley/...
RUN mkdir -p /app
ADD example.sh /app/
CMD ["volleyd", "run", "--", "/bin/sh", "-c", "/app/example.sh"]
