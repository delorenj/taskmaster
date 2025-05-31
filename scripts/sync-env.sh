#!/bin/bash

# This script synchronizes environment variables between .env file and system environment

# Check if .env file exists
if [ ! -f .env ]; then
  echo "Creating .env file..."
  touch .env
fi

# Function to update or add environment variable to .env file
update_env_var() {
  local key=$1
  local value=$2
  
  if grep -q "^${key}=" .env; then
    # Update existing variable
    sed -i '' "s|^${key}=.*|${key}=${value}|" .env
  else
    # Add new variable
    echo "${key}=${value}" >> .env
  fi
}

# List of API keys to sync
API_KEYS=(
  "ANTHROPIC_API_KEY"
  "PERPLEXITY_API_KEY"
  "OPENAI_API_KEY"
  "GOOGLE_API_KEY"
  "MISTRAL_API_KEY"
  "OPENROUTER_API_KEY"
  "XAI_API_KEY"
  "AZURE_OPENAI_API_KEY"
)

# Sync environment variables
for key in "${API_KEYS[@]}"; do
  # Get value from environment
  value="${!key}"
  
  # If value exists in environment, update .env file
  if [ -n "$value" ]; then
    update_env_var "$key" "$value"
    echo "Updated $key in .env file"
  else
    echo "Warning: $key not found in environment"
  fi
done

echo "Environment synchronization complete!"
