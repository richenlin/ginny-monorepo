FROM alpine:latest

RUN apk update && apk upgrade

# Time Zone
RUN apk add --no-cache bash tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

ADD ./dist/<appName>-linux-amd64 app
ADD ./configs/pro.yml .
ADD ./scripts/wait-for .