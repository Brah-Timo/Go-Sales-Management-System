BINARY_NAME=gestion-commerciale
GO=/usr/local/go/bin/go
BUILD_FLAGS=-ldflags="-s -w"

.PHONY: all build run clean tidy test

all: build

## Build l'application
build:
	export CGO_ENABLED=1 && $(GO) build $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd/app/

## Build pour Windows (cross-compilation)
build-windows:
	CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 \
	  $(GO) build $(BUILD_FLAGS) -o $(BINARY_NAME).exe ./cmd/app/

## Lancer l'application
run: build
	./$(BINARY_NAME)

## Nettoyer
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	rm -rf data/backups/*.zip

## Gestion des dépendances
tidy:
	$(GO) mod tidy

## Tests
test:
	$(GO) test ./... -v

## Vérifier la syntaxe
vet:
	$(GO) vet ./...

## Build avec race detector (développement)
build-dev:
	$(GO) build -race -o $(BINARY_NAME) ./cmd/app/
