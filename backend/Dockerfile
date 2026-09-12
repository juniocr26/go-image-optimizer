# syntax=docker/dockerfile:1

FROM golang:1.27.1-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/go-image-optimizer ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

ENV PORT=8080

COPY --from=build --chown=nonroot:nonroot /out/go-image-optimizer /app/go-image-optimizer

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app/go-image-optimizer"]
