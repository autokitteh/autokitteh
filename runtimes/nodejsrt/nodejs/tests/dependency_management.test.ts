import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { exec as execCallback } from 'child_process';

// Define types for mocked exec
type ExecCallback = (error: Error | null, stdout: string, stderr: string) => void;
type ExecOptions = { cwd?: string };

// Mock child_process.exec
const mockExec = jest.fn((command: string, options: ExecOptions, callback?: ExecCallback) => {
    if (callback) {
        callback(null, '', '');
    }
    return {} as ReturnType<typeof execCallback>;
});

jest.mock('child_process', () => ({
    exec: mockExec
}));

// Import after mocking
import { installDependencies } from '../local_runner';

describe('Dependency Management', () => {
    let tempDir: string;

    beforeEach(() => {
        tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'test-'));
        mockExec.mockClear();
    });

    afterEach(() => {
        fs.rmSync(tempDir, { recursive: true, force: true });
    });

    it('should install user dependencies first', async () => {
        await installDependencies(tempDir);
        expect(mockExec).toHaveBeenCalledWith(
            'npm install',
            { cwd: tempDir },
            expect.any(Function)
        );
    });

    it('should install all ak dependencies', async () => {
        await installDependencies(tempDir);
        
        // Verify ak dependencies were installed
        const expectedDeps = [
            '@connectrpc/connect@^2.0.1',
            '@connectrpc/connect-fastify@^2.0.1',
            '@connectrpc/connect-node@^2.0.1',
            '@bufbuild/protobuf@^2.2.3',
            'fastify@^5.2.1',
            'commander@^13.1.0',
            '@babel/core@^7.26.0',
            '@babel/traverse@^7.26.5',
            '@babel/generator@^7.26.5',
            '@babel/parser@^7.26.5',
            'typescript@^5.7.3',
            'ts-node@^10.9.2',
            '@types/node@^22.13.10'
        ];

        expectedDeps.forEach(dep => {
            expect(mockExec).toHaveBeenCalledWith(
                `npm install --save ${dep}`,
                { cwd: tempDir },
                expect.any(Function)
            );
        });
    });

    it('should handle npm install errors', async () => {
        const error = new Error('npm install failed');
        mockExec.mockImplementationOnce((command: string, options: ExecOptions, callback?: ExecCallback) => {
            if (callback) {
                callback(error, '', 'npm install failed');
            }
            return {} as ReturnType<typeof execCallback>;
        });

        await expect(installDependencies(tempDir)).rejects.toThrow('npm install failed');
    });

    it('should handle invalid directory', async () => {
        const invalidDir = path.join(tempDir, 'nonexistent');
        await expect(installDependencies(invalidDir)).rejects.toThrow();
    });
}); 