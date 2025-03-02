import path from "path";
import {patchCode} from "../ast_utils";
import {listFiles} from "../file_utils";
import fs from "fs";

const codeDir = 'test_data/invoices-app';
const rootDir = 'src';
const buildDir = path.join(codeDir,'build');

async function patchDir(dir: string): Promise<void> {

    await fs.promises.rm(buildDir, {recursive: true, force: true});

    const files = await listFiles(dir);

    const ignorePatterns = [/node_modules/, /dist/, /tests/];

    const filteredFiles = files.filter(file => {
        const relativePath = path.relative(dir, file);
        return ignorePatterns.every(pattern => !pattern.test(relativePath));
    });

    await Promise.all(filteredFiles.map(async (file) => {
        const relativePath = path.relative(codeDir, file);
        const destPath = path.join(buildDir, relativePath);
        await fs.promises.mkdir(path.dirname(destPath), {recursive: true});

        if (file.endsWith('.js') || file.endsWith('.ts')) {
            const originalCode = fs.readFileSync(file, 'utf-8');
            const patchedCode = await patchCode(originalCode);
            await fs.promises.writeFile(destPath, patchedCode, 'utf-8');
        } else {
            await fs.promises.copyFile(file, destPath);
        }
    }));

    // Copy pb directory and ak_call.ts to build/ak
    const akDir = path.join(buildDir, rootDir, 'ak');
    const akFile = 'ak_call.ts';

    await fs.promises.mkdir(akDir, {recursive: true});
    await fs.promises.cp('pb', path.join(akDir, 'pb'), {recursive: true});
    await fs.promises.copyFile(akFile, path.join(akDir, akFile));


}

(async () => {
    await patchDir(codeDir);
})();
