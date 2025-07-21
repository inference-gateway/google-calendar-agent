FROM golang:1.24-alpine AS builder
ARG VERSION=dev
ARG COMMIT=unknown
ARG DATE=unknown
ARG AGENT_NAME="Google Calendar Agent"
ARG AGENT_DESCRIPTION="AI agent for Google Calendar operations including listing events, creating events, managing schedules, and finding available time slots"
WORKDIR /app
RUN apk add --no-cache upx
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -trimpath \
    -ldflags "-w -s -extldflags '-static' -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.Date=${DATE} -X 'main.AgentName=${AGENT_NAME}' -X 'main.AgentDescription=${AGENT_DESCRIPTION}'" \
    -o dist/agent ./cmd/agent/main.go
RUN upx --best --lzma dist/agent

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/dist/agent /agent
COPY --from=builder /app/.well-known/agent.json .well-known/agent.json
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/agent"]
