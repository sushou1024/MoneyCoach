FROM golang:1.25-alpine AS builder

ARG TARGETARCH
WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETARCH} go build -trimpath -ldflags "-s -w" -o /out/app ./

FROM alpine:3.20 AS runtime
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
