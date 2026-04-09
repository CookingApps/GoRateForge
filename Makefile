.PHONY: run build clean

run:
	go run cmd/gorateforge/main.go

build:
	go build -o bin/gorateforge cmd/gorateforge/main.go

clean:
	rm -rf bin/