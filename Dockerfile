FROM golang:1.24-alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o artifacts/agent ./cmd/google-calendar-agent/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata curl
WORKDIR /root/
COPY --from=builder /app/artifacts/agent .
EXPOSE 8080
CMD ["./agent"]
