import {describe, expect, test, jest, beforeEach} from '@jest/globals';
import {SysCalls, SyscallError} from '../runtime/runner/syscalls';

// Type for mocking the client
interface MockClient {
    subscribe: jest.Mock;
    nextEvent: jest.Mock;
    unsubscribe: jest.Mock;
}

describe('SysCalls', () => {
    let sysCalls: SysCalls;
    let mockClient: MockClient;
    const testRunnerId = 'test-runner-id';

    beforeEach(() => {
        // Create mock client
        mockClient = {
            subscribe: jest.fn(),
            nextEvent: jest.fn(),
            unsubscribe: jest.fn()
        };

        // We need to cast here because our mock doesn't fully implement the client interface
        // @ts-expect-error Mock client doesn't implement all methods
        sysCalls = new SysCalls(testRunnerId,mockClient);
    });

    describe('subscribe', () => {
        test('should successfully subscribe to events', async () => {
            const expectedSignalId = 'test-signal-id';
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.subscribe.mockResolvedValue({
                signalId: expectedSignalId,
                error: '',
                $typeName: 'autokitteh.user_code.v1.SubscribeResponse'
            });

            const result = await sysCalls.subscribe('test-source');

            expect(mockClient.subscribe).toHaveBeenCalledWith({
                runnerId: testRunnerId,
                connection: 'test-source',
                filter: ''
            });
            expect(result).toBe(expectedSignalId);
        });

        test('should subscribe with filter if provided', async () => {
            const expectedSignalId = 'test-signal-id';
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.subscribe.mockResolvedValue({
                signalId: expectedSignalId,
                error: '',
                $typeName: 'autokitteh.user_code.v1.SubscribeResponse'
            });

            const result = await sysCalls.subscribe('test-source', 'event.type == "test"');

            expect(mockClient.subscribe).toHaveBeenCalledWith({
                runnerId: testRunnerId,
                connection: 'test-source',
                filter: 'event.type == "test"'
            });
            expect(result).toBe(expectedSignalId);
        });

        test('should throw error if no source provided', async () => {
            await expect(sysCalls.subscribe('')).rejects.toThrow(Error);
        });

        test('should throw error if server returns error', async () => {
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.subscribe.mockResolvedValue({
                signalId: '',
                error: 'Server error',
                $typeName: 'autokitteh.user_code.v1.SubscribeResponse'
            });

            await expect(sysCalls.subscribe('test-source')).rejects.toThrow(SyscallError);
        });
    });

    describe('nextEvent', () => {
        test('should retrieve the next event from a subscription', async () => {
            const testData = {message: 'test event data'};
            const encodedData = new TextEncoder().encode(JSON.stringify(testData));

            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.nextEvent.mockResolvedValue({
                event: {
                    data: encodedData
                },
                error: '',
                $typeName: 'autokitteh.user_code.v1.NextEventResponse'
            });

            const result = await sysCalls.nextEvent('test-signal-id');

            expect(mockClient.nextEvent).toHaveBeenCalledWith({
                runnerId: testRunnerId,
                signalIds: ['test-signal-id'],
                timeoutMs: BigInt(0)
            });
            expect(result).toEqual(testData);
        });

        test('should handle array of subscription IDs', async () => {
            const testData = {message: 'test event data'};
            const encodedData = new TextEncoder().encode(JSON.stringify(testData));

            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.nextEvent.mockResolvedValue({
                event: {
                    data: encodedData
                },
                error: '',
                $typeName: 'autokitteh.user_code.v1.NextEventResponse'
            });

            await sysCalls.nextEvent(['signal-1', 'signal-2']);

            expect(mockClient.nextEvent).toHaveBeenCalledWith({
                runnerId: testRunnerId,
                signalIds: ['signal-1', 'signal-2'],
                timeoutMs: BigInt(0)
            });
        });

        test('should handle timeout option', async () => {
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.nextEvent.mockResolvedValue({
                event: {
                    data: new TextEncoder().encode('{}')
                },
                error: '',
                $typeName: 'autokitteh.user_code.v1.NextEventResponse'
            });

            await sysCalls.nextEvent('test-signal-id', {timeout: 5});

            expect(mockClient.nextEvent).toHaveBeenCalledWith({
                runnerId: testRunnerId,
                signalIds: ['test-signal-id'],
                timeoutMs: BigInt(5000)
            });
        });

        test('should handle timeout as an object', async () => {
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.nextEvent.mockResolvedValue({
                event: {
                    data: new TextEncoder().encode('{}')
                },
                error: '',
                $typeName: 'autokitteh.user_code.v1.NextEventResponse'
            });

            await sysCalls.nextEvent('test-signal-id', {timeout: {seconds: 10}});

            expect(mockClient.nextEvent).toHaveBeenCalledWith({
                runnerId: testRunnerId,
                signalIds: ['test-signal-id'],
                timeoutMs: BigInt(10000)
            });
        });

        test('should return null if no event data', async () => {
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.nextEvent.mockResolvedValue({
                event: {},
                error: '',
                $typeName: 'autokitteh.user_code.v1.NextEventResponse'
            });

            const result = await sysCalls.nextEvent('test-signal-id');
            expect(result).toBeNull();
        });

        test('should throw error if server returns error', async () => {
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.nextEvent.mockResolvedValue({
                error: 'Server error',
                $typeName: 'autokitteh.user_code.v1.NextEventResponse'
            });

            await expect(sysCalls.nextEvent('test-signal-id')).rejects.toThrow(SyscallError);
        });
    });

    describe('unsubscribe', () => {
        test('should successfully unsubscribe', async () => {
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.unsubscribe.mockResolvedValue({
                error: '',
                $typeName: 'autokitteh.user_code.v1.UnsubscribeResponse'
            });

            await sysCalls.unsubscribe('test-signal-id');

            expect(mockClient.unsubscribe).toHaveBeenCalledWith({
                runnerId: testRunnerId,
                signalId: 'test-signal-id'
            });
        });

        test('should throw error if server returns error', async () => {
            // @ts-expect-error TypeScript doesn't understand jest mocks properly
            mockClient.unsubscribe.mockResolvedValue({
                error: 'Server error',
                $typeName: 'autokitteh.user_code.v1.UnsubscribeResponse'
            });

            await expect(sysCalls.unsubscribe('test-signal-id')).rejects.toThrow(SyscallError);
        });
    });
});
