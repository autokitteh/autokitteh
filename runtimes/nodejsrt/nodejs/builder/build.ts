import path from "path";
import {patchCode} from "../common/ast_utils";
import {listFiles} from "../common/file_utils";
import fs from "fs";

export async function build(inputDir: string, outputDir: string): Promise<void> {
    const files = await listFiles(inputDir);
    const ignorePatterns = [/node_modules/, /dist/, /\.git/];

    const filteredFiles = files.filter(file => {
        const relativePath = path.relative(inputDir, file);
        return ignorePatterns.every(pattern => !pattern.test(relativePath));
    });

    // Copy files to output directory, patching js/ts files
    await Promise.all(filteredFiles.map(async (file) => {
        const relativePath = path.relative(inputDir, file);
        const destPath = path.join(outputDir, relativePath);

        // Create directory if it doesn't exist
        await fs.promises.mkdir(path.dirname(destPath), { recursive: true });

        if (file.endsWith('.js') || file.endsWith('.ts')) {
            const code = await fs.promises.readFile(file, 'utf-8');
            const patchedCode = await patchCode(code);
            await fs.promises.writeFile(destPath, patchedCode, 'utf-8');
        } else {
            await fs.promises.copyFile(file, destPath);
        }
    }));
}

// Only run the command-line interface if this file is being run directly
if (require.main === module) {
    // Get input and output directories from command line args
    const [inputDir, outputDir] = process.argv.slice(2);
    if (!inputDir || !outputDir) {
        console.error("Please provide both input and output directories");
        process.exit(1);
    }

    (async () => {
        try {
            await build(inputDir, outputDir);
            console.log("Build completed successfully");
        } catch (error) {
            console.error("Build failed:", error);
            process.exit(1);
        }
    })();
}
