.PHONY: all frontend backend run dev clean upgrade

all: frontend backend

frontend:
	cd frontend && npm run build

backend:
	cd backend && cargo build --release

run: all
	cd backend && ./target/release/rust_backend

dev:
	cd frontend && npm run dev

clean:
	rm -rf backend/target backend/public/app

upgrade:
	cd frontend && npm upgrade
	cd backend && cargo update
