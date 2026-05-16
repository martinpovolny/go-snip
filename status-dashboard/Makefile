# wg-status Makefile
# Default target: build for the local machine
# 'make linux' cross-compiles for Oracle Cloud Ampere (linux/arm64)

BINARY  := status-dashboard
LDFLAGS := -ldflags="-s -w"

.PHONY: build linux-amd64 linux-arm64 install deploy clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

# Oracle Cloud free tier uses Ampere ARM64
linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .

# Intel/AMD Oracle shape
linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .

# first-time install: copies binary + service file and enables the unit
install: linux-amd64
	ssh ora-m 'sudo mkdir -p /opt/status-dashboard'
	scp $(BINARY)-linux-amd64 ora-m:/tmp/$(BINARY)
	ssh ora-m 'sudo mv /tmp/$(BINARY) /opt/status-dashboard/$(BINARY) && sudo chmod +x /opt/status-dashboard/$(BINARY)'
	scp $(BINARY).service ora-m:/tmp/$(BINARY).service
	ssh ora-m 'sudo mv /tmp/$(BINARY).service /etc/systemd/system/$(BINARY).service && sudo systemctl daemon-reload && sudo systemctl enable --now $(BINARY)'

# scp binary + restart
deploy: linux-amd64
	scp $(BINARY)-linux-amd64 ora-m:/tmp/$(BINARY)
	ssh ora-m 'sudo mv /tmp/$(BINARY) /opt/status-dashboard/$(BINARY) && sudo chmod +x /opt/status-dashboard/$(BINARY) && sudo systemctl restart $(BINARY)'

clean:
	rm -f $(BINARY) $(BINARY)-linux-arm64 $(BINARY)-linux-amd64

