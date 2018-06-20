FROM golang:1.9.4 as build
  
RUN curl -sL https://github.com/alexellis/license-check/releases/download/0.2.2/license-check \
      > /usr/bin/license-check \
      && chmod +x /usr/bin/license-check

WORKDIR /go/src/github.com/s8sg/nginx_faas/origin

COPY main.go        .

# Run a gofmt and exclude all vendored code.
RUN test -z "$(gofmt -l $(find . -type f -name '*.go' -not -path "./vendor/*"))" \
    && go test $(go list ./... | grep -v integration | grep -v /vendor/ | grep -v /template/) -cover \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

FROM       scratch
MAINTAINER s8sg <swarvanusg@gmail.com>

COPY       --from=build /go/src/github.com/s8sg/nginx_faas/origin/server tc-tiny-go-ws
EXPOSE     8080
ENTRYPOINT ["/tc-tiny-go-ws"]
