import * as path from 'path';
import { mainDirect } from './dev/direct_execution/main_direct';

test('test', async () => {
    await mainDirect(
        path.resolve('./testdata-build/akevents'),
        {name: 'test', value: 123456789},
        // 'testdata/simple-test/runtime/.ak/runtime/main.ts'
    );
})

