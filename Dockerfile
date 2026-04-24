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
FROM rust:1-slim AS backend

WORKDIR /build
COPY backend/Cargo.toml backend/Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release && rm -rf src

COPY backend/src ./src
RUN touch src/main.rs && cargo build --release

# ================================
# Runtime
# ================================
FROM debian:bookworm-slim

RUN adduser --disabled-password --gecos "" --home /app app
WORKDIR /app

COPY --from=backend --chown=app:app /build/target/release/rust_backend ./

COPY --chown=app:app backend/public/ ./public/
COPY --from=frontend --chown=app:app /frontend/dist ./public/app/

USER app
EXPOSE 8080
ENTRYPOINT ["./rust_backend"]
