# Build stage
FROM golang:1.25.4 AS builder

WORKDIR /src

# Cache deps first for faster rebuilds.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary.
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ltm-lab .

# Runtime stage
FROM gcr.io/distroless/base-debian12

# Run as non-root for safety.
USER nonroot:nonroot

COPY --from=builder /out/ltm-lab /ltm-lab

EXPOSE 8080

ENTRYPOINT ["/ltm-lab"]
