import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

// Import the functions we're testing
import { setupTypeScript, mergeConfigs } from '../local_runner';

describe('TypeScript Setup', () => {
    let testDir: string;

    beforeEach(async () => {
        // Create a temporary directory for each test
        testDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), 'ak-test-'));
    });

    afterEach(async () => {
        // Cleanup after each test
        await fs.promises.rm(testDir, { recursive: true, force: true });
    });

    it('should create proper TypeScript configuration', async () => {
        // Call the actual setup function
        await setupTypeScript(testDir);
        
        // Test global.d.ts creation
        const globalDtsPath = path.join(testDir, '.ak', 'types', 'global.d.ts');
        expect(fs.existsSync(globalDtsPath)).toBe(true);
        
        const globalDtsContent = await fs.promises.readFile(globalDtsPath, 'utf8');
        expect(globalDtsContent).toContain('function ak_call');
        expect(globalDtsContent).toContain('Promise<unknown>');

        // Test tsconfig.json creation
        const tsconfigPath = path.join(testDir, 'tsconfig.json');
        expect(fs.existsSync(tsconfigPath)).toBe(true);

        const tsconfig = JSON.parse(await fs.promises.readFile(tsconfigPath, 'utf8'));
        expect(tsconfig.compilerOptions.typeRoots).toContain('.ak/types');
    });

    it('should merge existing tsconfig with new settings', async () => {
        // Create existing tsconfig
        const existingTsConfig = {
            compilerOptions: {
                target: "ES2019",
                customOption: true
            }
        };
        await fs.promises.writeFile(
            path.join(testDir, 'tsconfig.json'),
            JSON.stringify(existingTsConfig, null, 2)
        );

        // Call the actual setup function
        await setupTypeScript(testDir);

        // Test tsconfig merging
        const tsconfigPath = path.join(testDir, 'tsconfig.json');
        const tsconfig = JSON.parse(await fs.promises.readFile(tsconfigPath, 'utf8'));
        
        expect(tsconfig.compilerOptions.customOption).toBe(true);
        expect(tsconfig.compilerOptions.typeRoots).toContain('.ak/types');
        expect(tsconfig.compilerOptions.target).toBe('ES2020'); // Our setting should override
    });

    it('should merge configs correctly', () => {
        const existing = {
            compilerOptions: {
                target: "ES2019",
                customOption: true
            }
        };

        const newConfig = {
            compilerOptions: {
                target: "ES2020",
                typeRoots: ['.ak/types']
            }
        };

        const merged = mergeConfigs(existing, newConfig);
        expect(merged.compilerOptions.target).toBe('ES2020');
        expect(merged.compilerOptions.customOption).toBe(true);
        expect(merged.compilerOptions.typeRoots).toContain('.ak/types');
    });
}); 