FROM golang:1.26.4-alpine AS build

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY go-api/go.mod go-api/go.sum ./
RUN go mod download

COPY go-api/ ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/go-api ./main.go

FROM alpine:3.24 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -D -H -u 10001 -G app app \
    && mkdir -p /app/upload \
    && chown -R app:app /app

ENV PORT=8080
ENV UPLOAD_DIR=/app/upload

COPY --from=build /out/go-api /usr/local/bin/go-api

VOLUME ["/app/upload"]

USER 10001:10001

EXPOSE 8080
EXPOSE 9090

CMD ["go-api"]
