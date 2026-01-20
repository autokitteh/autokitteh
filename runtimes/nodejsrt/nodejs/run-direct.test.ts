import { runDirect } from './dev/direct-execution/runner-direct';

test('test', async () => {
    await runDirect('testdata/simple-test/runtime');
})

