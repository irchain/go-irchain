FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 https://github.com/happyuc-project/happyuc-go && \
  (cd happyuc-go && make ghuc) && \
  cp happyuc-go/build/bin/ghuc /ghuc && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /happyuc-go && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 50505

ENTRYPOINT ["/ghuc"]
