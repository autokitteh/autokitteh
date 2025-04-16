import { DirectHandlerClient } from './direct_client';
import { TextDecoder } from 'util';

describe('DirectHandlerClient', () => {
    describe('nextEvent', () => {
        it('should return an error for non-subscribed signal IDs', async () => {
            // Arrange
            const client = new DirectHandlerClient();

            // Act - calling nextEvent without subscribing first
            const result = await client.nextEvent({ signalIds: ['non-existent-signal'] });

            // Assert
            expect(result.error).toContain('No valid subscription found');
            expect(result.event).toBeNull();
        });

        it('should return an event for a subscribed signal ID', async () => {
            // Arrange
            const client = new DirectHandlerClient();

            // Subscribe first to get a valid signal ID
            const subscribeResult = await client.subscribe({});
            expect(subscribeResult.error).toBe('');
            expect(subscribeResult.signalId).toBeDefined();

            // Act - call nextEvent with the valid signal ID
            const result = await client.nextEvent({ signalIds: [subscribeResult.signalId] });

            // Assert
            expect(result.error).toBe('');
            expect(result.event).toBeDefined();
            expect(result.event?.data).toBeDefined();

            // Verify data is binary buffer (Uint8Array)
            expect(result.event?.data instanceof Uint8Array).toBe(true);

            // Decode and parse the event
            const decoder = new TextDecoder();
            const eventJson = decoder.decode(result.event?.data);
            const eventData = JSON.parse(eventJson);

            // Verify event structure
            expect(eventData.id).toBeDefined();
            expect(eventData.type).toBe('test-event');
            expect(eventData.source).toBe('direct-execution');
            expect(eventData.data).toBeDefined();
            expect(eventData.data.message).toBe('This is a test event');
        });

        it('should handle unsubscribe and no longer return events', async () => {
            // Arrange
            const client = new DirectHandlerClient();

            // Subscribe first to get a valid signal ID
            const subscribeResult = await client.subscribe({});
            const signalId = subscribeResult.signalId;

            // Unsubscribe
            const unsubscribeResult = await client.unsubscribe({ signalId });
            expect(unsubscribeResult.error).toBe('');

            // Act - call nextEvent with the now-unsubscribed signal ID
            const result = await client.nextEvent({ signalIds: [signalId] });

            // Assert
            expect(result.error).toContain('No valid subscription found');
            expect(result.event).toBeNull();
        });
    });
});
