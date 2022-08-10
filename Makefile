hello:
	echo "Hello"

build:
	go build -o dist/main main.go

test:
	go test ./... -v

compile:
	echo "Compiling for every OS and Platform"
	GOOS=freebsd GOARCH=386 go build -o dist/main-freebsd-386 main.go
	GOOS=linux GOARCH=386 go build -o dist/main-linux-386 main.go
	GOOS=windows GOARCH=386 go build -o dist/main-windows-386 main.go