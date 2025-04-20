// import {setupRuntime, setupServer, configureServer, startRuntime} from "../../runtime/main";
import {DirectHandlerClient} from "./client-direct";
import {FastifyInstance} from "fastify";
import * as helpers from "./helpers";

/**
 * Directly run the main function with a specified directory by extracting configuration from autokitteh.yaml
 * Example usage: ts-node main_direct.ts ../../examples/simple-test
 */
export async function mainDirect(inputDir: string = "", eventArgs: any = {}, mainPath: string = "runtime/runner/main.ts") {

    mainPath = helpers.validateInputDirectory(mainPath);
    // const mainModule = require(mainPath);
    const mainModule = await import(mainPath);
    const {setupRuntime, setupServer, configureServer, startRuntime} = mainModule;

    const options = {
        workerAddress: "direct://localhost:0", // Dummy address since we're using direct client
        port: 0, // Will be assigned by the system
        runnerId: "direct-runner-id",
        codeDir: inputDir
    };

    // Prepare request
    inputDir = helpers.validateInputDirectory(inputDir);
    const config = helpers.readConfiguration(inputDir);
    const request = helpers.createRequest(config, eventArgs);

    // setup mock client
    const {mockRouter, mockServiceImplementation} = helpers.setupMockRouter();
    const directClient = new DirectHandlerClient();
    directClient.setRunnerService(mockServiceImplementation);

    //create the runner and apply mock
    const {runner} = setupRuntime(options, directClient);
    const clientService = runner.createService();
    (clientService as any)(mockRouter);

    // Setup server with the same service
    const server = await setupServer(clientService);
    configureServer(server as unknown as FastifyInstance, runner);

    try {
        await startRuntime(runner, server as unknown as FastifyInstance, options);
        console.log(`Executing request with entry point: ${request.entryPoint}`);
        const result = await directClient.start(request);
        console.log("Test execution result:", result);

    } catch (error) {
        console.error("Execution failed:", error);
    } finally {
        console.log("Cleaning up...");
        directClient.setActive(false);
        await (server as unknown as FastifyInstance).close();
        console.log("Execution completed");
        // process.exit(0);
    }
}

// Execute if run directly
if (require.main === module) {
    import {Command} from "commander";
    const program = new Command();
    program.requiredOption('--input-dir <TYPE>', 'inputDir')
    program.option('--args <TYPE>', 'eventArgs')

    program.parse(process.argv);
    const options = program.opts();

    const inputDir = process.argv[2];
    const eventArgs = process.argv[3] ? JSON.parse(process.argv[3]) : {};
    void mainDirect(options.inputDir, options.eventArgs);
}
