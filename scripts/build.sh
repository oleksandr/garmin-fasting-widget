#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# 1. Detect Connect IQ SDK
SDK_DIR=$(find "$HOME/.Garmin/ConnectIQ/Sdks" -maxdepth 1 -type d -name "connectiq-sdk-*" | sort -V | tail -n 1)

if [ -z "$SDK_DIR" ] || [ ! -f "$SDK_DIR/bin/monkeyc" ]; then
    echo "Error: Connect IQ SDK not found in ~/.Garmin/ConnectIQ/Sdks/"
    echo "Please download an SDK via: connect-iq-sdk-manager sdk download 9.2.0"
    exit 1
fi

MONKEYC="$SDK_DIR/bin/monkeyc"
DEVICE="instinct3amoled50mm"
KEY="developer_key.der"
OUTPUT="bin/FastingWidget.prg"

echo "=== Fasting Widget Build System ==="
echo "Using SDK: $SDK_DIR"
echo "Target Device: $DEVICE"

# 2. Check Developer Key
if [ ! -f "$KEY" ]; then
    echo "Developer key not found. Generating developer_key.der..."
    openssl genrsa -out developer_key.pem 4096
    openssl pkcs8 -topk8 -inform PEM -outform DER -in developer_key.pem -out developer_key.der -nocrypt
fi

# 3. Check Device Definition
DEVICE_DIR="$HOME/.Garmin/ConnectIQ/Devices/$DEVICE"
if [ ! -d "$DEVICE_DIR" ]; then
    echo ""
    echo "=========================================================================="
    echo "NOTICE: Device definition for '$DEVICE' is not yet downloaded."
    echo "Garmin requires an authenticated developer login to fetch device definitions."
    echo ""
    echo "To download your device definition, run:"
    echo "  1) connect-iq-sdk-manager login"
    echo "  2) connect-iq-sdk-manager device download -d $DEVICE"
    echo "=========================================================================="
    echo ""
    exit 1
fi

# 4. Compile Application
mkdir -p bin
echo "Compiling $OUTPUT for $DEVICE..."

"$MONKEYC" \
    -o "$OUTPUT" \
    -f monkey.jungle \
    -y "$KEY" \
    -d "$DEVICE" \
    -r

echo ""
echo "BUILD SUCCESSFUL!"
echo "Generated binary: $PROJECT_DIR/$OUTPUT"
echo ""
echo "To sideload to your Garmin Instinct 3 AMOLED:"
echo "1. Connect watch via USB cable."
echo "2. Copy '$OUTPUT' into the 'GARMIN/APPS/' directory on your watch."
echo "3. Safely disconnect."
