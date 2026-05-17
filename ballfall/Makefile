BINARY     := ballfall
LDFLAGS    := -ldflags="-s -w"
REMOTE     := ora-m
DEPLOY_DIR := /opt/ballfall
TMP        := tmp

.PHONY: build linux-amd64 linux-arm64 deploy install clean status logs

build:
	go build $(LDFLAGS) -o $(BINARY) .

linux-amd64:
	mkdir -p $(TMP)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(TMP)/$(BINARY)-linux-amd64 .

linux-arm64:
	mkdir -p $(TMP)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(TMP)/$(BINARY)-linux-arm64 .

# Routine update: build, copy binary, restart service.
# Web UI is embedded in the binary — no separate static files needed.
deploy: linux-amd64
	scp $(TMP)/$(BINARY)-linux-amd64 $(REMOTE):/tmp/$(BINARY)
	ssh $(REMOTE) 'sudo mkdir -p $(DEPLOY_DIR) && \
	  sudo mv /tmp/$(BINARY) $(DEPLOY_DIR)/$(BINARY) && \
	  sudo chmod +x $(DEPLOY_DIR)/$(BINARY) && \
	  sudo systemctl restart $(BINARY)'

# First-time install: deploy binary + register + enable systemd unit.
install: linux-amd64
	scp $(TMP)/$(BINARY)-linux-amd64 $(REMOTE):/tmp/$(BINARY)
	scp $(BINARY).service             $(REMOTE):/tmp/$(BINARY).service
	ssh $(REMOTE) 'sudo mkdir -p $(DEPLOY_DIR) && \
	  sudo mv /tmp/$(BINARY)         $(DEPLOY_DIR)/$(BINARY) && \
	  sudo chmod +x                  $(DEPLOY_DIR)/$(BINARY) && \
	  sudo mv /tmp/$(BINARY).service /etc/systemd/system/$(BINARY).service && \
	  sudo systemctl daemon-reload && \
	  sudo systemctl enable $(BINARY) && \
	  sudo systemctl start  $(BINARY)'

status:
	ssh $(REMOTE) 'sudo systemctl status $(BINARY)'

logs:
	ssh $(REMOTE) 'sudo journalctl -u $(BINARY) -f'

clean:
	rm -rf $(TMP) $(BINARY)
