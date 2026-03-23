FROM golang:1.25.5-alpine AS build

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY go-api/go.mod go-api/go.sum ./
RUN go mod download

COPY go-api/ ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/go-api ./main.go

FROM alpine:3.22 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app && mkdir -p /app/upload && chown -R app:app /app

ENV PORT=8080
ENV UPLOAD_DIR=/app/upload

COPY --from=build /out/go-api /usr/local/bin/go-api

VOLUME ["/app/upload"]

USER app

EXPOSE 8080

CMD ["go-api"]
