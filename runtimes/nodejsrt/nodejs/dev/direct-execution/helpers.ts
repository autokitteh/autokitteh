import * as path from "path";
import * as fs from "fs";
import * as yaml from "js-yaml";
import { symlink } from 'fs/promises';

export function validateInputDirectory(inputDir: string): string {

    // console.log(`Input directory: ${inputDir}`);
    // Create absolute path to the test directory
    inputDir = path.resolve(inputDir);
    if (!fs.existsSync(inputDir)) {
        throw new Error(`Input directory not found at ${inputDir}`);
    }

    // console.log(`Using test directory: ${inputDir}`);
    return inputDir;
}

export function readConfiguration(inputDir: string): any {
    // Read autokitteh.yaml to get configuration
    const yamlPath = path.join(inputDir, 'autokitteh.yaml');
    if (!fs.existsSync(yamlPath)) {
        throw new Error(`autokitteh.yaml not found at ${yamlPath}`);
    }

    // console.log(`Reading configuration from ${yamlPath}`);
    const yamlContent = fs.readFileSync(yamlPath, 'utf8');

    // Parse YAML file using js-yaml
    return yaml.load(yamlContent) as any;
}

export function setupMockRouter() {
    const mockServiceImplementation: any = {};
    const mockRouter = {
        service: (service: any, implementation: any) => {
            // console.info("Capturing service implementation");
            Object.assign(mockServiceImplementation, implementation);
            return mockRouter;
        }
    };
    return {mockRouter, mockServiceImplementation};
}

export function createRequest(config: any, args: string[] = []) {
    const Request = {
        entryPoint: config.project?.triggers?.[0]?.call,
        event: {
            data: Buffer.from(JSON.stringify({
                token: "direct-execution-token",
                args: args
            }))
        }
    };

    if (!Request.entryPoint) {
        throw new Error("No entry point found in autokitteh.yaml triggers");
    }
    return Request;
}



export async function linkAutokitteh(userProjectDir: string) {
    const linkPath = path.join(userProjectDir, 'node_modules', 'autokitteh');
    const targetPath = path.resolve(__dirname, '../../nodejs-sdk/autokitteh');

    await (path.dirname(linkPath), { recursive: true });
    try {
        await symlink(targetPath, linkPath, 'dir');
    } catch (err: any) {
        if (err.code !== 'EEXIST') {
            throw err;
        }
    }

}


