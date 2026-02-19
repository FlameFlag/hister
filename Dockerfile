# syntax=docker/dockerfile:1

FROM node:25-alpine AS web-builder

WORKDIR /build

COPY server/web/package.json server/web/package-lock.json ./server/web/

RUN cd server/web && npm ci --production=false

COPY server/web ./server/web

RUN cd server/web && npm run build

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS go-builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

COPY --from=web-builder /build/server/web/dist ./server/web/dist

ARG TARGETPLATFORM
ARG GOOS
ARG GOARCH
RUN CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build \
    -tags=netgo,osusergo \
    -ldflags="-w -s" \
    -o hister .

FROM alpine:3.21

RUN apk add --no-cache wget ca-certificates

RUN addgroup -g 1000 hister && \
    adduser -D -u 1000 -G hister hister

COPY --from=go-builder --chown=hister:hister /build/hister /hister

RUN chmod +x /hister

USER hister

WORKDIR /data

LABEL maintainer="hister" \
      version="0.9.0" \
      description="Hister - Web history management tool"

EXPOSE 4433

HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=10s \
    CMD wget --quiet --tries=1 --spider http://localhost:4433/ || exit 1

# Set entrypoint (will be overridden by docker-compose)
ENTRYPOINT ["/hister"]
CMD ["listen"]
