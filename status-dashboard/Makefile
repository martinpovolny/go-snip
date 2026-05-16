# wg-status Makefile
# Default target: build for the local machine
# 'make linux' cross-compiles for Oracle Cloud Ampere (linux/arm64)

BINARY  := status-dashboard
LDFLAGS := -ldflags="-s -w"

.PHONY: build linux-amd64 deploy clean linux-arm64

build:
	go build $(LDFLAGS) -o $(BINARY) .

# Oracle Cloud free tier uses Ampere ARM64
linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .

# Intel/AMD Oracle shape
linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .

# scp + restart
deploy: linux-arm64
	ssh ora-m 'mkdir -p status-dashboard/bin/'
	scp $(BINARY)-linux-arm64 names.conf ft2:/opt/status-dashboard/$(BINARY)
	ssh ora-m 'sudo systemctl restart status-dashboard'

clean:
	rm -f $(BINARY) $(BINARY)-linux-arm64 $(BINARY)-linux-amd64

