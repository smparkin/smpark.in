# ================================
# Frontend build
# ================================
FROM node:22-slim AS frontend

WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build -- --outDir dist --emptyOutDir

# ================================
# Backend build
# ================================
FROM golang:1.26-alpine AS backend

WORKDIR /build
COPY backend/ ./
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
COPY --chown=app:app backend/public/ ./public/

# Copy built frontend
COPY --from=frontend --chown=app:app /frontend/dist ./public/app/

USER app
EXPOSE 8080
ENTRYPOINT ["./server"]
