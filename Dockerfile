FROM golang:1.17-alpine AS builder
RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache git \
  && go env -w GO111MODULE=auto \
  && go env -w CGO_ENABLED=1 \
  && go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /build

COPY ./ .

RUN set -ex \
    && BUILD=`date +%FT%T%z` \
    && COMMIT_SHA1=`git rev-parse HEAD` \
    && go build -ldflags "-s -w -extldflags '-static' -X main.Version=${COMMIT_SHA1}|${BUILD}"  -o bot_app


FROM alpine AS production

RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

ENV UPDATE "1"
ENV HTTP_PORT "8080"
# 推送解密的密码
ENV APP_ENCRYPT_KEY ""
# APPID
ENV APP_ID ""
# APPSECRET
ENV APP_SECRET ""
ENV ADAPTER_ADDR "bot-adapter:8001"

ENV GITHUB_WEBHOOK_ENABLE "false"
ENV GITHUB_WEBHOOK_SECRET "supersecretcode"
ENV GITHUB_WEBHOOK_NOTIFY_QQ ""
ENV GITHUB_WEBHOOK_NOTIFY_GROUP ""
ENV SELENIUM_CHROME_ENABLE "false"
ENV SELENIUM_CHROME_ADDR "http://127.0.0.1:4444/wd/hub"
ENV SELENIUM_FIREFOX_ENABLE "false"
ENV SELENIUM_FIREFOX_ADDR "http://127.0.0.1:4444/wd/hub"

COPY ./init.sh /
COPY --from=builder /build/bot_app /usr/bin/bot_app
RUN chmod +x /usr/bin/bot_app && chmod +x /init.sh

WORKDIR /data

EXPOSE 8080 80

ENTRYPOINT [ "/init.sh" ]