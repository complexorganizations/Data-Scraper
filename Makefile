run:
	go run main.go

build:
	go build main.go

compile:
    # 32-Bit Systems
	# FreeBDS
	GOOS=freebsd GOARCH=386 go build main.go
	# MacOS
	GOOS=darwin GOARCH=386 go build main.go
	# Linux
	GOOS=linux GOARCH=386 go build main.go
	# Windows
	GOOS=windows GOARCH=386 go build main.go
    # 64-Bit Systems
	# FreeBDS
	GOOS=freebsd GOARCH=amd64 go build main.go
	# MacOS
	GOOS=darwin GOARCH=amd64 go build  main.go
	# Linux
	GOOS=linux GOARCH=amd64 go build main.go
	# Windows
	GOOS=windows GOARCH=amd64 go build main.go
