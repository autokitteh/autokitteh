import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { installDependencies, setupTypeScript } from '../local_runner';

describe('Local Runner', () => {
    let testDir: string;

    beforeEach(async () => {
        // Create a temporary directory for each test
        testDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), 'ak-test-'));
    });

    afterEach(async () => {
        // Cleanup after each test
        await fs.promises.rm(testDir, { recursive: true, force: true });
    });

    describe('Dependencies Installation', () => {
        it('should install user dependencies from package.json', async () => {
            // Create a basic package.json
            const packageJson = {
                name: "test-project",
                version: "1.0.0",
                dependencies: {
                    "lodash": "^4.17.21"
                }
            };
            await fs.promises.writeFile(
                path.join(testDir, 'package.json'),
                JSON.stringify(packageJson, null, 2)
            );

            // Install dependencies
            await installDependencies(testDir);

            // Verify node_modules exists and contains lodash
            const nodeModulesPath = path.join(testDir, 'node_modules');
            const lodashPath = path.join(nodeModulesPath, 'lodash');
            expect(fs.existsSync(nodeModulesPath)).toBe(true);
            expect(fs.existsSync(lodashPath)).toBe(true);

            // Verify ak framework dependencies are installed
            const akDeps = [
                '@connectrpc/connect',
                '@connectrpc/connect-fastify',
                '@connectrpc/connect-node',
                '@bufbuild/protobuf',
                'fastify',
                'commander',
                '@babel/core',
                '@babel/traverse',
                '@babel/generator',
                '@babel/parser',
                'typescript',
                'ts-node',
                '@types/node'
            ];

            for (const dep of akDeps) {
                expect(fs.existsSync(path.join(nodeModulesPath, dep))).toBe(true);
            }
        });

        it('should handle projects without package.json', async () => {
            // Install dependencies without a package.json
            await installDependencies(testDir);

            // Verify node_modules exists with ak framework dependencies
            const nodeModulesPath = path.join(testDir, 'node_modules');
            expect(fs.existsSync(nodeModulesPath)).toBe(true);

            // Verify package.json was created
            const packageJsonPath = path.join(testDir, 'package.json');
            expect(fs.existsSync(packageJsonPath)).toBe(true);
        });
    });

    describe('Project Structure', () => {
        it('should create proper directory structure', async () => {
            // Setup TypeScript configuration
            await setupTypeScript(testDir);

            // Verify .ak directory structure
            const akDir = path.join(testDir, '.ak');
            const typesDir = path.join(akDir, 'types');
            expect(fs.existsSync(akDir)).toBe(true);
            expect(fs.existsSync(typesDir)).toBe(true);

            // Verify global.d.ts exists and has correct content
            const globalDtsPath = path.join(typesDir, 'global.d.ts');
            expect(fs.existsSync(globalDtsPath)).toBe(true);
            const globalDtsContent = await fs.promises.readFile(globalDtsPath, 'utf8');
            expect(globalDtsContent).toContain('declare function ak_call');
            expect(globalDtsContent).toContain('Promise<unknown>');

            // Verify tsconfig.json exists and has correct content
            const tsconfigPath = path.join(testDir, 'tsconfig.json');
            expect(fs.existsSync(tsconfigPath)).toBe(true);
            const tsconfig = JSON.parse(await fs.promises.readFile(tsconfigPath, 'utf8'));
            expect(tsconfig.compilerOptions.typeRoots).toContain('.ak/types');
        });

        it('should merge existing tsconfig.json with default config', async () => {
            // Create existing tsconfig.json with custom settings
            const existingTsConfig = {
                compilerOptions: {
                    target: "ES2022",
                    baseUrl: "./src",
                    paths: {
                        "@/*": ["*"]
                    }
                }
            };
            await fs.promises.writeFile(
                path.join(testDir, 'tsconfig.json'),
                JSON.stringify(existingTsConfig, null, 2)
            );

            // Setup TypeScript configuration
            await setupTypeScript(testDir);

            // Verify merged config
            const tsconfig = JSON.parse(
                await fs.promises.readFile(path.join(testDir, 'tsconfig.json'), 'utf8')
            );
            expect(tsconfig.compilerOptions.target).toBe("ES2022"); // Preserved from existing
            expect(tsconfig.compilerOptions.baseUrl).toBe("./src"); // Preserved from existing
            expect(tsconfig.compilerOptions.paths).toEqual({"@/*": ["*"]}); // Preserved from existing
            expect(tsconfig.compilerOptions.typeRoots).toContain('.ak/types'); // Added from default
            expect(tsconfig.compilerOptions.module).toBe("commonjs"); // Added from default
        });
    });
});