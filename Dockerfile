# ================================
# Build image
# ================================
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

WORKDIR /build

# Download modules first so this layer stays cached until go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Templates and static files are embedded, so the binary is the whole app.
# Go cross-compiles natively, so the build runs on the host arch.
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /server .

# ================================
# Run image
# ================================
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
