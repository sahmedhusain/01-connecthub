# This line specifies the Dockerfile syntax version (currently version 1).
# Define a default version (1.18) for the GO_VERSION argument. This can be overridden during build.
ARG GO_VERSION=1.18

# Use the official Golang image with the specified version as the base image for the "build" stage.
# Optionally specify the target platform with --platform=$BUILDPLATFORM.
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build

# Set the working directory for the "build" stage to /src.
WORKDIR /src

# Run the go mod download command with cache and bind mounts:
#   - --mount=type=cache,target=/go/pkg/mod/: Cache Go module dependencies for faster builds.
#   - --mount=type=bind,source=go.sum,target=go.sum: Include the go.sum file for dependency resolution.
#   - --mount=type=bind,source=go.mod,target=go.mod: Include the go.mod file for dependency management.
#   - go mod download -x: Download all dependencies and external dependencies (with -x flag).
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

# Define another argument (TARGETARCH) that needs to be provided during build (specifies target architecture).
ARG TARGETARCH

# Run the go build command with cache and bind mounts:
#   - --mount=type=cache,target=/go/pkg/mod/: Reuse the Go module dependency cache.
#   - --mount=type=bind,target=.: Mount the current directory (source code) into the container.
#   - CGO_ENABLED=0: Disable cgo (calling C code from Go) for a smaller binary.
#   - GOARCH=$TARGETARCH: Set the target architecture based on the TARGETARCH argument.
#   - go build -o /bin/server ./main.go: Build the application from main.go and output it to /bin/server.
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./main.go

# Use the lightweight alpine:latest image as the base for the "final" stage.
FROM alpine:latest AS final

# Update package list and install necessary packages with cache:
#   - --mount=type=cache,target=/var/cache/apk: Cache downloaded packages for faster future installs.
#   - apk --update add ca-certificates tzdata: Install certificates and timezone data.
#   - && update-ca-certificates: Update the system's certificate store after installing certificates.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --update add ca-certificates tzdata \
        && \
        update-ca-certificates

# Define an argument (UID) with a default value (10001) for the user ID. This can be overridden during build.
ARG UID=10001

# Create a user named "appuser" with the specified user ID and limited privileges.
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

# Set the default user for the container to "appuser".
USER appuser

# Copy the built server binary (/bin/server) from the "build" stage to the "final" stage.
COPY --from=build /bin/server /bin/

# Copy the  directory from the build context to the directory in the final image.
COPY . .

# Expose port 8080 for the container to be accessible from the outside.
EXPOSE 8080

# Set the default command to run the server binary when the container starts.
ENTRYPOINT [ "/bin/server" ]