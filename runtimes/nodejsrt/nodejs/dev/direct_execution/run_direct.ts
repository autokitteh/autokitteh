// import Runner from "../../runtime/runner";
import Runner from "../../testdata/simple-test/runtime/.ak/runtime/runner";
import { DirectHandlerClient } from "./direct_client";
import * as path from "path";
import fs from "fs";
import { TextEncoder } from "util";
import * as yaml from "js-yaml";

/**
 * Directly run a function in a specified directory by extracting the entry point from autokitteh.yaml
 * Example usage: ts-node run_direct.ts ../../examples/simple-test
 */
async function runDirect() {
  // Parse command line arguments
  const args = process.argv.slice(2);
  if (args.length < 1) {
    console.error("Usage: ts-node run_direct.ts <test_directory>");
    console.error("Example: ts-node run_direct.ts ../../examples/simple-test");
    process.exit(1);
  }

  const inputDir = args[0];
  console.log(`Input directory: ${inputDir}`);

  console.log("Starting direct runner execution...");

  // Create absolute path to the test directory
  const testDir = path.resolve(inputDir);
  if (!fs.existsSync(testDir)) {
    throw new Error(`Test directory not found at ${testDir}`);
  }

  console.log(`Using test directory: ${testDir}`);

  // Read autokitteh.yaml to get the entry point
  const yamlPath = path.join(testDir, 'autokitteh.yaml');
  if (!fs.existsSync(yamlPath)) {
    throw new Error(`autokitteh.yaml not found at ${yamlPath}`);
  }

  console.log(`Reading configuration from ${yamlPath}`);
  const yamlContent = fs.readFileSync(yamlPath, 'utf8');

  // Parse YAML file using js-yaml
  const config = yaml.load(yamlContent) as any;

  // Extract the first trigger's call property (entry point)
  if (!config.project || !config.project.triggers || !config.project.triggers.length) {
    throw new Error('No triggers found in autokitteh.yaml');
  }

  const firstTrigger = config.project.triggers[0];
  if (!firstTrigger.call) {
    throw new Error('No call entry point found in the first trigger');
  }

  // Get the entry point and convert .js to .ts if needed for TypeScript
  let entryPoint = firstTrigger.call;

  console.log(`Using trigger: ${firstTrigger.name}`);
  console.log(`Entry point: ${entryPoint}`);

  // Create direct client
  const directClient = new DirectHandlerClient();
  console.log("Created DirectHandlerClient");

  // Create runner with direct client
  console.log("Creating runner...");
  const runner = new Runner(
    "direct-runner-id",
    testDir,
    directClient as any
  );
  console.log("Runner created");

  // Start the runner
  console.log("Starting runner...");
  await runner.start();
  console.log("Runner started");

  // Create a mock router to capture the service implementation
  console.log("Creating service...");
  const mockServiceImplementation: any = {};
  const mockRouter = {
    service: (service: any, implementation: any) => {
      console.log("Capturing service implementation");
      Object.assign(mockServiceImplementation, implementation);
      return mockRouter;
    }
  };

  // Register the runner service on our mock router
  console.log("Registering runner service...");
  (runner.createService() as any)(mockRouter);
  console.log("Available runner service methods:", Object.keys(mockServiceImplementation));

  // Connect the runner service back to the direct client
  console.log("Connecting runner service to direct client...");
  directClient.setRunnerService(mockServiceImplementation);
  console.log("Service connected");

  // Emit the started event to complete initialization
  console.log("Emitting started event...");
  try {
    // Access events property and emit the started event
    (runner as any).events.emit('started');
    console.log("Started event emitted");
  } catch (err) {
    console.error("Could not emit started event:", err);
    process.exit(1);
  }

  // Create a simple event object with data
  const encoder = new TextEncoder();
  const eventData = {
    token: "direct-execution-token",
    args: ["TestUser"]
  };
  console.log("Created event data:", eventData);

  // Create the start request
  console.log("Preparing to execute function...");
  const startRequest = {
    entryPoint: entryPoint,
    event: {
      data: encoder.encode(JSON.stringify(eventData))
    }
  };

  try {
    // Call the function
    console.log("Executing function...");
    const result = await directClient.start(startRequest);
    console.log("Execution result:", result);
  } catch (error) {
    console.error("Execution failed:", error);
  } finally {
    // Stop the runner and clean up
    console.log("Cleaning up...");
    directClient.setActive(false);

    // Give time for any operations to complete before exiting
    setTimeout(() => {
      console.log("Execution completed");
      process.exit(0);
    }, 1000);
  }
}

// Run the direct execution
runDirect().catch(error => {
  console.error("Direct execution failed:", error);
  process.exit(1);
});
