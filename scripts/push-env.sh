#!/bin/bash

# This script loads environment variables from .env file and exports them to the current shell

# Check if .env file exists
if [ ! -f .env ]; then
  echo "Error: .env file not found"
  exit 1
fi

# Read .env file and export variables
while IFS= read -r line || [[ -n "$line" ]]; do
  # Skip empty lines and comments
  if [[ -z "$line" || "$line" =~ ^# ]]; then
    continue
  fi
  
  # Extract key and value
  if [[ "$line" =~ ^([A-Za-z0-9_]+)=(.*)$ ]]; then
    key="${BASH_REMATCH[1]}"
    value="${BASH_REMATCH[2]}"
    
    # Remove quotes if present
    value="${value%\"}"
    value="${value#\"}"
    value="${value%\'}"
    value="${value#\'}"
    
    # Export the variable
    export "$key=$value"
    echo "Exported $key"
  fi
done < .env

echo "Environment variables loaded successfully!"
