APP_NAME ?= readwillbe

.PHONY: tailwind-dev
tailwind-dev:
	tailwindcss -i ./input.css -o ./static/css/style.css

.PHONY: tailwind-build
tailwind-build:
	tailwindcss -i ./input.css -o ./static/css/style.min.css --minify

.PHONY: templ-generate
templ-generate:
	templ generate

.PHONY: templ-watch
templ-watch:
	templ generate --watch

.PHONY: go-build
go-build:
	go build -o ./tmp/$(APP_NAME) ./cmd/$(APP_NAME)/

.PHONY: dev-build
dev-build: templ-generate tailwind-dev go-build

clean:
	rm -f docker-run

docker-run:
	echo "docker run --env-file .env --name ${APP_NAME} --rm -v $$(pwd)/tmp/:/data/ --env DB_PATH=\"/data/${APP_NAME}.db\" -i -p 8080:8080 ${APP_NAME}:latest" > ./docker-run
	chmod +x ./docker-run

.PHONY: docker-build
docker-build: templ-generate docker-run
	docker kill ${APP_NAME} || true
	docker build --build-arg VERSION="$$(git rev-parse --short HEAD)" -t ${APP_NAME}:latest .

.PHONY: dev
dev: docker-build
	air

.PHONY: build
build: templ-generate tailwind-build
	go build -ldflags "-X readwillbe/version.Tag=$$(git rev-parse --short HEAD)" -o ./bin/$(APP_NAME) ./cmd/$(APP_NAME)/

.PHONY: vet
vet:
	go vet ./...

.PHONY: staticcheck
staticcheck:
	staticcheck ./...

.PHONY: test
test:
	go test -race -v -timeout 30s ./...

.PHONY: run
run: dev-build
	./tmp/$(APP_NAME)

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: install-tools
install-tools:
	go install github.com/a-h/templ/cmd/templ@latest
