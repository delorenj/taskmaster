// Deno version of the Task Master MCP Server
import { dirname, join } from "https://deno.land/std/path/mod.ts";

// Initialize logging
const logFile = Deno.env.get("DENO_LOG") || "/tmp/deno-mcp.log";
function log(message: string) {
  try {
    Deno.writeTextFileSync(logFile, `${new Date().toISOString()} - ${message}\n`, { append: true });
  } catch (error) {
    // Fallback to stderr if unable to write to log file
    console.error(`[LOG] ${message}`);
  }
}

log("MCP server starting up");

// Simple line-based JSON reader from stdin
async function* readJsonLines() {
  const buffer = new Uint8Array(10240); // 10KB buffer
  const decoder = new TextDecoder();
  
  while (true) {
    try {
      const readResult = await Deno.stdin.read(buffer);
      if (readResult === null) break; // EOF
      
      const text = decoder.decode(buffer.subarray(0, readResult));
      const lines = text.trim().split('\n');
      
      for (const line of lines) {
        if (line.trim()) {
          try {
            log(`Received: ${line}`);
            yield JSON.parse(line);
          } catch (error) {
            log(`Error parsing JSON: ${error.message}`);
          }
        }
      }
    } catch (error) {
      log(`Error reading stdin: ${error.message}`);
      break;
    }
  }
}

// Write JSON response to stdout
async function writeResponse(response: any) {
  const encoder = new TextEncoder();
  const responseText = JSON.stringify(response) + "\n";
  log(`Sending response: ${responseText}`);
  await Deno.stdout.write(encoder.encode(responseText));
}

// Main MCP server process
async function main() {
  log("Starting MCP server main loop");
  
  try {
    for await (const message of readJsonLines()) {
      if (message.type === "initialize") {
        log("Handling initialize message");
        
        // Respond to initialize with minimal schema
        const response = {
          type: "response",
          id: message.id,
          result: {
            schema: {
              tools: [
                {
                  type: "function",
                  function: {
                    name: "ping",
                    description: "Test MCP connection",
                    parameters: {
                      type: "object",
                      properties: {},
                      required: []
                    }
                  }
                }
              ]
            }
          }
        };
        
        await writeResponse(response);
      } else if (message.type === "toolCall" && message.params?.name === "ping") {
        log("Handling ping tool call");
        
        // Respond to ping with success
        const response = {
          type: "response",
          id: message.id,
          result: { result: "pong" }
        };
        
        await writeResponse(response);
      } else {
        log(`Unknown message type: ${message.type}`);
      }
    }
  } catch (error) {
    log(`Fatal error in main loop: ${error.message}`);
    log(error.stack || "No stack trace available");
    Deno.exit(1);
  }
}

// Start the server
log("MCP server process started");
main().catch(error => {
  log(`Unhandled error in main: ${error.message}`);
  log(error.stack || "No stack trace available");
  Deno.exit(1);
});