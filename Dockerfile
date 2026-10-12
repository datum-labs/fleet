# Build stage
FROM golang:1.26-alpine AS builder

# Build arguments for version injection
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG GIT_TREE_STATE=unknown
ARG BUILD_DATE=unknown

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.mod
COPY go.sum go.sum

# Cache dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY internal/ internal/

# Build the binary with version information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X 'go.datum.net/fleet-manager/internal/version.Version=${VERSION}' \
              -X 'go.datum.net/fleet-manager/internal/version.GitCommit=${GIT_COMMIT}' \
              -X 'go.datum.net/fleet-manager/internal/version.GitTreeState=${GIT_TREE_STATE}' \
              -X 'go.datum.net/fleet-manager/internal/version.BuildDate=${BUILD_DATE}'" \
    -a -o fleet-manager ./cmd/fleet-manager

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /workspace/fleet-manager .
USER 65532:65532

ENTRYPOINT ["/fleet-manager"]
