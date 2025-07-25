---
description: Guidelines for integrating new features into the Task Master CLI
globs: scripts/modules/*.js
alwaysApply: false
---

# Task Master Feature Integration Guidelines

## Feature Placement Decision Process

- **Identify Feature Type** (See [`architecture.md`](mdc:.roo/rules/architecture.md) for module details):
  - **Data Manipulation**: Features that create, read, update, or delete tasks belong in [`task-manager.js`](mdc:scripts/modules/task-manager.js). Follow guidelines in [`tasks.md`](mdc:.roo/rules/tasks.md).
  - **Dependency Management**: Features that handle task relationships belong in [`dependency-manager.js`](mdc:scripts/modules/dependency-manager.js). Follow guidelines in [`dependencies.md`](mdc:.roo/rules/dependencies.md).
  - **User Interface**: Features that display information to users belong in [`ui.js`](mdc:scripts/modules/ui.js). Follow guidelines in [`ui.md`](mdc:.roo/rules/ui.md).
  - **AI Integration**: Features that use AI models belong in [`ai-services.js`](mdc:scripts/modules/ai-services.js).
  - **Cross-Cutting**: Features that don't fit one category may need components in multiple modules

- **Command-Line Interface** (See [`commands.md`](mdc:.roo/rules/commands.md)):
  - All new user-facing commands should be added to [`commands.js`](mdc:scripts/modules/commands.js)
  - Use consistent patterns for option naming and help text
  - Follow the Commander.js model for subcommand structure

## Implementation Pattern

The standard pattern for adding a feature follows this workflow:

1. **Core Logic**: Implement the business logic in the appropriate module (e.g., [`task-manager.js`](mdc:scripts/modules/task-manager.js)).
2. **AI Integration (If Applicable)**: 
   - Import necessary service functions (e.g., `generateTextService`, `streamTextService`) from [`ai-services-unified.js`](mdc:scripts/modules/ai-services-unified.js).
   - Prepare parameters (`role`, `session`, `systemPrompt`, `prompt`).
   - Call the service function.
   - Handle the response (direct text or stream object).
   - **Important**: Prefer `generateTextService` for calls sending large context (like stringified JSON) where incremental display is not needed. See [`ai_services.md`](mdc:.roo/rules/ai_services.md) for detailed usage patterns and cautions.
3. **UI Components**: Add any display functions to [`ui.js`](mdc:scripts/modules/ui.js) following [`ui.md`](mdc:.roo/rules/ui.md).
4. **Command Integration**: Add the CLI command to [`commands.js`](mdc:scripts/modules/commands.js) following [`commands.md`](mdc:.roo/rules/commands.md).
5. **Testing**: Write tests for all components of the feature (following [`tests.md`](mdc:.roo/rules/tests.md))
6. **Configuration**: Update configuration settings or add new ones in [`config-manager.js`](mdc:scripts/modules/config-manager.js) and ensure getters/setters are appropriate. Update documentation in [`utilities.md`](mdc:.roo/rules/utilities.md) and [`taskmaster.md`](mdc:.roo/rules/taskmaster.md). Update the `.taskmasterconfig` structure if needed.
7. **Documentation**: Update help text and documentation in [`dev_workflow.md`](mdc:.roo/rules/dev_workflow.md) and [`taskmaster.md`](mdc:.roo/rules/taskmaster.md).

## Critical Checklist for New Features

- **Comprehensive Function Exports**:
  - ✅ **DO**: Export **all core functions, helper functions (like `generateSubtaskPrompt`), and utility methods** needed by your new function or command from their respective modules.
  - ✅ **DO**: **Explicitly review the module's `export { ... }` block** at the bottom of the file to ensure every required dependency (even seemingly minor helpers like `findTaskById`, `taskExists`, specific prompt generators, AI call handlers, etc.) is included.
  - ❌ **DON'T**: Assume internal functions are already exported - **always verify**. A missing export will cause runtime errors (e.g., `ReferenceError: generateSubtaskPrompt is not defined`).
  - **Example**: If implementing a feature that checks task existence, ensure the helper function is in exports:
  ```javascript
  // At the bottom of your module file:
  export {
    // ... existing exports ...
    yourNewFunction,
    taskExists,  // Helper function used by yourNewFunction
    findTaskById, // Helper function used by yourNewFunction
    generateSubtaskPrompt, // Helper needed by expand/add features
    getSubtasksFromAI,     // Helper needed by expand/add features
  };
  ```

- **Parameter Completeness and Matching**:
  - ✅ **DO**: Pass all required parameters to functions you call within your implementation
  - ✅ **DO**: Check function signatures before implementing calls to them
  - ✅ **DO**: Verify that direct function parameters match their core function counterparts
  - ✅ **DO**: When implementing a direct function for MCP, ensure it only accepts parameters that exist in the core function
  - ✅ **DO**: Verify the expected *internal structure* of complex object parameters (like the `mcpLog` object, see mcp.md for the required logger wrapper pattern)
  - ❌ **DON'T**: Add parameters to direct functions that don't exist in core functions
  - ❌ **DON'T**: Assume default parameter values will handle missing arguments
  - ❌ **DON'T**: Assume object parameters will work without verifying their required internal structure or methods.
  - **Example**: When calling file generation, pass all required parameters:
  ```javascript
  // ✅ DO: Pass all required parameters
  await generateTaskFiles(tasksPath, path.dirname(tasksPath));
  
  // ❌ DON'T: Omit required parameters
  await generateTaskFiles(tasksPath); // Error - missing outputDir parameter
  ```
  
  **Example**: Properly match direct function parameters to core function:
  ```javascript
  // Core function signature
  async function expandTask(tasksPath, taskId, numSubtasks, useResearch = false, additionalContext = '', options = {}) {
    // Implementation...
  }
  
  // ✅ DO: Match direct function parameters to core function
  export async function expandTaskDirect(args, log, context = {}) {
    // Extract only parameters that exist in the core function
    const taskId = parseInt(args.id, 10);
    const numSubtasks = args.num ? parseInt(args.num, 10) : undefined;
    const useResearch = args.research === true;
    const additionalContext = args.prompt || '';
    
    // Call core function with matched parameters
    const result = await expandTask(
      tasksPath,
      taskId,
      numSubtasks,
      useResearch,
      additionalContext,
      { mcpLog: log, session: context.session }
    );
    
    // Return result
    return { success: true, data: result, fromCache: false };
  }
  
  // ❌ DON'T: Use parameters that don't exist in the core function
  export async function expandTaskDirect(args, log, context = {}) {
    // DON'T extract parameters that don't exist in the core function!
    const force = args.force === true; // ❌ WRONG - 'force' doesn't exist in core function
    
    // DON'T pass non-existent parameters to core functions
    const result = await expandTask(
      tasksPath,
      args.id,
      args.num,
      args.research,
      args.prompt,
      force, // ❌ WRONG - this parameter doesn't exist in the core function
      { mcpLog: log }
    );
  }
  ```

- **Consistent File Path Handling**:
  - ✅ DO: Use consistent file naming conventions: `task_${id.toString().padStart(3, '0')}.txt`
  - ✅ DO: Use `path.join()` for composing file paths
  - ✅ DO: Use appropriate file extensions (.txt for tasks, .json for data)
  - ❌ DON'T: Hardcode path separators or inconsistent file extensions
  - **Example**: Creating file paths for tasks:
  ```javascript
  // ✅ DO: Use consistent file naming and path.join
  const taskFileName = path.join(
    path.dirname(tasksPath), 
    `task_${taskId.toString().padStart(3, '0')}.txt`
  );
  
  // ❌ DON'T: Use inconsistent naming or string concatenation 
  const taskFileName = path.dirname(tasksPath) + '/' + taskId + '.md';
  ```

- **Error Handling and Reporting**:
  - ✅ DO: Use structured error objects with code and message properties
  - ✅ DO: Include clear error messages identifying the specific problem
  - ✅ DO: Handle both function-specific errors and potential file system errors
  - ✅ DO: Log errors at appropriate severity levels
  - **Example**: Structured error handling in core functions:
  ```javascript
  try {
    // Implementation...
  } catch (error) {
    log('error', `Error removing task: ${error.message}`);
    throw {
      code: 'REMOVE_TASK_ERROR',
      message: error.message,
      details: error.stack
    };
  }
  ```

- **Silent Mode Implementation**:
  - ✅ **DO**: Import all silent mode utilities together:
    ```javascript
    import { enableSilentMode, disableSilentMode, isSilentMode } from '../../../../scripts/modules/utils.js';
    ```
  - ✅ **DO**: Always use `isSilentMode()` function to check global silent mode status, never reference global variables.
  - ✅ **DO**: Wrap core function calls **within direct functions** using `enableSilentMode()` and `disableSilentMode()` in a `try/finally` block if the core function might produce console output (like banners, spinners, direct `console.log`s) that isn't reliably controlled by an `outputFormat` parameter.
    ```javascript
    // Direct Function Example:
    try {
      // Prefer passing 'json' if the core function reliably handles it
      const result = await coreFunction(...args, 'json'); 
      // OR, if outputFormat is not enough/unreliable:
      // enableSilentMode(); // Enable *before* the call
      // const result = await coreFunction(...args);
      // disableSilentMode(); // Disable *after* the call (typically in finally)

      return { success: true, data: result };
    } catch (error) {
       log.error(`Error: ${error.message}`);
       return { success: false, error: { message: error.message } };
    } finally {
       // If you used enable/disable, ensure disable is called here
       // disableSilentMode(); 
    }
    ```
  - ✅ **DO**: Core functions themselves *should* ideally check `outputFormat === 'text'` before displaying UI elements (banners, spinners, boxes) and use internal logging (`log`/`report`) that respects silent mode. The `enable/disableSilentMode` wrapper in the direct function is a safety net.
  - ✅ **DO**: Handle mixed parameter/global silent mode correctly for functions accepting both (less common now, prefer `outputFormat`):
    ```javascript
    // Check both the passed parameter and global silent mode
    const isSilent = silentMode || (typeof silentMode === 'undefined' && isSilentMode());
    ```
  - ❌ **DON'T**: Forget to disable silent mode in a `finally` block if you enabled it.
  - ❌ **DON'T**: Access the global `silentMode` flag directly.

- **Debugging Strategy**:
  - ✅ **DO**: If an MCP tool fails with vague errors (e.g., JSON parsing issues like `Unexpected token ... is not valid JSON`), **try running the equivalent CLI command directly in the terminal** (e.g., `task-master expand --all`). CLI output often provides much more specific error messages (like missing function definitions or stack traces from the core logic) that pinpoint the root cause.
  - ❌ **DON'T**: Rely solely on MCP logs if the error is unclear; use the CLI as a complementary debugging tool for core logic issues.

```javascript
// 1. CORE LOGIC: Add function to appropriate module (example in task-manager.js)
/**
 * Archives completed tasks to archive.json
 * @param {string} tasksPath - Path to the tasks.json file
 * @param {string} archivePath - Path to the archive.json file
 * @returns {number} Number of tasks archived
 */
async function archiveTasks(tasksPath, archivePath = 'tasks/archive.json') {
  // Implementation...
  return archivedCount;
}

// Export from the module
export {
  // ... existing exports ...
  archiveTasks,
};
```

```javascript
// 2. AI Integration: Add import and use necessary service functions
import { generateTextService } from './ai-services-unified.js';

// Example usage:
async function handleAIInteraction() {
  const role = 'user';
  const session = 'exampleSession';
  const systemPrompt = 'You are a helpful assistant.';
  const prompt = 'What is the capital of France?';

  const result = await generateTextService(role, session, systemPrompt, prompt);
  console.log(result);
}

// Export from the module
export {
  // ... existing exports ...
  handleAIInteraction,
};
```

```javascript
// 3. UI COMPONENTS: Add display function to ui.js
/**
 * Display archive operation results
 * @param {string} archivePath - Path to the archive file
 * @param {number} count - Number of tasks archived
 */
function displayArchiveResults(archivePath, count) {
  console.log(boxen(
    chalk.green(`Successfully archived ${count} tasks to ${archivePath}`),
    { padding: 1, borderColor: 'green', borderStyle: 'round' }
  ));
}

// Export from the module
export {
  // ... existing exports ...
  displayArchiveResults,
};
```

```javascript
// 4. COMMAND INTEGRATION: Add to commands.js
import { archiveTasks } from './task-manager.js';
import { displayArchiveResults } from './ui.js';

// In registerCommands function
programInstance
  .command('archive')
  .description('Archive completed tasks to separate file')
  .option('-f, --file <file>', 'Path to the tasks file', 'tasks/tasks.json')
  .option('-o, --output <file>', 'Archive output file', 'tasks/archive.json')
  .action(async (options) => {
    const tasksPath = options.file;
    const archivePath = options.output;
    
    console.log(chalk.blue(`Archiving completed tasks from ${tasksPath} to ${archivePath}...`));
    
    const archivedCount = await archiveTasks(tasksPath, archivePath);
    displayArchiveResults(archivePath, archivedCount);
  });
```

## Cross-Module Features

For features requiring components in multiple modules:

- ✅ **DO**: Create a clear unidirectional flow of dependencies
  ```javascript
  // In task-manager.js
  function analyzeTasksDifficulty(tasks) {
    // Implementation...
    return difficultyScores;
  }
  
  // In ui.js - depends on task-manager.js
  import { analyzeTasksDifficulty } from './task-manager.js';
  
  function displayDifficultyReport(tasks) {
    const scores = analyzeTasksDifficulty(tasks);
    // Render the scores...
  }
  ```

- ❌ **DON'T**: Create circular dependencies between modules
  ```javascript
  // In task-manager.js - depends on ui.js
  import { displayDifficultyReport } from './ui.js';
  
  function analyzeTasks() {
    // Implementation...
    displayDifficultyReport(tasks); // WRONG! Don't call UI functions from task-manager
  }
  
  // In ui.js - depends on task-manager.js
  import { analyzeTasks } from './task-manager.js';
  ```

## Command-Line Interface Standards

- **Naming Conventions**:
  - Use kebab-case for command names (`analyze-complexity`, not `analyzeComplexity`)
  - Use kebab-case for option names (`--output-format`, not `--outputFormat`) 
  - Use the same option names across commands when they represent the same concept

- **Command Structure**:
  ```javascript
  programInstance
    .command('command-name')
    .description('Clear, concise description of what the command does')
    .option('-s, --short-option <value>', 'Option description', 'default value')
    .option('--long-option <value>', 'Option description')
    .action(async (options) => {
      // Command implementation
    });
  ```

## Utility Function Guidelines

When adding utilities to [`utils.js`](mdc:scripts/modules/utils.js):

- Only add functions that could be used by multiple modules
- Keep utilities single-purpose and purely functional
- Document parameters and return values

```javascript
/**
 * Formats a duration in milliseconds to a human-readable string
 * @param {number} ms - Duration in milliseconds
 * @returns {string} Formatted duration string (e.g., "2h 30m 15s")
 */
function formatDuration(ms) {
  // Implementation...
  return formatted;
}
```

## Writing Testable Code

When implementing new features, follow these guidelines to ensure your code is testable:

- **Dependency Injection**
  - Design functions to accept dependencies as parameters
  - Avoid hard-coded dependencies that are difficult to mock
  ```javascript
  // ✅ DO: Accept dependencies as parameters
  function processTask(task, fileSystem, logger) {
    fileSystem.writeFile('task.json', JSON.stringify(task));
    logger.info('Task processed');
  }
  
  // ❌ DON'T: Use hard-coded dependencies
  function processTask(task) {
    fs.writeFile('task.json', JSON.stringify(task));
    console.log('Task processed');
  }
  ```

- **Separate Logic from Side Effects**
  - Keep pure logic separate from I/O operations or UI rendering
  - This allows testing the logic without mocking complex dependencies
  ```javascript
  // ✅ DO: Separate logic from side effects
  function calculateTaskPriority(task, dependencies) {
    // Pure logic that returns a value
    return computedPriority;
  }
  
  function displayTaskPriority(task, dependencies) {
    const priority = calculateTaskPriority(task, dependencies);
    console.log(`Task priority: ${priority}`);
  }
  ```

- **Callback Functions and Testing**
  - When using callbacks (like in Commander.js commands), define them separately
  - This allows testing the callback logic independently
  ```javascript
  // ✅ DO: Define callbacks separately for testing
  function getVersionString() {
    // Logic to determine version
    return version;
  }
  
  // In setupCLI
  programInstance.version(getVersionString);
  
  // In tests
  test('getVersionString returns correct version', () => {
    expect(getVersionString()).toBe('1.5.0');
  });
  ```

- **UI Output Testing**
  - For UI components, focus on testing conditional logic rather than exact output
  - Use string pattern matching (like `expect(result).toContain('text')`)
  - Pay attention to emojis and formatting which can make exact string matching difficult
  ```javascript
  // ✅ DO: Test the essence of the output, not exact formatting
  test('statusFormatter shows done status correctly', () => {
    const result = formatStatus('done');
    expect(result).toContain('done');
    expect(result).toContain('✅');
  });
  ```

## Testing Requirements

Every new feature **must** include comprehensive tests following the guidelines in [`tests.md`](mdc:.roo/rules/tests.md). Testing should include:

1. **Unit Tests**: Test individual functions and components in isolation
   ```javascript
   // Example unit test for a new utility function
   describe('newFeatureUtil', () => {
     test('should perform expected operation with valid input', () => {
       expect(newFeatureUtil('valid input')).toBe('expected result');
     });
     
     test('should handle edge cases appropriately', () => {
       expect(newFeatureUtil('')).toBeNull();
     });
   });
   ```

2. **Integration Tests**: Verify the feature works correctly with other components
   ```javascript
   // Example integration test for a new command
   describe('newCommand integration', () => {
     test('should call the correct service functions with parsed arguments', () => {
       const mockService = jest.fn().mockResolvedValue('success');
       // Set up test with mocked dependencies
       // Call the command handler
       // Verify service was called with expected arguments
     });
   });
   ```

3. **Edge Cases**: Test boundary conditions and error handling
   - Invalid inputs
   - Missing dependencies
   - File system errors
   - API failures

4. **Test Coverage**: Aim for at least 80% coverage for all new code

5. **Jest Mocking Best Practices**
   - Follow the mock-first-then-import pattern as described in [`tests.md`](mdc:.roo/rules/tests.md)
   - Use jest.spyOn() to create spy functions for testing
   - Clear mocks between tests to prevent interference
   - See the Jest Module Mocking Best Practices section in [`tests.md`](mdc:.roo/rules/tests.md) for details

When submitting a new feature, always run the full test suite to ensure nothing was broken:

```bash
npm test
```

## Documentation Requirements

For each new feature:

1. Add help text to the command definition
2. Update [`dev_workflow.md`](mdc:.roo/rules/dev_workflow.md) with command reference
3. Consider updating [`architecture.md`](mdc:.roo/rules/architecture.md) if the feature significantly changes module responsibilities.

Follow the existing command reference format:
```markdown
- **Command Reference: your-command**
  - CLI Syntax: `task-master your-command [options]`
  - Description: Brief explanation of what the command does
  - Parameters:
    - `--option1=<value>`: Description of option1 (default: 'default')
    - `--option2=<value>`: Description of option2 (required)
  - Example: `task-master your-command --option1=value --option2=value2`
  - Notes: Additional details, limitations, or special considerations
```

For more information on module structure, see [`MODULE_PLAN.md`](mdc:scripts/modules/MODULE_PLAN.md) and follow [`self_improve.md`](mdc:scripts/modules/self_improve.md) for best practices on updating documentation.

## Adding MCP Server Support for Commands

Integrating Task Master commands with the MCP server (for use by tools like Roo Code) follows a specific pattern distinct from the CLI command implementation, prioritizing performance and reliability.

- **Goal**: Leverage direct function calls to core logic, avoiding CLI overhead.
- **Reference**: See [`mcp.md`](mdc:.roo/rules/mcp.md) for full details.

**MCP Integration Workflow**:

1.  **Core Logic**: Ensure the command's core logic exists and is exported from the appropriate module (e.g., [`task-manager.js`](mdc:scripts/modules/task-manager.js)).
2.  **Direct Function Wrapper (`mcp-server/src/core/direct-functions/`)**:
    - Create a new file (e.g., `your-command.js`) in `mcp-server/src/core/direct-functions/` using **kebab-case** naming.
    - Import the core logic function, necessary MCP utilities like **`findTasksJsonPath` from `../utils/path-utils.js`**, and **silent mode utilities**: `import { enableSilentMode, disableSilentMode } from '../../../../scripts/modules/utils.js';`
    - Implement an `async function yourCommandDirect(args, log)` using **camelCase** with `Direct` suffix.
    - **Path Finding**: Inside this function, obtain the `tasksPath` by calling `const tasksPath = findTasksJsonPath(args, log);`. This relies on `args.projectRoot` (derived from the session) being passed correctly.
    - Perform validation on other arguments received in `args`.
    - **Implement Silent Mode**: Wrap core function calls with `enableSilentMode()` and `disableSilentMode()` to prevent logs from interfering with JSON responses.
    - **If Caching**: Implement caching using `getCachedOrExecute` from `../../tools/utils.js`.
    - **If Not Caching**: Directly call the core logic function within a try/catch block.
    - Format the return as `{ success: true/false, data/error, fromCache: boolean }`.
    - Export the wrapper function.

3.  **Update `task-master-core.js` with Import/Export**: Import and re-export your `*Direct` function and add it to the `directFunctions` map.

4.  **Create MCP Tool (`mcp-server/src/tools/`)**:
    - Create a new file (e.g., `your-command.js`) using **kebab-case**.
    - Import `zod`, `handleApiResult`, **`withNormalizedProjectRoot` HOF**, and your `yourCommandDirect` function.
    - Implement `registerYourCommandTool(server)`.
    - **Define parameters**: Make `projectRoot` optional (`z.string().optional().describe(...)`) as the HOF handles fallback.
    - Consider if this operation should run in the background using `AsyncOperationManager`.
    - Implement the standard `execute` method **wrapped with `withNormalizedProjectRoot`**:
      ```javascript
      execute: withNormalizedProjectRoot(async (args, { log, session }) => {
          // args.projectRoot is now normalized
          const { projectRoot /*, other args */ } = args;
          // ... resolve tasks path if needed using normalized projectRoot ...
          const result = await yourCommandDirect(
              { /* other args */, projectRoot /* if needed by direct func */ }, 
              log, 
              { session }
          );
          return handleApiResult(result, log);
      })
      ```

5.  **Register Tool**: Import and call `registerYourCommandTool` in `mcp-server/src/tools/index.js`.

6.  **Update `mcp.json`**: Add the new tool definition to the `tools` array in `.roo/mcp.json`.

## Implementing Background Operations

For long-running operations that should not block the client, use the AsyncOperationManager:

1. **Identify Background-Appropriate Operations**:
   - ✅ **DO**: Use async operations for CPU-intensive tasks like task expansion or PRD parsing
   - ✅ **DO**: Consider async operations for tasks that may take more than 1-2 seconds
   - ❌ **DON'T**: Use async operations for quick read/status operations
   - ❌ **DON'T**: Use async operations when immediate feedback is critical

2. **Use AsyncOperationManager in MCP Tools**:
   ```javascript
   import { asyncOperationManager } from '../core/utils/async-manager.js';
   
   // In execute method:
   const operationId = asyncOperationManager.addOperation(
     expandTaskDirect, // The direct function to run in background
     { ...args, projectRoot: rootFolder }, // Args to pass to the function
     { log, reportProgress, session } // Context to preserve for the operation
   );
   
   // Return immediate response with operation ID
   return createContentResponse({
     message: "Operation started successfully",
     operationId,
     status: "pending"
   });
   ```

3. **Implement Progress Reporting**:
   - ✅ **DO**: Use the reportProgress function in direct functions:
   ```javascript
   // In your direct function:
   if (reportProgress) {
     await reportProgress({ progress: 50 }); // 50% complete
   }
   ```
   - AsyncOperationManager will forward progress updates to the client

4. **Check Operation Status**:
   - Implement a way for clients to check status using the `get_operation_status` MCP tool
   - Return appropriate status codes and messages

## Project Initialization

When implementing project initialization commands:

1. **Support Programmatic Initialization**:
   - ✅ **DO**: Design initialization to work with both CLI and MCP
   - ✅ **DO**: Support non-interactive modes with sensible defaults
   - ✅ **DO**: Handle project metadata like name, description, version
   - ✅ **DO**: Create necessary files and directories

2. **In MCP Tool Implementation**:
   ```javascript
   // In initialize-project.js MCP tool:
   import { z } from "zod";
   import { initializeProjectDirect } from "../core/task-master-core.js";
   
   export function registerInitializeProjectTool(server) {
     server.addTool({
       name: "initialize_project",
       description: "Initialize a new Task Master project",
       parameters: z.object({
         projectName: z.string().optional().describe("The name for the new project"),
         projectDescription: z.string().optional().describe("A brief description"),
         projectVersion: z.string().optional().describe("Initial version (e.g., '0.1.0')"),
         // Add other parameters as needed
       }),
       execute: async (args, { log, reportProgress, session }) => {
         try {
           // No need for project root since we're creating a new project
           const result = await initializeProjectDirect(args, log);
           return handleApiResult(result, log, 'Error initializing project');
         } catch (error) {
           log.error(`Error in initialize_project: ${error.message}`);
           return createErrorResponse(`Failed to initialize project: ${error.message}`);
         }
       }
     });
   }
   ```
