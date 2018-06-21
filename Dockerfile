FROM golang:1.10.3 AS build-env
WORKDIR /go/src/github.com/reactiveops/rbac-manager/

RUN go get -u github.com/golang/dep/...

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o controller-manager ./cmd/controller-manager/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build-env /go/src/github.com/reactiveops/rbac-manager/controller-manager .

ENTRYPOINT ["./controller-manager"]
CMD ["--install-crds=false"]
