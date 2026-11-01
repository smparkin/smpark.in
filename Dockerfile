# ================================
# Image resizing
# ================================
# The repo keeps full-size originals. The deployed copies are resized to 2-3x
# their displayed size (a 150px avatar, a ~750px-wide photo) and converted to
# sRGB, so dropping the embedded color profile doesn't shift colors. Each one
# ships as a JPEG plus a smaller JPEG XL for browsers that support it.
FROM --platform=$BUILDPLATFORM alpine:3.22 AS images

RUN apk add --no-cache vips-tools libjxl-tools

COPY site/static/images/ /in/
ARG JPEG="Q=82,keep=none,optimize-coding,interlace"
RUN mkdir /out /resized \
    && vipsthumbnail /in/me.jpg --size 450 --export-profile srgb -o "/resized/%s.png" \
    && vipsthumbnail /in/trees.jpg --size 1600 --export-profile srgb -o "/resized/%s.png" \
    && for f in /resized/*.png; do \
         name=$(basename "$f" .png); \
         vips copy "$f" "/out/$name.jpg[$JPEG]" && cjxl --quiet -q 82 "$f" "/out/$name.jxl" || exit 1; \
       done

# ================================
# Build image
# ================================
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

WORKDIR /build
ENV CGO_ENABLED=0

COPY . .

# A failing check fails the image build, so CI never pushes a broken image
RUN go vet ./... && go test ./...

# Swap in the resized images before they're embedded
COPY --from=images /out/ site/static/images/

# Templates and static files are embedded, so the binary is the whole app.
# Go cross-compiles natively, so the build runs on the host arch.
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /server .

# ================================
# Run image
# ================================
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /server /server

EXPOSE 8080
ENTRYPOINT ["/server"]
