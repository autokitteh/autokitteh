import * as path from 'path';
import { mainDirect } from './main-direct';
import {buildDirect} from './build-direct';

const projectDir = path.resolve('./testdata/simple-test/runtime');
const deployDir = path.resolve('./testdata-build/simple-test/runtime');

test('test', async () => {

    await buildDirect(projectDir, deployDir);

    await mainDirect(
        deployDir,
        {name: 'test', value: 123456789},
        // 'testdata/simple-test/runtime/.ak/runtime/main.ts'
    );
})
