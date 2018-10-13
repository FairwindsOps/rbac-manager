FROM golang:1.11 AS build-env
WORKDIR /go/src/github.com/reactiveops/rbac-manager/
ENV GO111MODULE "on"

COPY . .
RUN go mod download && go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o rbac-manager ./cmd/controller-manager/main.go

FROM alpine:3.8
WORKDIR /usr/local/bin
RUN apk --no-cache add ca-certificates

USER nobody
COPY --from=build-env /go/src/github.com/reactiveops/rbac-manager/rbac-manager .

ENTRYPOINT ["rbac-manager"]
CMD ["--install-crds=false"]
