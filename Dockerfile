FROM golang:1.13-alpine

WORKDIR /go/src/github.com/mattes/gce-deploy-action

COPY . .

RUN go build -mod vendor .

ENTRYPOINT ["/go/src/github.com/mattes/gce-deploy-action/gce-deploy-action"]

