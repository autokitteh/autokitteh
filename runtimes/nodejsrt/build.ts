import path from "path";
import {patchCode} from "../ast_utils";
import {listFiles} from "../file_utils";
import fs from "fs";

const codeDir = 'test_data/invoices-app';
const rootDir = 'src';
const buildDir = path.join(codeDir, 'build');

async function build(dir: string): Promise<void> {

    await fs.promises.rm(buildDir, {recursive: true, force: true});

    const files = await listFiles(dir);

    const ignorePatterns = [/node_modules/, /dist/, /tests/];

    const filteredFiles = files.filter(file => {
        const relativePath = path.relative(dir, file);
        return ignorePatterns.every(pattern => !pattern.test(relativePath));
    });

    // Generate the the build folder
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

    // Modify package.json programmatically to add multiple packages using sync methods
    const packageJsonPath = path.join(buildDir, "package.json");
    const packagesToAdd = {
        "@bufbuild/protobuf": "^2.2.3", // Adjust versions as required
        "lodash": "^4.17.21",
        "axios": "^1.4.0"
    };

    // Read package.json synchronously
    const packageJsonData = JSON.parse(fs.readFileSync(packageJsonPath, "utf-8"));

    // Ensure dependencies field exists
    packageJsonData.dependencies ||= {};

    for (const [pkg, version] of Object.entries(packagesToAdd)) {
        if (!packageJsonData.dependencies[pkg]) {
            packageJsonData.dependencies[pkg] = version;
        }
    }

    // Write the updated package.json synchronously
    fs.writeFileSync(packageJsonPath, JSON.stringify(packageJsonData, null, 2), "utf-8");

}

(async () => {
    await build(codeDir);
})();
