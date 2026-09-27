#!/usr/bin/env bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PREVIEW_FILE="$SCRIPT_DIR/preview.html"

echo "Opening Fasting Widget Layout Preview in your browser..."
if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "$PREVIEW_FILE"
elif command -v firefox >/dev/null 2>&1; then
    firefox "$PREVIEW_FILE" &
elif command -v chromium >/dev/null 2>&1; then
    chromium "$PREVIEW_FILE" &
else
    echo "Please open this file in your browser:"
    echo "  file://$PREVIEW_FILE"
fi
