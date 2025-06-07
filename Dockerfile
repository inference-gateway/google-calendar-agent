FROM golang:1.24-alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
WORKDIR /app
RUN apk add --no-cache upx
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -trimpath \
    -ldflags "-w -s -extldflags '-static' -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
    -o artifacts/agent ./cmd/google-calendar-agent/main.go
RUN upx --best --lzma artifacts/agent

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/artifacts/agent /agent
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/agent"]
