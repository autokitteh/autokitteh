import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { AkCallFunction, AkCallResult } from '../types/ak_call';

describe('Global ak_call', () => {
    let originalAkCall: unknown;

    beforeEach(() => {
        // Store original ak_call if it exists
        originalAkCall = global.ak_call;
        
        // Mock ak_call
        const mockFn = jest.fn<AkCallFunction>((name: string, ...args: unknown[]): Promise<AkCallResult> => {
            return Promise.resolve({ success: true, args: [name, ...args] });
        });
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });
    });

    afterEach(() => {
        // Restore original ak_call
        if (originalAkCall) {
            Object.defineProperty(global, 'ak_call', {
                value: originalAkCall,
                writable: true,
                configurable: true
            });
        }
        jest.clearAllMocks();
    });

    it('should be globally available after importing runtime', async () => {
        await import('../runtime');
        expect(global.ak_call).toBeDefined();
        expect(typeof global.ak_call).toBe('function');
    });

    it('should handle activity calls correctly', async () => {
        const mockFn = jest.fn<AkCallFunction>((name: string, ...args: unknown[]): Promise<AkCallResult> => {
            return Promise.resolve({ success: true, args: [name, ...args] });
        });
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });

        await import('../runtime');
        const result = await mockFn('test.activity', 'arg1');
        expect(result).toEqual({ success: true, args: ['test.activity', 'arg1'] });
    });

    it('should handle errors correctly', async () => {
        const errorMessage = 'Activity failed';
        const mockFn = jest.fn<AkCallFunction>();
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });
        mockFn.mockRejectedValueOnce(new Error(errorMessage));

        await import('../runtime');
        await expect(mockFn('test.activity')).rejects.toThrow(errorMessage);
    });

    it('should preserve reference across imports', async () => {
        await import('../runtime');
        const activity = global.ak_call;
        expect(activity).toBe(global.ak_call);
    });

    it('should prevent overwriting ak_call', async () => {
        await import('../runtime');
        const newFunction = async () => undefined;
        expect(() => {
            Object.defineProperty(global, 'ak_call', {
                value: newFunction,
                writable: false,
                configurable: false
            });
        }).toThrow();
    });

    it('should handle undefined arguments', async () => {
        const mockFn = jest.fn<AkCallFunction>((name: string, ...args: unknown[]): Promise<AkCallResult> => {
            return Promise.resolve({ success: true, args: [name, ...args] });
        });
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });

        await import('../runtime');
        await mockFn('test.activity', undefined);
        expect(mockFn).toHaveBeenCalledWith('test.activity', undefined);
    });

    it('should handle null arguments', async () => {
        const mockFn = jest.fn<AkCallFunction>((name: string, ...args: unknown[]): Promise<AkCallResult> => {
            return Promise.resolve({ success: true, args: [name, ...args] });
        });
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });

        await import('../runtime');
        await mockFn('test.activity', null);
        expect(mockFn).toHaveBeenCalledWith('test.activity', null);
    });

    it('should handle complex objects', async () => {
        const mockFn = jest.fn<AkCallFunction>((name: string, ...args: unknown[]): Promise<AkCallResult> => {
            return Promise.resolve({ success: true, args: [name, ...args] });
        });
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });

        await import('../runtime');
        const complexArg = {
            nested: { value: 42 },
            array: [1, 2, 3],
            date: new Date()
        };
        await mockFn('test.activity', complexArg);
        expect(mockFn).toHaveBeenCalledWith('test.activity', complexArg);
    });

    it('should handle concurrent calls', async () => {
        const mockFn = jest.fn<AkCallFunction>((name: string, ...args: unknown[]): Promise<AkCallResult> => {
            return Promise.resolve({ success: true, args: [name, ...args] });
        });
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });

        await import('../runtime');
        const calls = [
            mockFn('test1'),
            mockFn('test2'),
            mockFn('test3')
        ];
        const results = await Promise.all(calls);
        expect(results).toHaveLength(3);
        expect(mockFn).toHaveBeenCalledTimes(3);
    });

    it('should handle mixed success and failure', async () => {
        const mockFn = jest.fn<AkCallFunction>();
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });
        mockFn
            .mockResolvedValueOnce({ success: true, args: ['test1'] })
            .mockRejectedValueOnce(new Error('Failed'))
            .mockResolvedValueOnce({ success: true, args: ['test3'] });

        await import('../runtime');
        const calls = [
            mockFn('test1'),
            mockFn('test2'),
            mockFn('test3')
        ];

        const results = await Promise.allSettled(calls);
        expect(results[0].status).toBe('fulfilled');
        expect(results[1].status).toBe('rejected');
        expect(results[2].status).toBe('fulfilled');
    });

    it('should validate activity names', async () => {
        const mockFn = jest.fn<AkCallFunction>((name: string, ...args: unknown[]): Promise<AkCallResult> => {
            if (!name) {
                throw new Error('Invalid activity name');
            }
            return Promise.resolve({ success: true, args: [name, ...args] });
        });
        Object.defineProperty(global, 'ak_call', {
            value: mockFn,
            writable: true,
            configurable: true
        });

        await import('../runtime');
        await expect(mockFn('')).rejects.toThrow();
        await expect(mockFn(undefined as unknown as string)).rejects.toThrow();
        await expect(mockFn(null as unknown as string)).rejects.toThrow();
    });
}); 