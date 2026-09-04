FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl && \
    update-ca-certificates && \
    addgroup -g 1001 teamcity-monitor-user && \
    adduser -u 1001 -D -G teamcity-monitor-user -s /sbin/nologin -g "go service user" teamcity-monitor-user

ADD ./bin/teamcity-monitor /app/bin/
WORKDIR /app

USER teamcity-monitor-user

EXPOSE 8080
ENTRYPOINT [ "/app/bin/teamcity-monitor" ]
CMD ["service"]
