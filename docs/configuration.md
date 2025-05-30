# Configuration

Taskmaster uses two primary methods for configuration:

1.  **`.taskmasterconfig` File (Project Root - Recommended for most settings)**

    - This JSON file stores most configuration settings, including AI model selections, parameters, logging levels, and project defaults.
    - **Location:** This file is created in the root directory of your project when you run the `task-master models --setup` interactive setup. You typically do this during the initialization sequence. Do not manually edit this file beyond adjusting Temperature and Max Tokens depending on your model.
    - **Management:** Use the `task-master models --setup` command (or `models` MCP tool) to interactively create and manage this file. You can also set specific models directly using `task-master models --set-<role>=<model_id>`, adding `--ollama` or `--openrouter` flags for custom models. Manual editing is possible but not recommended unless you understand the structure.
    - **Example Structure:**
      ```json
      {
      	"models": {
      		"main": {
      			"provider": "anthropic",
      			"modelId": "claude-3-7-sonnet-20250219",
      			"maxTokens": 64000,
      			"temperature": 0.2
      		},
      		"research": {
      			"provider": "perplexity",
      			"modelId": "sonar-pro",
      			"maxTokens": 8700,
      			"temperature": 0.1
      		},
      		"fallback": {
      			"provider": "anthropic",
      			"modelId": "claude-3-5-sonnet",
      			"maxTokens": 64000,
      			"temperature": 0.2
      		}
      	},
      	"global": {
      		"logLevel": "info",
      		"debug": false,
      		"defaultSubtasks": 5,
      		"defaultPriority": "medium",
      		"projectName": "Your Project Name",
      		"ollamaBaseUrl": "http://localhost:11434/api",
      		"tasksPath": "./tasks",
      		"azureOpenaiBaseUrl": "https://your-endpoint.openai.azure.com/"
      	}
      }
      ```

### Global Settings in `.taskmasterconfig`

The `global` object within `.taskmasterconfig` contains settings that apply to the overall behavior of Taskmaster:

*   `logLevel` (string): Sets the logging verbosity. Common values: "debug", "info", "warn", "error".
    *   Default: `"info"`
*   `debug` (boolean): Enables or disables debug mode, which may output more verbose logs or error stacks.
    *   Default: `false`
*   `defaultSubtasks` (number): The default number of subtasks to generate when expanding a task if not specified.
    *   Default: `5`
*   `defaultPriority` (string): The default priority for newly created tasks.
    *   Default: `"medium"`
*   `projectName` (string): The name of your project, which may be used in generated outputs or reports.
    *   Default: `"Your Project Name"` (or a name derived during `init`)
*   `ollamaBaseUrl` (string): The base URL for the Ollama API, if you are using Ollama for local model inference.
    *   Default: `"http://localhost:11434/api"`
*   `tasksPath` (string): Specifies the default directory for storing all task-related files, including `tasks.json` and individual task files (e.g., `task_001.txt`).
    *   Default: `"./tasks"`
    *   Example: If set to `"./my_project_tasks"`, then `tasks.json` will be looked for/created at `./my_project_tasks/tasks.json`, and generated task files will go into the `./my_project_tasks/` directory.
    *   Command-line options like `--file` and `--output` can still be used to override this default for specific commands.
*   `azureOpenaiBaseUrl` (string): The base URL for Azure OpenAI services. This is typically your Azure endpoint.
    *   Default: `null` or not present (Example shows `"https://your-endpoint.openai.azure.com/"` if configured)

2.  **Environment Variables (`.env` file or MCP `env` block - For API Keys Only)**
    - Used **exclusively** for sensitive API keys and specific endpoint URLs.
    - **Location:**
      - For CLI usage: Create a `.env` file in your project root.
      - For MCP/Cursor usage: Configure keys in the `env` section of your `.cursor/mcp.json` file.
    - **Required API Keys (Depending on configured providers):**
      - `ANTHROPIC_API_KEY`: Your Anthropic API key.
      - `PERPLEXITY_API_KEY`: Your Perplexity API key.
      - `OPENAI_API_KEY`: Your OpenAI API key.
      - `GOOGLE_API_KEY`: Your Google API key.
      - `MISTRAL_API_KEY`: Your Mistral API key.
      - `AZURE_OPENAI_API_KEY`: Your Azure OpenAI API key (also requires `AZURE_OPENAI_ENDPOINT`).
      - `OPENROUTER_API_KEY`: Your OpenRouter API key.
      - `XAI_API_KEY`: Your X-AI API key.
    - **Optional Endpoint Overrides (in .taskmasterconfig):**
      - `AZURE_OPENAI_ENDPOINT`: Required if using Azure OpenAI key.
      - `OLLAMA_BASE_URL`: Override the default Ollama API URL (Default: `http://localhost:11434/api`).

**Important:** Settings like model ID selections (`main`, `research`, `fallback`), `maxTokens`, `temperature`, `logLevel`, `defaultSubtasks`, `defaultPriority`, and `projectName` are **managed in `.taskmasterconfig`**, not environment variables.

## Example `.env` File (for API Keys)

```
# Required API keys for providers configured in .taskmasterconfig
ANTHROPIC_API_KEY=sk-ant-api03-your-key-here
PERPLEXITY_API_KEY=pplx-your-key-here
# OPENAI_API_KEY=sk-your-key-here
# GOOGLE_API_KEY=AIzaSy...
# etc.

# Optional Endpoint Overrides
# AZURE_OPENAI_ENDPOINT=https://your-azure-endpoint.openai.azure.com/
# OLLAMA_BASE_URL=http://custom-ollama-host:11434/api
```

## Troubleshooting

### Configuration Errors

- If Task Master reports errors about missing configuration or cannot find `.taskmasterconfig`, run `task-master models --setup` in your project root to create or repair the file.
- Ensure API keys are correctly placed in your `.env` file (for CLI) or `.cursor/mcp.json` (for MCP) and are valid for the providers selected in `.taskmasterconfig`.

### If `task-master init` doesn't respond:

Try running it with Node directly:

```bash
node node_modules/claude-task-master/scripts/init.js
```

Or clone the repository and run:

```bash
git clone https://github.com/eyaltoledano/claude-task-master.git
cd claude-task-master
node scripts/init.js
```
