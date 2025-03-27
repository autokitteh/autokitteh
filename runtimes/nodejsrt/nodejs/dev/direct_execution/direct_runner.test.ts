import {test} from '@jest/globals';

import Runner from "../../runtime/runner/runner";
import {DirectHandlerClient} from "./direct_client";
import * as helpers from "./helpers";


test('full flow', async () => {
    console.log("Starting direct runner test...");

    // Run parameters
    const testDir = helpers.validateInputDirectory(__dirname);
    const startRequest = {
        entryPoint: 'test_module.ts:callExternalService',
        event: {
            data: Buffer.from(JSON.stringify({
                token: "test-token",
                args: ["TestUser"]
            }))
        }
    };

    // setup mocks
    const {mockRouter, mockServiceImplementation} = helpers.setupMockRouter();

    // Create direct client - no need for separate waiter
    const directClient = new DirectHandlerClient();
    directClient.setRunnerService(mockServiceImplementation);

    //create the runner and apply mock
    const runner = new Runner("test-runner-id", testDir, directClient as any);
    (runner.createService() as any)(mockRouter);

    // Start the runner
    await runner.start();
    (runner as any).events.emit('started');

    // execute request
    console.log("Calling start method directly on client...");
    const result = await directClient.start(startRequest);
    console.log("Execution result:", result);

    // Stop client
    directClient.setActive(false);


    console.log("Test completed successfully!");
});
