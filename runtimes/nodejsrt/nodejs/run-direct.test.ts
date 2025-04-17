import { runDirect } from './dev/direct_execution/runner-direct';

test('test', async () => {
    await runDirect('testdata/simple-test/runtime');
})

