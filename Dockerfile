FROM golang:1.22 AS builder

WORKDIR /go/src/github.com/sapcc/kuberntes-eventexporter
ADD go.mod go.sum ./
RUN go mod download
ADD . .
RUN go test -v .
RUN CGO_ENABLED=0 go build -v -o /kubernetes-eventexporter

RUN apt update -qqq && \
    apt install -yqqq ca-certificates && \
    update-ca-certificates

FROM gcr.io/distroless/static-debian12
LABEL maintainer="jan.knipper@sap.com"
LABEL source_repository="https://github.com/sapcc/kubernetes-eventexporter"

COPY --from=builder /kubernetes-eventexporter /kubernetes-eventexporter
COPY --from=builder /etc/ssl/certs /etc/ssl/certs

ENTRYPOINT ["/kubernetes-eventexporter"]
CMD ["-logtostderr"]
