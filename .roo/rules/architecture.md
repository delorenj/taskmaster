---
description: Describes the high-level architecture of the Task Master CLI application.
globs: scripts/modules/*.js
alwaysApply: false
---
# Application Architecture Overview

- **Modular Structure**: The Task Master CLI is built using a modular architecture, with distinct modules responsible for different aspects of the application. This promotes separation of concerns, maintainability, and testability.

- **Main Modules and Responsibilities**:

  - **[`commands.js`](mdc:scripts/modules/commands.js): Command Handling**
    - **Purpose**: Defines and registers all CLI commands using Commander.js.
    - **Responsibilities** (See also: [`commands.md`](mdc:.roo/rules/commands.md)):
      - Parses command-line arguments and options.
      - Invokes appropriate core logic functions from `scripts/modules/`.
      - Handles user input/output for CLI.
      - Implements CLI-specific validation.

  - **[`task-manager.js`](mdc:scripts/modules/task-manager.js) & `task-manager/` directory: Task Data & Core Logic**
    - **Purpose**: Contains core functions for task data manipulation (CRUD), AI interactions, and related logic.
    - **Responsibilities**:
      - Reading/writing `tasks.json`.
      - Implementing functions for task CRUD, parsing PRDs, expanding tasks, updating status, etc.
      - **Delegating AI interactions** to the `ai-services-unified.js` layer.
      - Accessing non-AI configuration via `config-manager.js` getters.
    - **Key Files**: Individual files within `scripts/modules/task-manager/` handle specific actions (e.g., `add-task.js`, `expand-task.js`).

  - **[`dependency-manager.js`](mdc:scripts/modules/dependency-manager.js): Dependency Management**
    - **Purpose**: Manages task dependencies.
    - **Responsibilities**: Add/remove/validate/fix dependencies.

  - **[`ui.js`](mdc:scripts/modules/ui.js): User Interface Components**
    - **Purpose**: Handles CLI output formatting (tables, colors, boxes, spinners).
    - **Responsibilities**: Displaying tasks, reports, progress, suggestions.

  - **[`ai-services-unified.js`](mdc:scripts/modules/ai-services-unified.js): Unified AI Service Layer**
    - **Purpose**: Centralized interface for all LLM interactions using Vercel AI SDK.
    - **Responsibilities** (See also: [`ai_services.md`](mdc:.roo/rules/ai_services.md)):
      - Exports `generateTextService`, `generateObjectService`.
      - Handles provider/model selection based on `role` and `.taskmasterconfig`.
      - Resolves API keys (from `.env` or `session.env`).
      - Implements fallback and retry logic.
      - Orchestrates calls to provider-specific implementations (`src/ai-providers/`).

  - **[`src/ai-providers/*.js`](mdc:src/ai-providers/): Provider-Specific Implementations**
    - **Purpose**: Provider-specific wrappers for Vercel AI SDK functions.
    - **Responsibilities**: Interact directly with Vercel AI SDK adapters.

  - **[`config-manager.js`](mdc:scripts/modules/config-manager.js): Configuration Management**
    - **Purpose**: Loads, validates, and provides access to configuration.
    - **Responsibilities** (See also: [`utilities.md`](mdc:.roo/rules/utilities.md)):
      - Reads and merges `.taskmasterconfig` with defaults.
      - Provides getters (e.g., `getMainProvider`, `getLogLevel`, `getDefaultSubtasks`) for accessing settings.
      - **Note**: Does **not** store or directly handle API keys (keys are in `.env` or MCP `session.env`).

  - **[`utils.js`](mdc:scripts/modules/utils.js): Core Utility Functions**
    - **Purpose**: Low-level, reusable CLI utilities.
    - **Responsibilities** (See also: [`utilities.md`](mdc:.roo/rules/utilities.md)):
      - Logging (`log` function), File I/O (`readJSON`, `writeJSON`), String utils (`truncate`).
      - Task utils (`findTaskById`), Dependency utils (`findCycles`).
      - API Key Resolution (`resolveEnvVariable`).
      - Silent Mode Control (`enableSilentMode`, `disableSilentMode`).

  - **[`mcp-server/`](mdc:mcp-server/): MCP Server Integration**
    - **Purpose**: Provides MCP interface using FastMCP.
    - **Responsibilities** (See also: [`mcp.md`](mdc:.roo/rules/mcp.md)):
      - Registers tools (`mcp-server/src/tools/*.js`). Tool `execute` methods **should be wrapped** with the `withNormalizedProjectRoot` HOF (from `tools/utils.js`) to ensure consistent path handling.
      - The HOF provides a normalized `args.projectRoot` to the `execute` method.
      - Tool `execute` methods call **direct function wrappers** (`mcp-server/src/core/direct-functions/*.js`), passing the normalized `projectRoot` and other args.
      - Direct functions use path utilities (`mcp-server/src/core/utils/`) to resolve paths based on `projectRoot` from session.
      - Direct functions implement silent mode, logger wrappers, and call core logic functions from `scripts/modules/`.
      - Manages MCP caching and response formatting.

  - **[`init.js`](mdc:scripts/init.js): Project Initialization Logic**
    - **Purpose**: Sets up new Task Master project structure.
    - **Responsibilities**: Creates directories, copies templates, manages `package.json`, sets up `.roo/mcp.json`.

- **Data Flow and Module Dependencies (Updated)**:

  - **CLI**: `bin/task-master.js` -> `scripts/dev.js` (loads `.env`) -> `scripts/modules/commands.js` -> Core Logic (`scripts/modules/*`) -> Unified AI Service (`ai-services-unified.js`) -> Provider Adapters -> LLM API.
  - **MCP**: External Tool -> `mcp-server/server.js` -> Tool (`mcp-server/src/tools/*`) -> Direct Function (`mcp-server/src/core/direct-functions/*`) -> Core Logic (`scripts/modules/*`) -> Unified AI Service (`ai-services-unified.js`) -> Provider Adapters -> LLM API.
  - **Configuration**: Core logic needing non-AI settings calls `config-manager.js` getters (passing `session.env` via `explicitRoot` if from MCP). Unified AI Service internally calls `config-manager.js` getters (using `role`) for AI params and `utils.js` (`resolveEnvVariable` with `session.env`) for API keys.

## Silent Mode Implementation Pattern in MCP Direct Functions

Direct functions (the `*Direct` functions in `mcp-server/src/core/direct-functions/`) need to carefully implement silent mode to prevent console logs from interfering with the structured JSON responses required by MCP. This involves both using `enableSilentMode`/`disableSilentMode` around core function calls AND passing the MCP logger via the standard wrapper pattern (see mcp.md). Here's the standard pattern for correct implementation:

1. **Import Silent Mode Utilities**:
   ```javascript
   import { enableSilentMode, disableSilentMode, isSilentMode } from '../../../../scripts/modules/utils.js';
   ```

2. **Parameter Matching with Core Functions**:
   - ✅ **DO**: Ensure direct function parameters match the core function parameters
   - ✅ **DO**: Check the original core function signature before implementing
   - ❌ **DON'T**: Add parameters to direct functions that don't exist in core functions
   ```javascript
   // Example: Core function signature
   // async function expandTask(tasksPath, taskId, numSubtasks, useResearch, additionalContext, options)
   
   // Direct function implementation - extract only parameters that exist in core
   export async function expandTaskDirect(args, log, context = {}) {
     // Extract parameters that match the core function
     const taskId = parseInt(args.id, 10);
     const numSubtasks = args.num ? parseInt(args.num, 10) : undefined;
     const useResearch = args.research === true;
     const additionalContext = args.prompt || '';
     
     // Later pass these parameters in the correct order to the core function
     const result = await expandTask(
       tasksPath, 
       taskId, 
       numSubtasks, 
       useResearch, 
       additionalContext,
       { mcpLog: log, session: context.session }
     );
   }
   ```

3. **Checking Silent Mode State**:
   - ✅ **DO**: Always use `isSilentMode()` function to check current status
   - ❌ **DON'T**: Directly access the global `silentMode` variable or `global.silentMode`
   ```javascript
   // CORRECT: Use the function to check current state
   if (!isSilentMode()) {
     // Only create a loading indicator if not in silent mode
     loadingIndicator = startLoadingIndicator('Processing...');
   }
   
   // INCORRECT: Don't access global variables directly
   if (!silentMode) { // ❌ WRONG
     loadingIndicator = startLoadingIndicator('Processing...');
   }
   ```

4. **Wrapping Core Function Calls**:
   - ✅ **DO**: Use a try/finally block pattern to ensure silent mode is always restored
   - ✅ **DO**: Enable silent mode before calling core functions that produce console output
   - ✅ **DO**: Disable silent mode in a finally block to ensure it runs even if errors occur
   - ❌ **DON'T**: Enable silent mode without ensuring it gets disabled
   ```javascript
   export async function someDirectFunction(args, log) {
     try {
       // Argument preparation
       const tasksPath = findTasksJsonPath(args, log);
       const someArg = args.someArg;
       
       // Enable silent mode to prevent console logs
       enableSilentMode();
       
       try {
         // Call core function which might produce console output
         const result = await someCoreFunction(tasksPath, someArg);
         
         // Return standardized result object
         return { 
           success: true, 
           data: result, 
           fromCache: false 
         };
       } finally {
         // ALWAYS disable silent mode in finally block
         disableSilentMode();
       }
     } catch (error) {
       // Standard error handling
       log.error(`Error in direct function: ${error.message}`);
       return { 
         success: false, 
         error: { code: 'OPERATION_ERROR', message: error.message }, 
         fromCache: false 
       };
     }
   }
   ```

5. **Mixed Parameter and Global Silent Mode Handling**:
   - For functions that need to handle both a passed `silentMode` parameter and check global state:
   ```javascript
   // Check both the function parameter and global state
   const isSilent = options.silentMode || (typeof options.silentMode === 'undefined' && isSilentMode());
   
   if (!isSilent) {
     console.log('Operation starting...');
   }
   ```

By following these patterns consistently, direct functions will properly manage console output suppression while ensuring that silent mode is always properly reset, even when errors occur. This creates a more robust system that helps prevent unexpected silent mode states that could cause logging problems in subsequent operations.

- **Testing Architecture**:

  - **Test Organization Structure** (See also: [`tests.md`](mdc:.roo/rules/tests.md)):
    - **Unit Tests**: Located in `tests/unit/`, reflect the module structure with one test file per module
    - **Integration Tests**: Located in `tests/integration/`, test interactions between modules
    - **End-to-End Tests**: Located in `tests/e2e/`, test complete workflows from a user perspective
    - **Test Fixtures**: Located in `tests/fixtures/`, provide reusable test data

  - **Module Design for Testability**:
    - **Explicit Dependencies**: Functions accept their dependencies as parameters rather than using globals
    - **Functional Style**: Pure functions with minimal side effects make testing deterministic
    - **Separate Logic from I/O**: Core business logic is separated from file system operations
    - **Clear Module Interfaces**: Each module has well-defined exports that can be mocked in tests
    - **Callback Isolation**: Callbacks are defined as separate functions for easier testing
    - **Stateless Design**: Modules avoid maintaining internal state where possible

  - **Mock Integration Patterns**:
    - **External Libraries**: Libraries like `fs`, `commander`, and `@anthropic-ai/sdk` are mocked at module level
    - **Internal Modules**: Application modules are mocked with appropriate spy functions
    - **Testing Function Callbacks**: Callbacks are extracted from mock call arguments and tested in isolation
    - **UI Elements**: Output functions from `ui.js` are mocked to verify display calls

  - **Testing Flow**:
    - Module dependencies are mocked (following Jest's hoisting behavior)
    - Test modules are imported after mocks are established
    - Spy functions are set up on module methods
    - Tests call the functions under test and verify behavior
    - Mocks are reset between test cases to maintain isolation

- **Benefits of this Architecture**:

  - **Maintainability**: Modules are self-contained and focused, making it easier to understand, modify, and debug specific features.
  - **Testability**:  Each module can be tested in isolation (unit testing), and interactions between modules can be tested (integration testing).
    - **Mocking Support**: The clear dependency boundaries make mocking straightforward
    - **Test Isolation**: Each component can be tested without affecting others
    - **Callback Testing**: Function callbacks can be extracted and tested independently
  - **Reusability**: Utility functions and UI components can be reused across different parts of the application.
  - **Scalability**:  New features can be added as new modules or by extending existing ones without significantly impacting other parts of the application.
  - **Clarity**: The modular structure provides a clear separation of concerns, making the codebase easier to navigate and understand for developers.

This architectural overview should help AI models understand the structure and organization of the Task Master CLI codebase, enabling them to more effectively assist with code generation, modification, and understanding.

## Implementing MCP Support for a Command

Follow these steps to add MCP support for an existing Task Master command (see [`new_features.md`](mdc:.roo/rules/new_features.md) for more detail):

1.  **Ensure Core Logic Exists**: Verify the core functionality is implemented and exported from the relevant module in `scripts/modules/`.

2.  **Create Direct Function File in `mcp-server/src/core/direct-functions/`:**
    - Create a new file (e.g., `your-command.js`) using **kebab-case** naming.
    - Import necessary core functions, **`findTasksJsonPath` from `../utils/path-utils.js`**, and **silent mode utilities**.
    - Implement `async function yourCommandDirect(args, log)` using **camelCase** with `Direct` suffix:
        - **Path Resolution**: Obtain the tasks file path using `const tasksPath = findTasksJsonPath(args, log);`. This relies on `args.projectRoot` being provided.
        - Parse other `args` and perform necessary validation.
        - **Implement Silent Mode**: Wrap core function calls with `enableSilentMode()` and `disableSilentMode()`.
        - Implement caching with `getCachedOrExecute` if applicable.
        - Call core logic.
        - Return `{ success: true/false, data/error, fromCache: boolean }`.
    - Export the wrapper function.

3.  **Update `task-master-core.js` with Import/Export**: Add imports/exports for the new `*Direct` function.

4.  **Create MCP Tool (`mcp-server/src/tools/`)**:
    - Create a new file (e.g., `your-command.js`) using **kebab-case**.
    - Import `zod`, `handleApiResult`, **`getProjectRootFromSession`**, and your `yourCommandDirect` function.
    - Implement `registerYourCommandTool(server)`.
    - **Define parameters, making `projectRoot` optional**: `projectRoot: z.string().optional().describe(...)`.
    - Consider if this operation should run in the background using `AsyncOperationManager`.
    - Implement the standard `execute` method:
      - Get `rootFolder` using `getProjectRootFromSession` (with fallback to `args.projectRoot`).
      - Call `yourCommandDirect({ ...args, projectRoot: rootFolder }, log)` or use `asyncOperationManager.addOperation`.
      - Pass the result to `handleApiResult`.

5.  **Register Tool**: Import and call `registerYourCommandTool` in `mcp-server/src/tools/index.js`.

6.  **Update `mcp.json`**: Add the new tool definition.

## Project Initialization

The `initialize_project` command provides a way to set up a new Task Master project:

- **CLI Command**: `task-master init`
- **MCP Tool**: `initialize_project`
- **Functionality**:
  - Creates necessary directories and files for a new project
  - Sets up `tasks.json` and initial task files
  - Configures project metadata (name, description, version)
  - Handles shell alias creation if requested
  - Works in both interactive and non-interactive modes
  - Creates necessary directories and files for a new project
  - Sets up `tasks.json` and initial task files
  - Configures project metadata (name, description, version)
  - Handles shell alias creation if requested
  - Works in both interactive and non-interactive modes