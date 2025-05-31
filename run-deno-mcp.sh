#!/bin/bash

echo "Installing Deno if not already installed..."
if ! command -v deno &> /dev/null; then
  curl -fsSL https://deno.land/install.sh | sh
  export DENO_INSTALL="$HOME/.deno"
  export PATH="$DENO_INSTALL/bin:$PATH"
fi

echo "Running Task Master MCP Server with Deno..."
deno run --allow-read --allow-write --allow-net --allow-env --allow-run --allow-ffi --unstable-ffi --unstable-fs --unstable-kv ./mcp-server/deno-server.ts