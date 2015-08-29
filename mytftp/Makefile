## simple makefile to log workflow
.PHONY: all test clean buildserver buildclient installserver installclient

GOFLAGS ?= $(GOFLAGS:)

#all: install test
all: buildserver buildclient

buildserver:
	@go build $(GOFLAGS) mytftpserver.go mypkt.go mytftpworker.go\
	 myfile.go mytftpcursor.go

buildclient:
	@go build $(GOFLAGS) mytftpclient.go mypkt.go mytftpworker.go\
	 myfile.go mytftpcursor.go

install:
	@go get $(GOFLAGS) ./...

test: 
	@go test $(GOFLAGS) mytftp_test.go mypkt.go mytftpworker.go myfile.go mytftpcursor.go

bench: install
	@go test -run=NONE -bench=. $(GOFLAGS) ./...

clean:
	@go clean $(GOFLAGS) -i ./...

## EOF
