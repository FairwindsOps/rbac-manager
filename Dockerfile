FROM golang:1.13 AS build-env
ENV GO111MODULE on
WORKDIR /go/src/github.com/FairwindsOps/rbac-manager/

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -a -o rbac-manager ./cmd/manager/main.go

FROM alpine:3.10 AS alpine
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates


FROM scratch
COPY --from=alpine /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=alpine /etc/passwd /etc/passwd


USER nobody
COPY --from=build-env /go/src/github.com/FairwindsOps/rbac-manager/rbac-manager /

ENTRYPOINT ["/rbac-manager"]
CMD ["--log-level=info"]
