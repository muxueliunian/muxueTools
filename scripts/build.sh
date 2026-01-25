#!/bin/bash
# MuxueTools Build Script (Linux/macOS)
# Usage: ./build.sh [target]
#   target: server | desktop | all | clean

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BIN_DIR="$PROJECT_ROOT/bin"

# Version info (can be overridden by CI, otherwise use git describe)
VERSION="${VERSION:-$(git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "dev")}"
BUILD_TIME="$(date '+%Y-%m-%d %H:%M:%S')"
GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")"

# LDFlags for version injection
LDFLAGS="-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME' -X 'main.GitCommit=$GIT_COMMIT'"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
GRAY='\033[0;90m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${CYAN}========================================${NC}"
    echo -e "${CYAN} $1${NC}"
    echo -e "${CYAN}========================================${NC}\n"
}

ensure_bin_dir() {
    if [ ! -d "$BIN_DIR" ]; then
        mkdir -p "$BIN_DIR"
        echo -e "${GREEN}Created bin directory: $BIN_DIR${NC}"
    fi
}

check_dependencies() {
    local missing=()
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing+=("go")
    fi
    
    # Check Node.js (for frontend build)
    if ! command -v node &> /dev/null; then
        missing+=("node")
    fi
    
    # Check npm
    if ! command -v npm &> /dev/null; then
        missing+=("npm")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        echo -e "${RED}Error: Missing required dependencies: ${missing[*]}${NC}"
        echo -e "${YELLOW}Please install the missing dependencies and try again.${NC}"
        exit 1
    fi
}

build_frontend() {
    print_header "Building Frontend (Vite + Vue)"
    
    local web_dir="$PROJECT_ROOT/web"
    
    if [ ! -d "$web_dir" ]; then
        echo -e "${RED}Error: Web directory not found at $web_dir${NC}"
        return 1
    fi
    
    cd "$web_dir"
    
    # Check if node_modules exists
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}Installing dependencies...${NC}"
        npm ci
        if [ $? -ne 0 ]; then
            echo -e "${RED}npm ci failed!${NC}"
            return 1
        fi
    fi
    
    # Build frontend
    echo -e "${CYAN}Building frontend...${NC}"
    npm run build
    if [ $? -ne 0 ]; then
        echo -e "${RED}Frontend build failed!${NC}"
        return 1
    fi
    
    echo -e "${GREEN}Frontend built successfully: web/dist${NC}"
    cd "$PROJECT_ROOT"
    return 0
}

build_server() {
    print_header "Building Server (Pure Go)"
    
    export CGO_ENABLED=0
    export GOOS=linux
    export GOARCH=amd64
    
    local output_path="$BIN_DIR/muxuetools-server"
    
    cd "$PROJECT_ROOT"
    go build -ldflags "$LDFLAGS" -o "$output_path" ./cmd/server
    
    echo -e "${GREEN}Built: $output_path${NC}"
}

build_desktop() {
    print_header "Building Desktop (CGO + WebView)"
    
    # Check for GTK dependencies
    if ! pkg-config --exists gtk+-3.0 2>/dev/null; then
        echo -e "${RED}Error: GTK3 dependencies not found.${NC}"
        echo -e "${YELLOW}Please install: sudo apt-get install libgtk-3-dev${NC}"
        exit 1
    fi
    
    # Check for WebKit2GTK (try multiple versions)
    local webkit_found=false
    local webkit_version=""
    
    for ver in webkit2gtk-4.1 webkit2gtk-4.0; do
        if pkg-config --exists $ver 2>/dev/null; then
            webkit_found=true
            webkit_version=$ver
            break
        fi
    done
    
    if [ "$webkit_found" = false ]; then
        echo -e "${RED}Error: WebKit2GTK dependencies not found.${NC}"
        echo -e "${YELLOW}Please install: sudo apt-get install libwebkit2gtk-4.1-dev${NC}"
        echo -e "${YELLOW}Or for older systems: sudo apt-get install libwebkit2gtk-4.0-dev${NC}"
        exit 1
    fi
    
    echo -e "${CYAN}Found WebKit: $webkit_version${NC}"
    
    # Ubuntu 24.04+ uses webkit2gtk-4.1, but webview_go hardcodes webkit2gtk-4.0
    # Create a pkg-config wrapper to redirect the request
    if [ "$webkit_version" = "webkit2gtk-4.1" ]; then
        echo -e "${YELLOW}Using webkit2gtk-4.1 (Ubuntu 24.04+)${NC}"
        echo -e "${CYAN}Creating pkg-config wrapper to redirect webkit2gtk-4.0 -> webkit2gtk-4.1${NC}"
        
        # Create temporary directory for wrapper
        local wrapper_dir=$(mktemp -d)
        trap "rm -rf $wrapper_dir" EXIT
        
        # Create pkg-config wrapper script
        cat > "$wrapper_dir/pkg-config" << 'WRAPPER_EOF'
#!/bin/bash
# Wrapper script to redirect webkit2gtk-4.0 requests to webkit2gtk-4.1

# Replace webkit2gtk-4.0 with webkit2gtk-4.1 in arguments
args=()
for arg in "$@"; do
    if [ "$arg" = "webkit2gtk-4.0" ]; then
        args+=("webkit2gtk-4.1")
    else
        args+=("$arg")
    fi
done

# Call the real pkg-config
exec /usr/bin/pkg-config "${args[@]}"
WRAPPER_EOF
        
        chmod +x "$wrapper_dir/pkg-config"
        
        # Put wrapper at the front of PATH
        export PATH="$wrapper_dir:$PATH"
        
        echo -e "${GREEN}pkg-config wrapper created at: $wrapper_dir/pkg-config${NC}"
    fi
    
    export CGO_ENABLED=1
    export GOOS=linux
    export GOARCH=amd64
    
    local output_path="$BIN_DIR/muxuetools-desktop"
    
    cd "$PROJECT_ROOT"
    go build -ldflags "$LDFLAGS" -o "$output_path" ./cmd/desktop
    
    echo -e "${GREEN}Built: $output_path${NC}"
}

clean_build() {
    print_header "Cleaning build artifacts"
    
    if [ -d "$BIN_DIR" ]; then
        rm -rf "$BIN_DIR"
        echo -e "${YELLOW}Removed: $BIN_DIR${NC}"
    fi
    
    # Also clean web/dist if exists
    local dist_dir="$PROJECT_ROOT/web/dist"
    if [ -d "$dist_dir" ]; then
        rm -rf "$dist_dir"
        echo -e "${YELLOW}Removed: $dist_dir${NC}"
    fi
    
    echo -e "${GREEN}Clean complete${NC}"
}

show_usage() {
    echo "MuxueTools Build Script (Linux/macOS)"
    echo ""
    echo "Usage: $0 [target]"
    echo ""
    echo "Targets:"
    echo "  server    Build HTTP server only (no CGO, pure Go)"
    echo "  desktop   Build desktop GUI app (requires GTK3 + WebKit2GTK)"
    echo "  all       Build frontend, server, and desktop (default)"
    echo "  clean     Remove all build artifacts"
    echo ""
    echo "Examples:"
    echo "  $0            # Build all targets"
    echo "  $0 server     # Build server only"
    echo "  $0 clean      # Clean build artifacts"
    echo ""
    echo "Environment Variables:"
    echo "  VERSION       Override version string (default: git tag)"
    echo ""
}

# Main execution
echo -e "${MAGENTA}MuxueTools Build Script${NC}"
echo -e "${GRAY}Version: $VERSION | Commit: $GIT_COMMIT${NC}"

TARGET="${1:-all}"

# Handle help
if [ "$TARGET" = "-h" ] || [ "$TARGET" = "--help" ] || [ "$TARGET" = "help" ]; then
    show_usage
    exit 0
fi

# Check dependencies
check_dependencies

# Ensure bin directory exists
ensure_bin_dir

case "$TARGET" in
    server)
        build_server
        ;;
    desktop)
        # Build frontend first, then desktop
        if build_frontend; then
            build_desktop
        fi
        ;;
    all)
        # Build frontend first, then both server and desktop
        if build_frontend; then
            build_server
            build_desktop
        fi
        ;;
    clean)
        clean_build
        ;;
    *)
        echo -e "${RED}Unknown target: $TARGET${NC}"
        echo ""
        show_usage
        exit 1
        ;;
esac

echo -e "\n${GREEN}Build complete!${NC}"
