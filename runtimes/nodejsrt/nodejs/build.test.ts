import { build } from './runtime/builder/build';
import {exec} from 'child_process';
import {promisify} from 'util';

test('test', async () => {
    await build('./testdata/akevents', './testdata-build/akevents');
    const execAsync = promisify(exec);
    await execAsync('npm install', {cwd: './testdata-build/akevents'});

})

