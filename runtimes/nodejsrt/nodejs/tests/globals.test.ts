/**
 * Tests for the global autokitteh object
 *
 * This file tests the actual initialization of global objects.
 * Since Jest runs tests in isolation, we can test the initialization
 * without worrying about conflicts with other tests.
 */
import {describe, expect, test, jest, afterEach} from '@jest/globals';
import {EventSubscriber} from '../runtime/nodejs-sdk/autokitteh/events';
import {initializeGlobals} from '../runtime/runner/runtime';
import {Waiter} from '../runtime/runner/ak_call';

describe('Global autokitteh object initialization', () => {
    // Mock EventSubscriber for initialization
    const mockSubscribe = jest.fn().mockImplementation(() => Promise.resolve('test-signal-id'));
    const mockNextEvent = jest.fn().mockImplementation(() => Promise.resolve({test: 'data'}));
    const mockUnsubscribe = jest.fn().mockImplementation(() => Promise.resolve());

    // Mock waiter with proper typing
    const mockWaiter = {
        execute_signal: jest.fn().mockImplementation(() => Promise.resolve({})),
        reply_signal: jest.fn().mockImplementation(() => Promise.resolve()),
        wait: jest.fn().mockImplementation(() => Promise.resolve({})),
        done: jest.fn().mockImplementation(() => Promise.resolve()),
        getRunId: jest.fn().mockReturnValue('test-run-id'),
        setRunId: jest.fn(),
        setRunnerId: jest.fn()
    } as unknown as Waiter;

    // Create event subscriber with mocked methods
    const eventSubscriber = {
        subscribe: mockSubscribe,
        nextEvent: mockNextEvent,
        unsubscribe: mockUnsubscribe
    } as unknown as EventSubscriber;

    beforeAll(() =>{
        initializeGlobals(mockWaiter, eventSubscriber);
    })

    afterEach(() => {
        jest.clearAllMocks();
    });

    test('should initialize global.ak_call property', () => {
        expect('ak_call' in global).toBe(true);
        expect(typeof (global as any).ak_call).toBe('function');
    });

    test('should initialize global.autokitteh property', () => {
        expect('autokitteh' in global).toBe(true);
        expect(typeof (global as any).autokitteh).toBe('object');
    });

    test('global.autokitteh.subscribe should call EventSubscriber.subscribe', async () => {
        await (global as any).autokitteh.subscribe('test-source', 'filter');

        expect(mockSubscribe).toHaveBeenCalledWith('test-source', 'filter');
    });

    test('global.autokitteh.nextEvent should call EventSubscriber.nextEvent', async () => {
        await (global as any).autokitteh.nextEvent('test-signal-id', {timeout: 5});

        expect(mockNextEvent).toHaveBeenCalledWith('test-signal-id', {timeout: 5});
    });

    test('global.autokitteh.unsubscribe should call EventSubscriber.unsubscribe', async () => {
        await (global as any).autokitteh.unsubscribe('test-signal-id');

        expect(mockUnsubscribe).toHaveBeenCalledWith('test-signal-id');
    });

    test('global properties should be non-configurable', () => {
        const akCallDescriptor = Object.getOwnPropertyDescriptor(global, 'ak_call');
        expect(akCallDescriptor?.configurable).toBe(false);
        expect(akCallDescriptor?.writable).toBe(false);

        const autokittehDescriptor = Object.getOwnPropertyDescriptor(global, 'autokitteh');
        expect(autokittehDescriptor?.configurable).toBe(false);
        expect(autokittehDescriptor?.writable).toBe(false);
    });
});

// Test to verify the actual functionality of the global object when used in user code
describe('Global autokitteh usage', () => {
    // These tests simulate how a user would call the global methods
    // rather than testing that they're properly wired up

    test('autokitteh global methods exist', () => {
        expect(typeof (global as any).autokitteh.subscribe).toBe('function');
        expect(typeof (global as any).autokitteh.nextEvent).toBe('function');
        expect(typeof (global as any).autokitteh.unsubscribe).toBe('function');
    });

    test('ak_call global function exists', () => {
        expect(typeof (global as any).ak_call).toBe('function');
    });
});
