# ================================
# Frontend build
# ================================
FROM node:22-slim AS frontend

WORKDIR /frontend
COPY Frontend/package.json Frontend/package-lock.json ./
RUN npm ci
COPY Frontend/ ./
RUN npm run build -- --outDir dist --emptyOutDir

# ================================
# Backend build
# ================================
FROM golang:1.23-alpine AS backend

WORKDIR /build
COPY Backend/go.mod ./
COPY Backend/main.go ./
RUN CGO_ENABLED=0 go build -o server .

# ================================
# Runtime
# ================================
FROM alpine:3.21

RUN adduser -D -h /app app
WORKDIR /app

# Copy Go binary
COPY --from=backend --chown=app:app /build/server ./

# Copy static assets
COPY --chown=app:app Backend/public/ ./public/

# Copy built frontend
COPY --from=frontend --chown=app:app /frontend/dist ./public/app/

USER app
EXPOSE 8080
ENTRYPOINT ["./server"]
