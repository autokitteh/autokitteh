import { runDirect } from './dev/direct_execution/run_direct';

test('test', async () => {
    await runDirect('testdata/simple-test/runtime');
})

