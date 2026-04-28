.PHONY: all frontend backend run dev clean upgrade

all: frontend backend

frontend:
	cd frontend && npm run build

backend:
	cd backend && go build -o server .

run: all
	cd backend && ./server

dev:
	cd frontend && npm run dev

clean:
	rm -rf backend/server backend/public/app

upgrade:
	cd frontend && npm upgrade
	cd backend && go get -u ./... && go mod tidy
