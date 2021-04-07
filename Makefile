
VERSION="v0.0.1"

all:clean
	mkdir -p build
	GOOS="darwin" GOARCH="amd64" go build -o build/j2-darwin-amd64-${VERSION} cmd/j2/main.go
	GOOS="linux" GOARCH="amd64" go build -o build/j2-linux-amd64-${VERSION} cmd/j2/main.go

clean:
	rm -fr build