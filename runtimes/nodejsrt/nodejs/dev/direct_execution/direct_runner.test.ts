import Runner from "../../runtime/runner";
import { DirectActivityWaiter } from "./direct_waiter";
import { DirectHandlerClient } from "./direct_client";
import * as path from "path";
import * as fs from "fs";
import { TextEncoder, TextDecoder } from "util";
import { ConnectRouter } from "@connectrpc/connect";
import { test } from '@jest/globals';

test('full flow', async () => {
  console.log("Starting direct runner test...");

  // Create absolute path to the test directory
  const testDir = path.resolve(__dirname);
  console.log(`Using absolute test directory: ${testDir}`);

  // Confirm test module exists
  const testModulePath = path.join(testDir, "test_module.ts");
  if (!fs.existsSync(testModulePath)) {
    throw new Error(`Test module not found at ${testModulePath}`);
  }
  console.log(`Using test module at ${testModulePath}`);

  // Create direct execution components in the correct order
  const directWaiter = new DirectActivityWaiter("test-runner-id");
  const directClient = new DirectHandlerClient(directWaiter);

  // Create runner with direct client and waiter
  console.log("Creating runner...");
  const runner = new Runner(
    "test-runner-id",
    testDir,
    directClient as any,
    directWaiter as any
  );

  // Start the runner
  console.log("Starting runner...");
  await runner.start();

  // Create a mock router to capture the service implementation
  console.log("Creating service...");
  const mockServiceImplementation: any = {};
  const mockRouter = {
    service: (service: any, implementation: any) => {
      Object.assign(mockServiceImplementation, implementation);
    }
  } as ConnectRouter;

  // Register the runner service on our mock router
  runner.createService()(mockRouter);

  // Emit the started event to complete initialization
  console.log("Emitting started event...");
  try {
    // Try accessing the events property safely
    const runnerAny = runner as any;
    if (runnerAny.events && typeof runnerAny.events.emit === 'function') {
      runnerAny.events.emit('started');
    }
  } catch (err) {
    console.log("Could not emit started event, continuing anyway:", err);
  }

  // Create a simple event object with data
  const encoder = new TextEncoder();
  const eventData = {
    token: "test-token",
    args: ["TestUser"]
  };

  // Create the start request with entry point and event
  console.log("Executing test function...");
  const startRequest = {
    entryPoint: "test_module.ts:callExternalService",
    event: {
      data: encoder.encode(JSON.stringify(eventData))
    }
  };

  // Call the start method on the service implementation
  const result = await mockServiceImplementation.start(startRequest);

  // Log the result
  console.log("Execution result:", result);

  // Clean up the runner
  console.log("Stopping runner...");
  runner.stop();

  // Set the client's active state to false so the runner can exit
  console.log("Setting client isActive to false...");
  directClient.setActive(false);

  console.log("Test completed successfully!");
});
