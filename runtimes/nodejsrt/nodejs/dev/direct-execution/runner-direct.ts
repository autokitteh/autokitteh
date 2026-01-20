import Runner from "../../runtime/runner/runner"
import {DirectHandlerClient} from "./client-direct";
import * as helpers from "./helpers";
/**
 * Directly run a function in a specified directory by extracting the entry point from autokitteh.yaml
 * Example usage: ts-node runner_direct.ts ../../examples/simple-test
 */
export async function runDirect(inputDir: string = "", eventArgs: any = {}) {
    const args = process.argv.slice(2);
    if (args.length < 1) {
        console.error("Usage: ts-node runner_direct.ts <test_directory>");
        console.error("Example: ts-node runner_direct.ts ../../examples/simple-test");
        process.exit(1);
    }
    console.log(`Input directory: ${inputDir}`);
    console.log(`event arguments: ${eventArgs}`);
    console.log("Starting direct runner execution...");

    // Prepare request
    inputDir = helpers.validateInputDirectory(inputDir);
    const config = helpers.readConfiguration(inputDir);
    const request = helpers.createRequest(config, eventArgs);

    // Setup mock client
    const {mockRouter, mockServiceImplementation} = helpers.setupMockRouter();
    const directClient = new DirectHandlerClient();
    directClient.setRunnerService(mockServiceImplementation);

    // Create the runner and apply mock
    const runner = new Runner("direct-runner-id", inputDir, directClient as any);
    const clientService = runner.createService();
    (clientService as any)(mockRouter);

    // Register the runner service mock router
    (runner.createService() as any)(mockRouter);

    // Connect the runner service back to the direct client
    directClient.setRunnerService(mockServiceImplementation);


    // Start Runner and
    await runner.start();
    try {
        (runner as any).events.emit('started');
        console.log("Started event emitted");
    } catch (err) {
        console.error("Could not emit started event:", err);
        process.exit(1);
    }

    // Process request
    try {
        const result = await directClient.start(request);
        console.log("Execution result:", result);
    } catch (error) {
        console.error("Execution failed:", error);
    } finally {
        console.log("Cleaning up...");
        directClient.setActive(false);
        setTimeout(() => {console.log("Execution completed");
            process.exit(0);
        }, 1000);
    }
}
