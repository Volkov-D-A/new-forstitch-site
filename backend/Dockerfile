FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=builder --chown=app:app /out/api /usr/local/bin/api

USER app

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/api"]
