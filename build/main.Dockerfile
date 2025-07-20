FROM golang:alpine AS builder

RUN apk add --no-cache ca-certificates

COPY . /github.com/totorialman/vk-marketplace-api
WORKDIR /github.com/totorialman/vk-marketplace-api

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -o ./.bin ./cmd/main.go

FROM scratch AS runner

WORKDIR /build_v1/

COPY --from=builder /github.com/totorialman/vk-marketplace-api/.bin .

COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /
ENV TZ="Europe/Moscow"
ENV ZONEINFO=/zoneinfo.zip

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT ["./.bin"]
