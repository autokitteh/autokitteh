import * as path from "path"
import * as fs from "fs";
import { build } from "../../runtime/builder/build";

/**
 * Builds a project by patching its code and copying to an output directory.
 * This is a direct execution wrapper around the build function.
 *
 * @param inputDir - The directory containing the source code to build
 * @param outputDir - The directory where the built code should be placed
 * @returns A promise that resolves when the build is complete
 */
export async function buildDirect(inputDir: string, outputDir: string): Promise<void> {
    inputDir = path.resolve(inputDir);
    outputDir = path.resolve(outputDir);
    
    // Check if input directory exists
    try {
        const stats = await fs.promises.stat(inputDir);
        if (!stats.isDirectory()) {
            throw new Error(`Input path ${inputDir} is not a directory`);
        }
    } catch {
        // Any error means the directory doesn't exist or can't be accessed
        throw new Error(`Input directory ${inputDir} does not exist or is not accessible`);
    }
    
    await fs.promises.mkdir(outputDir, { recursive: true });
    await build(inputDir, outputDir);
}

// Command-line interface when run directly
if (require.main === module) {
    const [inputDir, outputDir] = process.argv.slice(2);
    if (!inputDir || !outputDir) {
        console.error("Usage: node build_direct.js <input_directory> <output_directory>");
        process.exit(1);
    }

    (async () => {
        try {
            await buildDirect(inputDir, outputDir);
        } catch (error) {
            if (error instanceof Error) {
                console.error(error.message);
            } else {
                console.error(String(error));
            }
            process.exit(1);
        }
    })();
}
