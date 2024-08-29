FROM alpine:edge
RUN apk add --no-cache ca-certificates tzdata
COPY foxtrot /foxtrot
ENTRYPOINT ["/foxtrot"]