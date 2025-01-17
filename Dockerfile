# Dockerfile
FROM golang:1.23-alpine AS build-stage

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

# Copy the built binary from build-stage
COPY --from=build-stage /app/main /

# Expose port
EXPOSE 5800

USER nonroot:nonroot

ENTRYPOINT ["/main"]