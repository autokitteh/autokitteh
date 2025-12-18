import * as path from 'path';
import { mainDirect } from './dev/direct-execution/main-direct';

test('test', async () => {
    await mainDirect(
        path.resolve('./examples-build/invoices-app'),
        {year: '2025', month: '1'},
        // 'testdata/simple-test/runtime/.ak/runtime/main.ts'
    );
})

