.PHONY: all build-ui build-local build-linux build-windows build-mac clean

all: build-ui build-local

build-ui:
	@echo "[*] Building React frontend (Single File)..."
	@cd ui && npm install && npm run build
	@mkdir -p pkg/generator/template
	@cp ui/dist/index.html pkg/generator/template/index.html

build-local:
	@echo "[*] Compiling Go binary for native host..."
	go build -ldflags="-w -s" -o git-recon ./cmd/git-recon

build-linux: build-ui
	@echo "[*] Compiling for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/git-recon-linux-amd64 ./cmd/git-recon

build-windows: build-ui
	@echo "[*] Compiling for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o bin/git-recon-windows-amd64.exe ./cmd/git-recon

build-mac: build-ui
	@echo "[*] Compiling for macOS (arm64/Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o bin/git-recon-darwin-arm64 ./cmd/git-recon

clean:
	@echo "[*] Cleaning build artifacts..."
	rm -rf bin/
	rm -f git-recon git-recon.exe
	rm -rf pkg/generator/template/
	rm -rf ui/dist/
