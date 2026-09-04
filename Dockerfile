FROM node:26-alpine AS web

WORKDIR /web

RUN npm install -g pnpm@12

COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY web/ ./
RUN pnpm run build

FROM golang:1.26-alpine AS backend

ENV CGO_ENABLED=0

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
COPY --from=web /web/dist ./internal/webui/dist

RUN go build -trimpath -o /out/teamcity-monitor ./cmd/teamcity-monitor

FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl && \
    update-ca-certificates && \
    addgroup -g 1001 teamcity-monitor-user && \
    adduser -u 1001 -D -G teamcity-monitor-user -s /sbin/nologin -g "go service user" teamcity-monitor-user

COPY --from=backend /out/teamcity-monitor /app/bin/teamcity-monitor
WORKDIR /app

USER teamcity-monitor-user

EXPOSE 8080
ENTRYPOINT [ "/app/bin/teamcity-monitor" ]
CMD ["service"]
