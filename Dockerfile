# syntax=docker/dockerfile:1

# ─── Stage 1: Build ───────────────────────────────────────────────────────────
# Use the official Go image as the build environment. This stage is discarded
# after compilation — it never appears in the final image.
FROM golang:1.25 AS builder

WORKDIR /build

# Copy dependency manifests first so Docker can cache the module download layer.
# This layer is only invalidated when go.mod or go.sum change, not on every
# source change — keeps incremental builds fast.
COPY src/go.mod src/go.sum ./
RUN go mod download

# Copy source and compile. Flags explained:
#   CGO_ENABLED=0  — fully static binary; no dynamic libc linkage (required for distroless)
#   GOOS=linux     — target OS regardless of where Docker is running
#   -trimpath      — remove local file system paths from the binary (security + reproducibility)
#   -ldflags
#     -s -w              remove debug symbol table and DWARF info (~30% smaller binary)
#     -X main.version    inject the VERSION build-arg into the package-level var at link time
ARG VERSION=dev
COPY src/ ./
RUN CGO_ENABLED=0 GOOS=linux go build \
      -trimpath \
      -ldflags="-s -w -X main.version=${VERSION}" \
      -o /k8s-platform \
      .

# ─── Stage 2: Runtime ─────────────────────────────────────────────────────────
# distroless/static contains only CA certificates and timezone data.
# No shell, no package manager, no libc — minimal attack surface.
# The :nonroot tag sets USER to 65534 by default; we make it explicit below.
FROM gcr.io/distroless/static:nonroot

# Copy only the compiled binary from the builder stage.
COPY --from=builder /k8s-platform /k8s-platform

# Run as UID 65534 (nonroot). Never run application containers as root.
USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/k8s-platform"]
