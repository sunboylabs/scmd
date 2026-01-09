#!/bin/bash
# Download llama-server binaries for bundling with scmd
# This runs during goreleaser build
# Compatible with bash 3.2+ (macOS default)

set -e

VERSION="b4498"  # llama.cpp release tag
BASE_URL="https://github.com/ggerganov/llama.cpp/releases/download/${VERSION}"

mkdir -p dist/llama-server

# Platform and file arrays (bash 3.2 compatible)
PLATFORMS=(darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64)
FILES=(
    "llama-${VERSION}-bin-macos-x64.zip"
    "llama-${VERSION}-bin-macos-arm64.zip"
    "llama-${VERSION}-bin-ubuntu-x64.zip"
    "llama-${VERSION}-bin-ubuntu-aarch64.zip"
    "llama-${VERSION}-bin-win-llvm-x64.zip"
)

# Download for each platform
for i in "${!PLATFORMS[@]}"; do
    platform="${PLATFORMS[$i]}"
    file="${FILES[$i]}"

    echo "Downloading llama-server for $platform..."

    # Create platform-specific directory
    mkdir -p "dist/llama-server/$platform"

    # Download and extract
    if command -v curl &> /dev/null; then
        curl -fsSL "${BASE_URL}/${file}" -o "dist/llama-server/${platform}.zip"
    else
        wget -q "${BASE_URL}/${file}" -O "dist/llama-server/${platform}.zip"
    fi

    # Extract llama-server binary only
    unzip -q "dist/llama-server/${platform}.zip" -d "dist/llama-server/$platform"

    # Find and rename llama-server binary
    if [[ "$platform" == windows-* ]]; then
        find "dist/llama-server/$platform" -name "llama-server.exe" -exec mv {} "dist/llama-server/$platform/" \;
    else
        find "dist/llama-server/$platform" -name "llama-server" -exec mv {} "dist/llama-server/$platform/" \;
        chmod +x "dist/llama-server/$platform/llama-server"
    fi

    # Clean up extracted files (keep only llama-server)
    find "dist/llama-server/$platform" -mindepth 1 -maxdepth 1 ! -name "llama-server*" -exec rm -rf {} \;
    rm "dist/llama-server/${platform}.zip"

    echo "✓ Downloaded llama-server for $platform"
done

echo ""
echo "All llama-server binaries downloaded successfully!"
echo "Location: dist/llama-server/"
