#!/bin/sh
# RouterPilot install script for OpenWrt
# Usage: sh <(wget -O - https://raw.githubusercontent.com/shagston/routerpilot/main/install.sh)
#
# This script downloads and installs the RouterPilot daemon and LuCI app
# on an OpenWrt system. It requires an active internet connection.

set -e

REPO="shagston/routerpilot"
BRANCH="${BRANCH:-main}"
BASE_URL="https://raw.githubusercontent.com/$REPO/$BRANCH"
RELEASE_URL="https://github.com/$REPO/releases/latest/download"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { printf "${GREEN}%s${NC}\n" "$1"; }
warn()  { printf "${YELLOW}WARNING: %s${NC}\n" "$1"; }
err()   { printf "${RED}ERROR: %s${NC}\n" "$1"; }

detect_arch() {
	case "$(uname -m)" in
		armv5l)    echo "linux-arm5" ;;
		armv6l)    echo "linux-arm6" ;;
		armv7l|armv8l) echo "linux-arm7" ;;
		aarch64|arm64) echo "linux-arm64" ;;
		mips)      echo "linux-mips" ;;
		mipsel)    echo "linux-mipsle" ;;
		mips64)    echo "linux-mips64" ;;
		x86_64|amd64) echo "linux-amd64" ;;
		i386|i486|i586|i686) echo "linux-386" ;;
		*)         echo "unknown" ;;
	esac
}

download() {
	local url="$1"
	local out="$2"
	local desc="${3:-file}"

	printf "  Downloading %s... " "$desc"
	if command -v wget >/dev/null 2>&1; then
		wget -q -O "$out" "$url" && info "OK" && return 0
	elif command -v curl >/dev/null 2>&1; then
		curl -sL -o "$out" "$url" && info "OK" && return 0
	else
		err "neither wget nor curl found"
		return 1
	fi
	err "failed"
	return 1
}

install_binary() {
	local arch
	arch=$(detect_arch)

	if [ "$arch" = "unknown" ]; then
		warn "Unknown architecture: $(uname -m). Skipping binary download."
		return 1
	fi

	local binary_name="routerpilot-${arch}"
	local gz_name="${binary_name}.gz"
	local url="$RELEASE_URL/$gz_name"

	info "Detected architecture: $arch"
	info "Downloading RouterPilot binary..."

	if download "$url" "$TMPDIR/$gz_name" "RouterPilot binary ($arch)"; then
		gzip -d "$TMPDIR/$gz_name"
		cp "$TMPDIR/$binary_name" /usr/bin/routerpilot
		chmod 0755 /usr/bin/routerpilot
		info "Binary installed: /usr/bin/routerpilot"
		return 0
	fi

	warn "Binary not found in release assets."
	warn "No GitHub release found yet, or binary not available for this arch."
	warn ""
	warn "To build from source:"
	warn "  1. Install Go on your build machine"
	warn "  2. git clone https://github.com/$REPO"
	warn "  3. cd routerpilot"
	warn "  4. GOOS=linux GOARCH=$(detect_arch | sed 's/linux-//') go build -o routerpilot ./cmd/routerpilot"
	warn "  5. scp routerpilot root@<router>:/usr/bin/"
	return 1
}

install_lua_runtime() {
	if command -v lua >/dev/null 2>&1; then
		info "Lua runtime already installed."
		return 0
	fi

	for pmgr in opkg apk; do
		if command -v "$pmgr" >/dev/null 2>&1; then
			info "Lua runtime not found. Attempting to install via $pmgr..."
			case "$pmgr" in
				opkg) $pmgr update 2>/dev/null || true; $pmgr install lua 2>/dev/null ;;
				apk)  $pmgr add lua 2>/dev/null ;;
			esac
			if command -v lua >/dev/null 2>&1; then
				info "Lua runtime installed."
				return 0
			fi
			warn "Could not install Lua via $pmgr."
		fi
	done

	warn "No package manager (opkg/apk) found or Lua unavailable."
	warn ""
	warn "LuCI will use JS-only mode (menu.d). Lua-based features (CBI settings) will be unavailable."
	warn "The web UI on port 8080 will still work."
	return 1
}

install_luci() {
	info "Installing LuCI app..."

	local failed=0

	local has_lua
	if command -v lua >/dev/null 2>&1; then
		has_lua=1
	fi

	# JS-only files (always installed — menu.d, views, init, config)
	local js_files="
		htdocs/luci-static/resources/view/routerpilot/dashboard.js:/www/luci-static/resources/view/routerpilot/dashboard.js
		htdocs/luci-static/resources/view/routerpilot/execute.js:/www/luci-static/resources/view/routerpilot/execute.js
		htdocs/luci-static/resources/view/routerpilot/chat.js:/www/luci-static/resources/view/routerpilot/chat.js
		htdocs/luci-static/resources/view/routerpilot/settings.js:/www/luci-static/resources/view/routerpilot/settings.js
		root/usr/share/luci/menu.d/luci-app-routerpilot.json:/usr/share/luci/menu.d/luci-app-routerpilot.json
		root/etc/init.d/routerpilot:/etc/init.d/routerpilot
		root/etc/uci-defaults/40_luci-routerpilot:/etc/uci-defaults/40_luci-routerpilot
	"

	for entry in $js_files; do
		local src="${entry%%:*}"
		local dst="${entry##*:}"
		mkdir -p "$(dirname "$dst")"
		local url="$BASE_URL/luci-app-routerpilot/$src"
		if ! download "$url" "$dst" "$(basename "$dst")"; then
			err "Failed to download $src"
			failed=1
		fi
	done

	# Lua files (optional — only if Lua runtime is available)
	if [ -n "$has_lua" ]; then
		info "Lua runtime found. Installing Lua controller + CBI + templates..."

		local lua_files="
			luasrc/controller/routerpilot.lua:/usr/lib/lua/luci/controller/routerpilot.lua
			luasrc/model/cbi/routerpilot.lua:/usr/lib/lua/luci/model/cbi/routerpilot.lua
			luasrc/view/routerpilot_dashboard.htm:/usr/lib/lua/luci/view/routerpilot_dashboard.htm
			luasrc/view/routerpilot_execute.htm:/usr/lib/lua/luci/view/routerpilot_execute.htm
			luasrc/view/routerpilot_chat.htm:/usr/lib/lua/luci/view/routerpilot_chat.htm
		"

		for entry in $lua_files; do
			local src="${entry%%:*}"
			local dst="${entry##*:}"
			mkdir -p "$(dirname "$dst")"
			local url="$BASE_URL/luci-app-routerpilot/$src"
			if ! download "$url" "$dst" "$(basename "$dst")"; then
				err "Failed to download $src"
				failed=1
			fi
		done
	else
		info "Lua runtime not available — using JS-only mode (menu.d)."
	fi

	for entry in $js_files; do
		local src="${entry%%:*}"
		local dst="${entry##*:}"
		mkdir -p "$(dirname "$dst")"
		local url="$BASE_URL/luci-app-routerpilot/$src"
		if ! download "$url" "$dst" "$(basename "$dst")"; then
			err "Failed to download $src"
			failed=1
		fi
	done

	chmod 0755 /etc/init.d/routerpilot 2>/dev/null || true
	chmod 0755 /etc/uci-defaults/40_luci-routerpilot 2>/dev/null || true

	return $failed
}

configure() {
	info "Configuring RouterPilot..."

	# Create UCI config (defaults overridden by JSON config at runtime)
	uci -q batch <<-EOF 2>/dev/null || true
		del uci.routerpilot 2>/dev/null
		set routerpilot.general=general
		set routerpilot.general.port=8080
		set routerpilot.general.host=0.0.0.0
		set routerpilot.general.log_level=info
		set routerpilot.general.read_only=1
		set routerpilot.general.dry_run=0
		set routerpilot.telegram=telegram
		set routerpilot.telegram.tel_token=""
		set routerpilot.llm=llm
		set routerpilot.llm.llm_type=simple
		set routerpilot.llm.llm_api_key=""
		set routerpilot.llm.llm_endpoint=""
		set routerpilot.llm.llm_model=""
		commit routerpilot
	EOF

	# Run uci-defaults if this is first install
	if [ -f /etc/uci-defaults/40_luci-routerpilot ]; then
		chmod 0755 /etc/uci-defaults/40_luci-routerpilot
		sh /etc/uci-defaults/40_luci-routerpilot 2>/dev/null || true
		# Rename to prevent re-run
		mv /etc/uci-defaults/40_luci-routerpilot /etc/uci-defaults/40_luci-routerpilot.done 2>/dev/null || true
	fi

	info "Configuration complete."
	info "  Edit settings:  uci set routerpilot.settings.read_only=0 && uci commit routerpilot"
	info "  View settings:  uci show routerpilot"
}

show_help() {
	cat <<-EOF

	$(info "RouterPilot installed successfully!")

	Quick start:
	  $(info "/etc/init.d/routerpilot enable")
	  $(info "/etc/init.d/routerpilot start")

	  RouterPilot will be available at:
	    http://$(uci get network.lan.ipaddr 2>/dev/null || echo "192.168.1.1"):8080/

	  LuCI interface:
	    Login to LuCI → Services → RouterPilot

	First run (safe mode, read-only):
	  By default, RouterPilot starts in read-only mode.
	  To disable: uci set routerpilot.settings.read_only=0 && uci commit routerpilot && /etc/init.d/routerpilot reload

	Logs:
	  logread -e routerpilot

	Uninstall:
	  /etc/init.d/routerpilot stop
	  /etc/init.d/routerpilot disable
	  rm -f /usr/bin/routerpilot
	  rm -f /etc/init.d/routerpilot
	  rm -f /usr/lib/lua/luci/controller/routerpilot.lua
	  rm -f /usr/lib/lua/luci/model/cbi/routerpilot.lua
	  rm -f /usr/lib/lua/luci/view/routerpilot_*.htm
	  rm -f /usr/share/luci/menu.d/luci-app-routerpilot.json
	  rm -rf /www/luci-static/resources/view/routerpilot
	  rm -f /etc/uci-defaults/40_luci-routerpilot*
	  uci delete routerpilot 2>/dev/null && uci commit routerpilot

	EOF
}

# ====== MAIN ======

if [ "$(id -u)" -ne 0 ]; then
	err "This script must be run as root (on OpenWrt)."
	exit 1
fi

info "========================================="
info "  RouterPilot Installer for OpenWrt"
info "  Repo: $REPO"
info "========================================="
echo ""

install_binary
echo ""
install_lua_runtime
echo ""
install_luci
echo ""
configure
show_help
