## L2Radar demo Makefile
#
# Targets:
#   all / record   Full pipeline: build UI, install deps, record, convert to GIF.
#   gif            Convert an existing demo/output.webm → assets/demo.gif.
#   simulate       Run the host simulator standalone (writes to /tmp/l2radar-demo).
#   serve          Run the demo HTTP server standalone (reads from /tmp/l2radar-demo).
#   clean          Remove generated files (output.webm, *.webm leftovers).
#
# Prerequisites for 'record' / 'gif':
#   ffmpeg and gifsicle must be installed on the host.
#   On Debian/Ubuntu:  sudo apt-get install -y ffmpeg gifsicle
#   On macOS:          brew install ffmpeg gifsicle
#
# Usage:
#   cd demo && make               # full pipeline
#   cd demo && make simulate      # run sim only (Ctrl-C to stop)
#   cd demo && make serve         # run server only (Ctrl-C to stop)

DEMO_DIR   := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
ROOT_DIR   := $(abspath $(DEMO_DIR)/..)
UI_DIR     := $(ROOT_DIR)/ui
ASSETS_DIR := $(ROOT_DIR)/assets

WEBM       := $(DEMO_DIR)output.webm
GIF        := $(ASSETS_DIR)/demo.gif

GIF_SCALE  := 900
GIF_FPS    := 10
GIF_COLORS := 128

.PHONY: all record gif simulate serve clean

all: record

## ── Full recording pipeline ──────────────────────────────────────────────────

record: _build-ui _install-deps _playwright-install _record gif

_build-ui:
	@echo "==> Building UI..."
	cd $(UI_DIR) && npm ci && npm run build

_install-deps:
	@echo "==> Installing demo dependencies..."
	cd $(DEMO_DIR) && npm install

_playwright-install:
	@echo "==> Installing Playwright browser..."
	cd $(DEMO_DIR) && npx playwright install --with-deps chromium

_record:
	@echo "==> Recording demo..."
	cd $(DEMO_DIR) && node record.js

## ── GIF conversion (can be run standalone after _record) ─────────────────────

gif: $(WEBM)
	@echo "==> Converting to GIF..."
	ffmpeg -y -i $(WEBM) \
	  -vf "fps=$(GIF_FPS),scale=$(GIF_SCALE):-1:flags=lanczos,split[s0][s1];[s0]palettegen=max_colors=$(GIF_COLORS)[p];[s1][p]paletteuse=dither=bayer:bayer_scale=5" \
	  -loop 0 \
	  $(GIF)
	gifsicle -O3 --lossy=80 --colors $(GIF_COLORS) -o $(GIF) $(GIF)
	@echo "==> GIF written to $(GIF)"

$(WEBM):
	$(MAKE) _build-ui _install-deps _playwright-install _record

## ── Standalone helpers ────────────────────────────────────────────────────────

simulate: _install-deps
	@echo "==> Running simulator (Ctrl-C to stop)..."
	cd $(DEMO_DIR) && node simulate.js /tmp/l2radar-demo-data

serve: _install-deps _build-ui
	@echo "==> Running demo server on http://localhost:3737 (Ctrl-C to stop)..."
	@echo "    Data dir: /tmp/l2radar-demo-data"
	cd $(DEMO_DIR) && node server.js /tmp/l2radar-demo-data 3737

## ── Cleanup ───────────────────────────────────────────────────────────────────

clean:
	@echo "==> Cleaning generated files..."
	rm -f $(DEMO_DIR)*.webm
	@echo "    (assets/demo.gif is kept — remove manually if needed)"
