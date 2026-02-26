APP_NAME    := sc
VERSION     := 0.1.0
LDFLAGS     := -s -w -X main.version=$(VERSION)

build:
	@mkdir -p build
	CGO_ENABLED=0 GOOS=darwin go build -ldflags "$(LDFLAGS)" -o build/$(APP_NAME) .

install:
	install -m 755 build/$(APP_NAME) /usr/local/bin/$(APP_NAME)

restart: install
	sudo launchctl kickstart -k system/com.sc.daemon

clean:
	rm -rf build
