# status-dashboard Makefile

BINARY  := status-dashboard
LDFLAGS := -ldflags="-s -w"
REMOTE  := ora-m
DEPLOY_DIR := /opt/status-dashboard

.PHONY: build linux-amd64 linux-arm64 deploy install clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

linux-arm64:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY)-linux-arm64 .

linux-amd64:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY)-linux-amd64 .

# Routine update: build, copy binary + config, restart service
deploy: linux-amd64
	scp $(BINARY)-linux-amd64 $(REMOTE):/tmp/$(BINARY)
	scp monitors.conf $(REMOTE):/tmp/monitors.conf
	ssh $(REMOTE) 'sudo mkdir -p $(DEPLOY_DIR) && \
	  sudo mv /tmp/$(BINARY) $(DEPLOY_DIR)/$(BINARY) && \
	  sudo chmod +x $(DEPLOY_DIR)/$(BINARY) && \
	  sudo mv /tmp/monitors.conf $(DEPLOY_DIR)/monitors.conf && \
	  sudo systemctl restart $(BINARY)'

# First-time setup: deploy everything including the systemd unit
install: deploy
	scp $(BINARY).service $(REMOTE):/tmp/$(BINARY).service
	ssh $(REMOTE) 'sudo mv /tmp/$(BINARY).service /etc/systemd/system/$(BINARY).service && \
	  sudo systemctl daemon-reload && \
	  sudo systemctl enable $(BINARY)'

clean:
	rm -f $(BINARY) $(BINARY)-linux-arm64 $(BINARY)-linux-amd64
