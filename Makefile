APP_NAME    := sc
VERSION     := 0.1.0
LDFLAGS     := -s -w -X main.version=$(VERSION)
BIN_DIR     := /usr/local/bin
BIN_PATH    := $(BIN_DIR)/$(APP_NAME)

.PHONY: build install uninstall restart clean

build:
	@mkdir -p build
	CGO_ENABLED=0 GOOS=darwin go build -ldflags "$(LDFLAGS)" -o build/$(APP_NAME) .

install:
	install -m 755 build/$(APP_NAME) $(BIN_PATH)

uninstall:
	@echo "Removing $(APP_NAME)..."
	-[ -x $(BIN_PATH) ] && sudo $(BIN_PATH) uninstall 2>/dev/null || true
	sudo rm -f $(BIN_PATH)
	@echo "Done."

restart: install
	sudo launchctl kickstart -k system/com.sc.daemon

clean:
	rm -rf build
