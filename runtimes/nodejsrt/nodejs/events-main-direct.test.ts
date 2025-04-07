import { mainDirect } from './dev/direct_execution/main_direct';

test('test', async () => {
    await mainDirect(
        'testdata/events',
        {name: 'test', value: 123456789},
        // 'testdata/simple-test/runtime/.ak/runtime/main.ts'
    );
})

