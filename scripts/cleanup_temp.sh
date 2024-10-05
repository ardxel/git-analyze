#!/bin/bash

TMPDIR="${TMPDIR:-/tmp}"

find "$TMPDIR" -mindepth 1 -maxdepth 1 -type d -name '*git*' | while IFS= read -r dir; do
	echo "Deleting: $dir"
	rm -rf "$dir"
done

echo "Cleanup complete"
