import { describe, expect, test, jest, beforeEach, afterEach } from '@jest/globals';
import { createClient } from '@connectrpc/connect';
import { HandlerService } from '../runtime/pb/autokitteh/user_code/v1/handler_svc_pb';
import Runner from '../runtime/runner';
import { EventEmitter } from 'events';
import { HandlerHealthResponse, IsActiveRunnerResponse, PrintResponse } from '../runtime/pb/autokitteh/user_code/v1/handler_svc_pb';

// Mock the client type
type MockClient = {
  health: jest.Mock<() => Promise<HandlerHealthResponse>>;
  isActiveRunner: jest.Mock<() => Promise<IsActiveRunnerResponse>>;
  print: jest.Mock<() => Promise<PrintResponse>>;
  activity: jest.Mock;
  done: jest.Mock;
  log: jest.Mock;
  sleep: jest.Mock;
  subscribe: jest.Mock;
  nextEvent: jest.Mock;
  unsubscribe: jest.Mock;
  startSession: jest.Mock;
  encodeJWT: jest.Mock;
  refreshOAuthToken: jest.Mock;
};

// Mock the HandlerService client
jest.mock('@connectrpc/connect', () => ({
    createClient: jest.fn()
}));

describe('Runner', () => {
    let runner: Runner;
    let mockClient: jest.Mocked<ReturnType<typeof createClient<typeof HandlerService>>>;

    beforeEach(() => {
        // Reset all mocks
        jest.clearAllMocks();

        // Create mock client with proper response types
        mockClient = {
            health: jest.fn<() => Promise<HandlerHealthResponse>>().mockResolvedValue(
                { error: '' } as HandlerHealthResponse
            ),
            isActiveRunner: jest.fn<() => Promise<IsActiveRunnerResponse>>().mockResolvedValue(
                { error: '', isActive: true } as IsActiveRunnerResponse
            ),
            print: jest.fn<() => Promise<PrintResponse>>().mockResolvedValue(
                { error: '' } as PrintResponse
            ),
        } as any;

        // Create runner instance
        runner = new Runner('test-runner-id', '/test/code/dir', mockClient);

        // Mock console methods
        console.log = jest.fn();
        console.error = jest.fn();
    });

    afterEach(() => {
        // Clean up any timers
        jest.useRealTimers();
    });

    describe('initialization', () => {
        test('should create runner with correct properties', () => {
            expect(runner).toBeInstanceOf(Runner);
            expect((runner as any).id).toBe('test-runner-id');
            expect((runner as any).codeDir).toBe('/test/code/dir');
            expect((runner as any).client).toBe(mockClient);
            expect((runner as any).events).toBeInstanceOf(EventEmitter);
            expect((runner as any).isStarted).toBe(false);
        });
    });

    describe('start', () => {
        beforeEach(() => {
            jest.useFakeTimers();
        });

        test('should start health checks and set up event listener', async () => {
            const startPromise = runner.start();

            // Verify health check is started
            expect(mockClient.isActiveRunner).toHaveBeenCalledWith({
                runnerId: 'test-runner-id'
            });

            // Emit the started event
            (runner as any).events.emit('started');

            // Fast-forward timers
            jest.advanceTimersByTime(1000);

            await startPromise;

            expect((runner as any).isStarted).toBe(true);
        });

        test('should not allow multiple starts', async () => {
            await runner.start();
            (runner as any).events.emit('started');

            await expect(runner.start()).rejects.toThrow('Runner already started');
        });

        it('should handle health check failures with retries', async () => {
            // Setup fake timers
            jest.useFakeTimers();

            // Constants that should match the implementation in runner.ts
            const RETRY_DELAY = 1000; // 1 second retry delay from startHealthCheck

            const mockClient: MockClient = {
                health: jest.fn<() => Promise<HandlerHealthResponse>>().mockResolvedValue(
                    { error: '' } as HandlerHealthResponse
                ),
                isActiveRunner: jest.fn<() => Promise<IsActiveRunnerResponse>>()
                    .mockResolvedValueOnce({ isActive: false, error: 'test error' } as IsActiveRunnerResponse)
                    .mockResolvedValueOnce({ isActive: false, error: 'test error' } as IsActiveRunnerResponse)
                    .mockResolvedValue({ isActive: true, error: '' } as IsActiveRunnerResponse),
                print: jest.fn<() => Promise<PrintResponse>>().mockResolvedValue(
                    { error: '' } as PrintResponse
                ),
                activity: jest.fn(),
                done: jest.fn(),
                log: jest.fn(),
                sleep: jest.fn(),
                subscribe: jest.fn(),
                nextEvent: jest.fn(),
                unsubscribe: jest.fn(),
                startSession: jest.fn(),
                encodeJWT: jest.fn(),
                refreshOAuthToken: jest.fn(),
            };

            const runner = new Runner('test-id', '/test/dir', mockClient as any);

            // Start the runner
            const startPromise = runner.start();

            // Emit the started event to complete the start process
            (runner as any).events.emit('started');
            await startPromise;

            // Helper function to advance timers and process promises
            const runPromiseWithFakeTimers = async () => {
                // Allow any pending promises to resolve
                await Promise.resolve();
                // Fast forward past the retry delay
                jest.advanceTimersByTime(RETRY_DELAY);
                // Allow any promises that were queued by the timer to resolve
                await Promise.resolve();
            };

            // Process each health check attempt (two failures followed by success)
            await runPromiseWithFakeTimers(); // First check (fails)
            await runPromiseWithFakeTimers(); // Second check (fails)
            await runPromiseWithFakeTimers(); // Third check (succeeds)

            // Verify isActiveRunner was called the expected number of times
            expect(mockClient.isActiveRunner).toHaveBeenCalledTimes(3);

            // Clean up
            runner.stop();
            jest.useRealTimers();
        });
    });

    describe('stop', () => {
        it('should clear health check timer and emit stop event', async () => {
            const mockClient: MockClient = {
                health: jest.fn<() => Promise<HandlerHealthResponse>>().mockResolvedValue(
                    { error: '' } as HandlerHealthResponse
                ),
                isActiveRunner: jest.fn<() => Promise<IsActiveRunnerResponse>>()
                    .mockResolvedValue({ isActive: true, error: '' } as IsActiveRunnerResponse),
                print: jest.fn<() => Promise<PrintResponse>>().mockResolvedValue(
                    { error: '' } as PrintResponse
                ),
                activity: jest.fn(),
                done: jest.fn(),
                log: jest.fn(),
                sleep: jest.fn(),
                subscribe: jest.fn(),
                nextEvent: jest.fn(),
                unsubscribe: jest.fn(),
                startSession: jest.fn(),
                encodeJWT: jest.fn(),
                refreshOAuthToken: jest.fn(),
            };

            const runner = new Runner('test-id', '/test/dir', mockClient as any);
            await runner.start();

            // Wait for health check to start
            await new Promise(resolve => setTimeout(resolve, 100));

            runner.stop();

            // @ts-ignore - accessing private property for testing
            expect(runner['healthcheckTimer']._destroyed).toBe(true);
        });
    });

    describe('akPrint', () => {
        it('should handle print failures gracefully', async () => {
            const mockClient: MockClient = {
                health: jest.fn<() => Promise<HandlerHealthResponse>>().mockResolvedValue(
                    { error: '' } as HandlerHealthResponse
                ),
                isActiveRunner: jest.fn<() => Promise<IsActiveRunnerResponse>>()
                    .mockResolvedValue({ isActive: true, error: '' } as IsActiveRunnerResponse),
                print: jest.fn<() => Promise<PrintResponse>>().mockRejectedValue(new Error('print failed')),
                activity: jest.fn(),
                done: jest.fn(),
                log: jest.fn(),
                sleep: jest.fn(),
                subscribe: jest.fn(),
                nextEvent: jest.fn(),
                unsubscribe: jest.fn(),
                startSession: jest.fn(),
                encodeJWT: jest.fn(),
                refreshOAuthToken: jest.fn(),
            };

            const runner = new Runner('test-id', '/test/dir', mockClient as any);

            // Save original console.log before it's modified by Runner
            const originalConsoleLog = console.log;

            // Create spy after runner is instantiated since runner modifies console.log
            const consoleSpy = jest.spyOn(runner as any, 'originalConsoleLog');

            // Call the private akPrint method
            await (runner as any).akPrint('test message');

            // Verify error was logged correctly
            expect(consoleSpy).toHaveBeenCalledWith('Failed to send print message:', expect.any(Error));

            // Restore console.log
            consoleSpy.mockRestore();
            console.log = originalConsoleLog;
        });
    });

    describe('graceful shutdown', () => {
        test('should handle shutdown process correctly', async () => {
            const processExitSpy = jest.spyOn(process, 'exit').mockImplementation(() => undefined as never);

            await runner.start();
            (runner as any).events.emit('started');

            await (runner as any).gracefulShutdown();

            expect(processExitSpy).toHaveBeenCalledWith(0);

            processExitSpy.mockRestore();
        });

        test('should handle shutdown errors', async () => {
            const processExitSpy = jest.spyOn(process, 'exit').mockImplementation(() => undefined as never);
            const error = new Error('Shutdown error');

            await runner.start();
            (runner as any).events.emit('started');

            // Mock stop to throw an error
            jest.spyOn(runner, 'stop').mockImplementation(() => {
                throw error;
            });

            await (runner as any).gracefulShutdown();

            expect(console.error).toHaveBeenCalledWith('Error during shutdown:', error);
            expect(processExitSpy).toHaveBeenCalledWith(1);

            processExitSpy.mockRestore();
        });
    });
});
