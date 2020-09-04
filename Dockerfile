FROM alpine:3.8
LABEL maintainer="jan.knipper@sap.com"
LABEL source_repository="https://github.com/sapcc/kubernetes-eventexporter"

RUN apk --no-cache add ca-certificates
COPY kubernetes-eventexporter /kubernetes-eventexporter 
USER nobody:nobody

ENTRYPOINT ["/kubernetes-eventexporter"]
CMD ["-logtostderr"]
