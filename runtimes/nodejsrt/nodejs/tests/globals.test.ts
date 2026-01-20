/**
 * Tests for the global autokitteh object
 *
 * This file tests the actual initialization of global objects.
 * Since Jest runs tests in isolation, we can test the initialization
 * without worrying about conflicts with other tests.
 */
import {describe, expect, test, jest, beforeAll, afterAll} from '@jest/globals';
import {initializeGlobals} from '../runtime/runner/runtime';
import {Waiter} from '../runtime/runner/ak_call';
import {SysCalls} from '../runtime/runner/syscalls';

// Extend the global object type

describe('Global Functions', () => {
    // Mock waiter for initialization
    const mockWaiter = {
        execute_signal: jest.fn().mockImplementation(() => Promise.resolve({})),
        reply_signal: jest.fn().mockImplementation(() => Promise.resolve()),
        wait: jest.fn().mockImplementation(() => Promise.resolve({})),
        done: jest.fn().mockImplementation(() => Promise.resolve()),
        getRunId: jest.fn().mockReturnValue('test-run-id'),
        setRunId: jest.fn(),
        setRunnerId: jest.fn()
    } as unknown as Waiter;

    // Mock syscalls for initialization
    const mockSyscalls = {
        subscribe: jest.fn(),
        nextEvent: jest.fn(),
        unsubscribe: jest.fn()
    } as unknown as SysCalls;

    beforeAll(() => {
        initializeGlobals(mockWaiter, mockSyscalls);
    });

    afterAll(() => {
        // Clean up globals
        const globalWithCustomProps = global as { ak_call?: unknown; syscalls?: unknown };
        delete globalWithCustomProps.ak_call;
        delete globalWithCustomProps.syscalls;
    });

    test('global.ak_call should be defined', () => {
        expect((global as any).ak_call).toBeDefined();
        expect(typeof (global as any).ak_call).toBe('function');
    });

    test('global.syscalls should be defined', () => {
        expect((global as any).syscalls).toBeDefined();
        expect(typeof (global as any).syscalls).toBe('object');
    });

    test('global.syscalls.subscribe should call SysCalls.subscribe', async () => {
        await (global as any).syscalls.subscribe('test-source', 'test-filter');
        expect(mockSyscalls.subscribe).toHaveBeenCalledWith('test-source', 'test-filter');
    });

    test('global.syscalls.nextEvent should call SysCalls.nextEvent', async () => {
        await (global as any).syscalls.nextEvent('test-id', {timeout: 30});
        expect(mockSyscalls.nextEvent).toHaveBeenCalledWith('test-id', {timeout: 30});
    });

    test('global.syscalls.unsubscribe should call SysCalls.unsubscribe', async () => {
        await (global as any).syscalls.unsubscribe('test-id');
        expect(mockSyscalls.unsubscribe).toHaveBeenCalledWith('test-id');
    });

    test('global properties should be non-configurable and non-writable', () => {
        const akCallDescriptor = Object.getOwnPropertyDescriptor(global, 'ak_call');
        expect(akCallDescriptor?.configurable).toBe(false);
        expect(akCallDescriptor?.writable).toBe(false);

        const syscallsDescriptor = Object.getOwnPropertyDescriptor(global, 'syscalls');
        expect(syscallsDescriptor?.configurable).toBe(false);
        expect(syscallsDescriptor?.writable).toBe(false);
    });
});
