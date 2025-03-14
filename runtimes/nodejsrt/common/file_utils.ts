import fs from "fs";
import path from "path";

export async function listFiles(dir: string): Promise<string[]> {
    let results: string[] = [];
    const files = await fs.promises.readdir(dir);

    for (const file of files) {
        const filePath = path.join(dir, file);
        const stat = await fs.promises.stat(filePath);

        if (stat.isDirectory()) {
            results = results.concat(await listFiles(filePath));
        } else {
            results.push(filePath);
        }
    }

    return results;
}
