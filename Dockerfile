FROM golang:1.23.4-bookworm as builder
ARG GIT_SHA="<not provided>"

ADD *.go /src/.
ADD reddit /src/reddit
ADD rate_limit /src/rate_limit
ADD go.mod /src/.

RUN cd /src && go mod tidy
RUN cd /src && go build -ldflags "-X ytdlp-site/config.gitSHA=${GIT_SHA} -X ytdlp-site/config.buildDate=$(date +%Y-%m-%d)" -o reimager *.go

FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends --no-install-suggests \
   ca-certificates
 && rm -rf /var/lib/apt/lists/*

COPY --from=0 /usr/local/bin/yt-dlp /usr/local/bin/yt-dlp 
COPY --from=0 /src/server /opt/reimager

WORKDIR /opt
CMD ["/opt/reimager"]
