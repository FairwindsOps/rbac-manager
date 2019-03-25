FROM golang:1.12 AS build-env
WORKDIR /go/src/github.com/reactiveops/rbac-manager/

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o rbac-manager ./cmd/manager/main.go

FROM alpine:3.9
WORKDIR /usr/local/bin
RUN apk --no-cache add ca-certificates

USER nobody
COPY --from=build-env /go/src/github.com/reactiveops/rbac-manager/rbac-manager .

ENTRYPOINT ["rbac-manager"]
CMD ["--log-level=info"]
