import { main } from './testdata/simple-test/runtime/.ak/runtime/runner/main';

test('test', async () => {

    const options = {
        workerAddress: 'localhost:8080',
        port: 61320,
        runnerId: "runner_01jqkhbbc1e7wt99z9k39qt55t",
        codeDir: "testdata/simple-test"
    }
    await main(options);
})

