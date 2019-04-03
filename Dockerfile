FROM alpine

EXPOSE 9009

ENTRYPOINT ["/git-sync-static_linux_386"]

RUN apk --no-cache add curl jq

COPY bin/git-sync-static_linux_386 /
