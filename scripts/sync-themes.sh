#!/bin/bash
set -euo pipefail

NEXTCLOUD_DIR="/Users/smparkin/Nextcloud/iWeb"
THEMES_DIR="$(dirname "$0")/../Public/themes"

for outer in "$NEXTCLOUD_DIR"/*/; do
    name=$(basename "$outer")
    inner=$(find "$outer" -maxdepth 1 -type d ! -path "$outer" | head -1)
    [ -z "$inner" ] && continue

    dest=$(echo "$name" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
    echo "syncing '$name' -> Public/themes/$dest"
    mkdir -p "$THEMES_DIR/$dest"
    rsync -a --delete "$inner/" "$THEMES_DIR/$dest/"
done

echo "done"
