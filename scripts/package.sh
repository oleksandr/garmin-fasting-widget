#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

# 1. Detect Connect IQ SDK
SDK_DIR=$(find "$HOME/.Garmin/ConnectIQ/Sdks" -maxdepth 1 -type d -name "connectiq-sdk-*" | sort -V | tail -n 1)

if [ -z "$SDK_DIR" ] || [ ! -f "$SDK_DIR/bin/monkeyc" ]; then
    echo "Error: Connect IQ SDK not found in ~/.Garmin/ConnectIQ/Sdks/"
    exit 1
fi

MONKEYC="$SDK_DIR/bin/monkeyc"
KEY="developer_key.der"
OUTPUT="bin/FastingWidget.iq"

echo "=== Fasting Widget Package Builder (.iq) ==="
echo "Using SDK: $SDK_DIR"
echo "Output: $OUTPUT"

mkdir -p bin
"$MONKEYC" -e -o "$OUTPUT" -f monkey.jungle -y "$KEY" -r

echo ""
echo "PACKAGE BUILD SUCCESSFUL!"
echo "Generated package: $PROJECT_DIR/$OUTPUT"
echo ""
echo "To install wirelessly and enable App Settings on iOS:"
echo "1. Go to https://apps.garmin.com/developer/dashboard"
echo "2. Log in with your Garmin account."
echo "3. Click 'Upload an App' -> select 'bin/FastingWidget.iq'."
echo "4. Under 'Visibility / Release Type', select 'Beta' or 'Private'."
echo "5. Open Connect IQ Store on your iPhone -> install wirelessly & configure settings!"
