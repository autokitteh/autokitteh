import { mainDirect } from './dev/direct_execution/main_direct';

test('test', async () => {
    await mainDirect('testdata/simple-test/runtime',{},'testdata/simple-test/runtime/.ak/runtime/main.ts');
})

