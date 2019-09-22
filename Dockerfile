FROM  golang:1.13.0

WORKDIR /go/src/github.com/stefansaftic/docker-gelf-multi-log-driver
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . /go/src/github.com/stefansaftic/docker-gelf-multi-log-driver

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/bin/gelf-multi-log-driver

FROM alpine:3.7
RUN apk --no-cache add ca-certificates
COPY --from=0 /usr/bin/gelf-multi-log-driver /usr/bin/
WORKDIR /usr/bin/
ENTRYPOINT ["/usr/bin/gelf-multi-log-driver"]