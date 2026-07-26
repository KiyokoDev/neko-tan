BINARY=neko-tan
PREFIX?=/usr/local
GOBUILD=go build -ldflags="-s -w"
SUDO?=$(shell command -v sudo 2>/dev/null && echo "sudo" || echo "")

.PHONY: all build install uninstall clean deps

all: build

build:
	$(GOBUILD) -o $(BINARY) .

deps:
	@if command -v yt-dlp >/dev/null 2>&1; then \
		echo "yt-dlp already installed"; \
	else \
		echo "Installing yt-dlp..."; \
		if command -v pacman >/dev/null 2>&1; then \
			$(SUDO) pacman -S --noconfirm yt-dlp; \
		elif command -v apt >/dev/null 2>&1; then \
			$(SUDO) apt update && $(SUDO) apt install -y yt-dlp; \
		elif command -v dnf >/dev/null 2>&1; then \
			$(SUDO) dnf install -y yt-dlp; \
		elif command -v zypper >/dev/null 2>&1; then \
			$(SUDO) zypper install -y yt-dlp; \
		elif command -v apk >/dev/null 2>&1; then \
			$(SUDO) apk add yt-dlp; \
		elif command -v pip3 >/dev/null 2>&1; then \
			pip3 install --user yt-dlp; \
		elif command -v pip >/dev/null 2>&1; then \
			pip install --user yt-dlp; \
		else \
			echo "ERROR: No known package manager found. Install yt-dlp manually:"; \
			echo "  https://github.com/yt-dlp/yt-dlp#installation"; \
			exit 1; \
		fi; \
	fi

install: deps build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 $(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	@echo "Installed $(BINARY) to $(DESTDIR)$(PREFIX)/bin/$(BINARY)"

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY)
	@echo "Removed $(DESTDIR)$(PREFIX)/bin/$(BINARY)"

clean:
	rm -f $(BINARY)
