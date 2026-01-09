#!/bin/bash
# Download llama-server binaries for bundling with scmd
# This runs during goreleaser build
# Compatible with bash 3.2+ (macOS default)

set -e

VERSION="b7688"  # llama.cpp release tag
BASE_URL="https://github.com/ggml-org/llama.cpp/releases/download/${VERSION}"

# Download to .llama-server (hidden directory, not in dist)
# goreleaser will copy from here to the archives
LLAMA_DIR=".llama-server"
mkdir -p "$LLAMA_DIR"

# Platform and file arrays (bash 3.2 compatible)
PLATFORMS=(darwin-amd64 darwin-arm64 linux-amd64 windows-amd64)
FILES=(
    "llama-${VERSION}-bin-macos-x64.tar.gz"
    "llama-${VERSION}-bin-macos-arm64.tar.gz"
    "llama-${VERSION}-bin-ubuntu-x64.tar.gz"
    "llama-${VERSION}-bin-win-cpu-x64.zip"
)

# Download for each platform
for i in "${!PLATFORMS[@]}"; do
    platform="${PLATFORMS[$i]}"
    file="${FILES[$i]}"

    echo "Downloading llama-server for $platform..."

    # Create platform-specific directory
    mkdir -p "$LLAMA_DIR/$platform"

    # Determine file extension
    if [[ "$file" == *.tar.gz ]]; then
        archive_ext="tar.gz"
    else
        archive_ext="zip"
    fi

    # Download archive
    if command -v curl &> /dev/null; then
        curl -fsSL "${BASE_URL}/${file}" -o "$LLAMA_DIR/${platform}.${archive_ext}"
    else
        wget -q "${BASE_URL}/${file}" -O "$LLAMA_DIR/${platform}.${archive_ext}"
    fi

    # Extract based on file type
    if [[ "$archive_ext" == "tar.gz" ]]; then
        tar -xzf "$LLAMA_DIR/${platform}.${archive_ext}" -C "$LLAMA_DIR/$platform"
    else
        unzip -oq "$LLAMA_DIR/${platform}.${archive_ext}" -d "$LLAMA_DIR/$platform"
    fi

    # Find llama-server binary in subdirectories and move to platform root
    if [[ "$platform" == windows-* ]]; then
        # Find .exe file not already in root
        find "$LLAMA_DIR/$platform" -mindepth 2 -name "llama-server.exe" -exec mv {} "$LLAMA_DIR/$platform/" \;
    else
        # Find binary not already in root
        find "$LLAMA_DIR/$platform" -mindepth 2 -name "llama-server" -exec mv {} "$LLAMA_DIR/$platform/" \;
        chmod +x "$LLAMA_DIR/$platform/llama-server"
    fi

    # Clean up extracted subdirectories (keep only llama-server in root)
    find "$LLAMA_DIR/$platform" -mindepth 1 -type d -exec rm -rf {} \; 2>/dev/null || true
    find "$LLAMA_DIR/$platform" -mindepth 1 ! -name "llama-server*" -exec rm -f {} \; 2>/dev/null || true
    rm "$LLAMA_DIR/${platform}.${archive_ext}"

    echo "✓ Downloaded llama-server for $platform"
done

echo ""

# Create darwin-all directory for universal binary (uses arm64 llama-server)
if [ -d "$LLAMA_DIR/darwin-arm64" ]; then
    echo "Creating darwin-all for universal binary..."
    mkdir -p "$LLAMA_DIR/darwin-all"
    cp "$LLAMA_DIR/darwin-arm64/llama-server" "$LLAMA_DIR/darwin-all/llama-server"
    echo "✓ Created darwin-all"
fi

echo "All llama-server binaries downloaded successfully!"
echo "Location: $LLAMA_DIR/"
