import { build } from './builder/build';

test('test', async () => {
    await build('./simple-test', './simple-test-build');
})

