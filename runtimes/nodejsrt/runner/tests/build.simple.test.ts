import { describe, it, expect, beforeAll, afterAll } from '@jest/globals';
import * as fs from 'fs/promises';
import * as path from 'path';
import { build } from '../build';

describe('Simple Build Process', () => {
    const testDir = path.join(__dirname, 'temp_test');
    const outputDir = path.join(__dirname, 'temp_output');

    beforeAll(async () => {
        // Create test directories
        await fs.mkdir(testDir, { recursive: true });
        await fs.mkdir(outputDir, { recursive: true });
    });

    afterAll(async () => {
        // Cleanup test directories
        await fs.rm(testDir, { recursive: true, force: true });
        await fs.rm(outputDir, { recursive: true, force: true });
    });

    it('should copy and patch JavaScript files', async () => {
        // Create test JS file
        const jsContent = `
            async function testFunction() {
                return await someOtherFunction();
            }
            export { testFunction };
        `;
        await fs.writeFile(path.join(testDir, 'test.js'), jsContent);

        // Run build
        await build(testDir, outputDir);

        // Verify file was copied and patched
        const outputContent = await fs.readFile(path.join(outputDir, 'test.js'), 'utf8');
        expect(outputContent).toContain('ak_call');
        expect(outputContent).toContain('testFunction');
    });

    it('should copy and patch TypeScript files', async () => {
        // Create test TS file
        const tsContent = `
            async function greet(name: string): Promise<string> {
                return await Promise.resolve(\`Hello \${name}!\`);
            }
            export { greet };
        `;
        await fs.writeFile(path.join(testDir, 'test.ts'), tsContent);

        // Run build
        await build(testDir, outputDir);

        // Verify file was copied and patched
        const outputContent = await fs.readFile(path.join(outputDir, 'test.ts'), 'utf8');
        expect(outputContent).toContain('ak_call');
        expect(outputContent).toContain('greet');
    });

    it('should copy non-JS/TS files without modification', async () => {
        // Create test JSON file
        const jsonContent = '{"key": "value"}';
        await fs.writeFile(path.join(testDir, 'test.json'), jsonContent);

        // Run build
        await build(testDir, outputDir);

        // Verify file was copied without changes
        const outputContent = await fs.readFile(path.join(outputDir, 'test.json'), 'utf8');
        expect(outputContent).toBe(jsonContent);
    });

    it('should ignore specified patterns', async () => {
        // Create files in node_modules, dist, and .git directories
        await fs.mkdir(path.join(testDir, 'node_modules'), { recursive: true });
        await fs.mkdir(path.join(testDir, 'dist'), { recursive: true });
        await fs.mkdir(path.join(testDir, '.git'), { recursive: true });

        await fs.writeFile(path.join(testDir, 'node_modules/test.js'), 'test');
        await fs.writeFile(path.join(testDir, 'dist/test.js'), 'test');
        await fs.writeFile(path.join(testDir, '.git/test.js'), 'test');

        // Run build
        await build(testDir, outputDir);

        // Verify ignored files were not copied
        const nodeModulesExists = await fs.stat(path.join(outputDir, 'node_modules'))
            .then(() => true)
            .catch(() => false);
        const distExists = await fs.stat(path.join(outputDir, 'dist'))
            .then(() => true)
            .catch(() => false);
        const gitExists = await fs.stat(path.join(outputDir, '.git'))
            .then(() => true)
            .catch(() => false);

        expect(nodeModulesExists).toBe(false);
        expect(distExists).toBe(false);
        expect(gitExists).toBe(false);
    });

    it('should handle nested directories', async () => {
        // Create nested directory structure
        const nestedDir = path.join(testDir, 'src/components');
        await fs.mkdir(nestedDir, { recursive: true });

        const jsContent = `
            async function Component() {
                return await Promise.resolve('Hello');
            }
        `;
        await fs.writeFile(path.join(nestedDir, 'component.js'), jsContent);

        // Run build
        await build(testDir, outputDir);

        // Verify nested structure was preserved and file was patched
        const outputContent = await fs.readFile(path.join(outputDir, 'src/components/component.js'), 'utf8');
        expect(outputContent).toContain('ak_call');
        expect(outputContent).toContain('Component');
    });

    it('should handle runtime generated functions', async () => {
        // Create test JS file with runtime function generation
        const jsContent = `
            // Function constructor
            const dynamicFn = new Function('param', 'return async function() { return await fetch(param); }');
            export { dynamicFn };

            // eval generated function
            const evalFn = eval('async function evalFunction(url) { return await fetch(url); }');
            export { evalFn };

            // Arrow function assigned to variable
            const arrowFn = async (url) => { return await fetch(url); };
            export { arrowFn };
        `;
        await fs.writeFile(path.join(testDir, 'dynamic.js'), jsContent);

        // Run build
        await build(testDir, outputDir);

        // Verify file was copied and patched
        const outputContent = await fs.readFile(path.join(outputDir, 'dynamic.js'), 'utf8');
        expect(outputContent).toContain('ak_call');
        expect(outputContent).toContain('fetch');
        expect(outputContent).toContain('dynamicFn');
        expect(outputContent).toContain('evalFn');
        expect(outputContent).toContain('arrowFn');
    });
}); 