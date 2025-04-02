import { runDirect } from './dev/direct_execution/runner_direct';

test('test', async () => {
    await runDirect('testdata/simple-test/runtime');
})

