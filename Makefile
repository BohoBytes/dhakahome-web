run:
	@export API_BASE_URL=http://localhost:3000/api/v1; \
	export ADDR=:5173; \
	echo "API_BASE_URL=$$API_BASE_URL"; \
	go run ./cmd/web

css:
	npm run css:dev

build:
	npm run css:build
	GOOS=linux GOARCH=amd64 go build -o bin/server ./cmd/web
