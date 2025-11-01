#!/bin/bash
# Build for all major platforms script

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}Building WaifuVault for all major platforms...${NC}"
echo "This will take several minutes as it builds for multiple targets."
echo ""

# Check if build script exists
if [ ! -f "./build.sh" ]; then
    echo -e "${RED}ERROR: build.sh not found in current directory${NC}"
    echo "Make sure both build.sh and build-all.sh are in the same folder"
    echo "Also make sure build.sh is executable: chmod +x build.sh"
    exit 1
fi

# Make sure build.sh is executable
chmod +x ./build.sh

# Clean previous builds
rm -rf build
mkdir -p build

start_time=$(date +%s)

echo "================================================"
echo -e "${BLUE}Building Windows 64-bit...${NC}"
echo "================================================"
./build.sh windows amd64
if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: Windows build failed${NC}"
    exit 1
fi

echo ""
echo "================================================"
echo -e "${BLUE}Building Linux 64-bit...${NC}"
echo "================================================"
./build.sh linux amd64
if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: Linux build failed${NC}"
    exit 1
fi

echo ""
echo "================================================"
echo -e "${BLUE}Building macOS Apple Silicon...${NC}"
echo "================================================"
./build.sh darwin arm64
if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: macOS ARM build failed${NC}"
    exit 1
fi

echo ""
echo "================================================"
echo -e "${BLUE}Building macOS Intel...${NC}"
echo "================================================"
./build.sh darwin amd64
if [ $? -ne 0 ]; then
    echo -e "${RED}ERROR: macOS Intel build failed${NC}"
    exit 1
fi

end_time=$(date +%s)
duration=$((end_time - start_time))

echo ""
echo "================================================"
echo -e "${GREEN}ALL BUILDS COMPLETED SUCCESSFULLY!${NC}"
echo "================================================"
echo ""
echo "Built applications:"
ls -la build/
echo ""
echo "You can now distribute these files to users on each platform:"
echo "- waifuvault-windows-amd64.exe  (Windows users)"
echo "- waifuvault-linux-amd64        (Linux users)"
echo "- waifuvault-darwin-arm64.app   (Mac Apple Silicon users)"
echo "- waifuvault-darwin-amd64.app   (Mac Intel users)"
echo ""
echo -e "${GREEN}Total build time: ${duration} seconds${NC}"

# Calculate total size
total_size=$(du -sh build/ | cut -f1)
echo -e "${BLUE}Total size of all builds: ${total_size}${NC}"
