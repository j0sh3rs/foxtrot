FROM alpine:edge
RUN apk add --no-cache ca-certificates
COPY foxtrot /foxtrot
ENTRYPOINT ["/foxtrot"]