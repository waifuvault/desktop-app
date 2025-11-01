#!/bin/bash
# Cross-compilation build script for Fyne apps using fyne-cross

APP_NAME="waifuvault"
VERSION="1.0.0"
APP_ID="com.waifuvault.app"  # Required for Windows/Mac/Mobile builds

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

show_help() {
    echo -e "${BLUE}Fyne Cross-Compilation Build Script${NC}"
    echo "==================================="
    echo ""
    echo "Usage: ./build.sh [GOOS] [GOARCH]"
    echo ""
    echo "Supported platforms:"
    echo "  Windows:  ./build.sh windows amd64"
    echo "  Linux:    ./build.sh linux amd64"
    echo "  Linux:    ./build.sh linux arm64"
    echo "  macOS:    ./build.sh darwin amd64"
    echo "  macOS:    ./build.sh darwin arm64"
    echo "  Android:  ./build.sh android arm64"
    echo "  iOS:      ./build.sh ios arm64"
    echo ""
    echo "Examples:"
    echo "  ./build.sh windows amd64    # Windows 64-bit"
    echo "  ./build.sh linux amd64      # Linux 64-bit"
    echo "  ./build.sh darwin arm64     # Mac Apple Silicon"
    echo "  ./build.sh android arm64    # Android"
    echo ""
    echo "This script uses fyne-cross for reliable cross-compilation"
    echo "of Fyne GUI applications across platforms."
    echo ""
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}ERROR: Docker is required for fyne-cross but not found${NC}"
        echo "Please install Docker first:"
        echo "  - Ubuntu/Debian: sudo apt install docker.io"
        echo "  - Fedora: sudo dnf install docker"
        echo "  - macOS: Install Docker Desktop"
        echo "  - Windows: Install Docker Desktop"
        exit 1
    fi

    # Check if Docker is running
    if ! docker info &> /dev/null; then
        echo -e "${YELLOW}WARNING: Docker is not running${NC}"
        echo "Please start Docker and try again"
        exit 1
    fi
}

install_fyne_cross() {
    echo -e "${YELLOW}Installing fyne-cross...${NC}"
    go install github.com/fyne-io/fyne-cross@latest

    if ! command -v fyne-cross &> /dev/null; then
        echo -e "${RED}ERROR: Failed to install fyne-cross${NC}"
        echo "Make sure your GOPATH/bin is in your PATH"
        echo "Add this to your shell profile:"
        echo "  export PATH=\$PATH:\$(go env GOPATH)/bin"
        exit 1
    fi

    echo -e "${GREEN}fyne-cross installed successfully${NC}"
}

# Main script
if [ $# -lt 2 ] || [ "$1" = "help" ] || [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    show_help
    exit 0
fi

GOOS=$1
GOARCH=$2

echo -e "${BLUE}Building $APP_NAME for $GOOS/$GOARCH...${NC}"
echo ""

# Check for Docker
check_docker

# Check if fyne-cross is installed
if ! command -v fyne-cross &> /dev/null; then
    echo -e "${YELLOW}fyne-cross not found. Installing...${NC}"
    install_fyne_cross
fi

# Create build directory
mkdir -p build

# Build using fyne-cross with app-id
echo -e "${BLUE}Running: fyne-cross $GOOS -arch=$GOARCH -app-version=$VERSION -app-id=$APP_ID${NC}"
fyne-cross $GOOS -arch=$GOARCH -app-version=$VERSION -app-id=$APP_ID

if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: Build failed for $GOOS/$GOARCH${NC}"
    exit 1
fi

# Move output to build directory
echo ""
echo -e "${YELLOW}Moving output to build directory...${NC}"

# First, let's see what files were actually created
echo -e "${BLUE}Files in fyne-cross/dist/$GOOS-$GOARCH/:${NC}"
ls -la "fyne-cross/dist/$GOOS-$GOARCH/" 2>/dev/null || echo "Directory not found"

case $GOOS in
    "windows")
        # Handle both possible naming conventions and file types
        if [ -f "fyne-cross/dist/windows-$GOARCH/Waifuvault.exe.zip" ]; then
            # If it's a zip file, extract it first
            cd "fyne-cross/dist/windows-$GOARCH/"
            unzip -o "Waifuvault.exe.zip"
            cd - > /dev/null
            if [ -f "fyne-cross/dist/windows-$GOARCH/Waifuvault.exe" ]; then
                cp "fyne-cross/dist/windows-$GOARCH/Waifuvault.exe" "build/$APP_NAME-windows-$GOARCH.exe"
                echo -e "${GREEN}SUCCESS: build/$APP_NAME-windows-$GOARCH.exe${NC}"
            fi
        elif [ -f "fyne-cross/dist/windows-$GOARCH/Waifuvault.exe" ]; then
            cp "fyne-cross/dist/windows-$GOARCH/Waifuvault.exe" "build/$APP_NAME-windows-$GOARCH.exe"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-windows-$GOARCH.exe${NC}"
        elif [ -f "fyne-cross/dist/windows-$GOARCH/$APP_NAME.exe" ]; then
            cp "fyne-cross/dist/windows-$GOARCH/$APP_NAME.exe" "build/$APP_NAME-windows-$GOARCH.exe"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-windows-$GOARCH.exe${NC}"
        else
            echo -e "${YELLOW}WARNING: No Windows executable found. Files in directory:${NC}"
            ls -la "fyne-cross/dist/windows-$GOARCH/" 2>/dev/null
        fi
        ;;
    "linux")
        # Handle different possible names
        if [ -f "fyne-cross/dist/linux-$GOARCH/Waifuvault" ]; then
            cp "fyne-cross/dist/linux-$GOARCH/Waifuvault" "build/$APP_NAME-linux-$GOARCH"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-linux-$GOARCH${NC}"
        elif [ -f "fyne-cross/dist/linux-$GOARCH/$APP_NAME" ]; then
            cp "fyne-cross/dist/linux-$GOARCH/$APP_NAME" "build/$APP_NAME-linux-$GOARCH"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-linux-$GOARCH${NC}"
        else
            echo -e "${YELLOW}WARNING: No Linux executable found. Files in directory:${NC}"
            ls -la "fyne-cross/dist/linux-$GOARCH/" 2>/dev/null
        fi
        ;;
    "darwin")
        # Handle different possible names
        if [ -d "fyne-cross/dist/darwin-$GOARCH/Waifuvault.app" ]; then
            cp -r "fyne-cross/dist/darwin-$GOARCH/Waifuvault.app" "build/$APP_NAME-darwin-$GOARCH.app"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-darwin-$GOARCH.app${NC}"
        elif [ -d "fyne-cross/dist/darwin-$GOARCH/$APP_NAME.app" ]; then
            cp -r "fyne-cross/dist/darwin-$GOARCH/$APP_NAME.app" "build/$APP_NAME-darwin-$GOARCH.app"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-darwin-$GOARCH.app${NC}"
        else
            echo -e "${YELLOW}WARNING: No macOS app found. Files in directory:${NC}"
            ls -la "fyne-cross/dist/darwin-$GOARCH/" 2>/dev/null
        fi
        ;;
    "android")
        if [ -f "fyne-cross/dist/android-$GOARCH/Waifuvault.apk" ]; then
            cp "fyne-cross/dist/android-$GOARCH/Waifuvault.apk" "build/$APP_NAME-android-$GOARCH.apk"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-android-$GOARCH.apk${NC}"
        elif [ -f "fyne-cross/dist/android-$GOARCH/$APP_NAME.apk" ]; then
            cp "fyne-cross/dist/android-$GOARCH/$APP_NAME.apk" "build/$APP_NAME-android-$GOARCH.apk"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-android-$GOARCH.apk${NC}"
        else
            echo -e "${YELLOW}WARNING: No Android APK found. Files in directory:${NC}"
            ls -la "fyne-cross/dist/android-$GOARCH/" 2>/dev/null
        fi
        ;;
    "ios")
        if [ -d "fyne-cross/dist/ios-$GOARCH/Waifuvault.app" ]; then
            cp -r "fyne-cross/dist/ios-$GOARCH/Waifuvault.app" "build/$APP_NAME-ios-$GOARCH.app"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-ios-$GOARCH.app${NC}"
        elif [ -d "fyne-cross/dist/ios-$GOARCH/$APP_NAME.app" ]; then
            cp -r "fyne-cross/dist/ios-$GOARCH/$APP_NAME.app" "build/$APP_NAME-ios-$GOARCH.app"
            echo -e "${GREEN}SUCCESS: build/$APP_NAME-ios-$GOARCH.app${NC}"
        else
            echo -e "${YELLOW}WARNING: No iOS app found. Files in directory:${NC}"
            ls -la "fyne-cross/dist/ios-$GOARCH/" 2>/dev/null
        fi
        ;;
    *)
        echo -e "${YELLOW}WARNING: Unknown OS $GOOS, check fyne-cross/dist/ directory manually${NC}"
        ls -la "fyne-cross/dist/" 2>/dev/null
        ;;
esac

echo ""
echo -e "${GREEN}Build complete! Check the build/ directory:${NC}"
ls -la build/
