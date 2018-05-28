FROM golang:1.10.2 AS build-env
WORKDIR /go/src/github.com/reactiveops/rbac-manager/
COPY . .

RUN go get -u github.com/golang/dep/...
RUN dep ensure

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o rbac-manager

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build-env /go/src/github.com/reactiveops/rbac-manager/rbac-manager .
CMD ["./rbac-manager"]
