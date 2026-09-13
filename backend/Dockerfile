FROM golang:1.27.1-alpine AS build

WORKDIR /src

RUN apk add --no-cache build-base libheif-dev libheif-libde265 libheif-x265 pkgconf

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/go-image-optimizer ./cmd/api

FROM alpine:3.24

WORKDIR /app

ENV PORT=8080

RUN apk add --no-cache libheif libheif-libde265 libheif-x265 \
	&& addgroup -S optimizer \
	&& adduser -S optimizer -G optimizer

COPY --from=build --chown=optimizer:optimizer /out/go-image-optimizer /app/go-image-optimizer

USER optimizer:optimizer

EXPOSE 8080

ENTRYPOINT ["/app/go-image-optimizer"]
