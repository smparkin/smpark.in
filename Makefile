.PHONY: all frontend backend run dev clean

all: frontend backend

frontend:
	cd Frontend && npm run build

backend:
	cd Backend && go build -o server .

run: all
	cd Backend && ./server

dev:
	cd Frontend && npm run dev

clean:
	rm -rf Backend/server Backend/public/app
