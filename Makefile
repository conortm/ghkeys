package = github.com/conortm/ghkeys

.PHONY: install release test travis

install:
	go get -t -v ./...

release:
	mkdir -p release
	GOOS=linux GOARCH=amd64 go build -o release/ghkeys-linux-amd64 $(package)
	GOOS=linux GOARCH=386 go build -o release/ghkeys-linux-386 $(package)
	GOOS=linux GOARCH=arm go build -o release/ghkeys-linux-arm $(package)

test:
	go test

travis:
	$(HOME)/gopath/bin/goveralls -service=travis-ci
