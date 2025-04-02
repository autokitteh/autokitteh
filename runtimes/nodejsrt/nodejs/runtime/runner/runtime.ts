import { ak_call, Waiter } from './ak_call';
import { EventSubscriber } from '../nodejs-sdk/autokitteh/events';

/**
 * Initializes global objects and functions that will be available to user code.
 *
 * @param waiter - The activity waiter instance
 * @param eventSubscriber - The event subscriber instance
 */
export function initializeGlobals(waiter: Waiter, eventSubscriber: EventSubscriber) {
    // Initialize ak_call
    Object.defineProperty(global, 'ak_call', {
        value: ak_call(waiter),
        configurable: false,
        writable: false
    });

    // Initialize the autokitteh global object
    const autokitteh = {
        /**
         * Subscribes to events from a specific source.
         *
         * @param connection - The connection name or event source
         * @param filter - CEL filter expression
         * @returns A subscription ID string
         */
        subscribe: (connection: string, filter: string = "") => {
            return eventSubscriber.subscribe(connection, filter);
        },

        /**
         * Waits for the next event from one or more subscriptions.
         *
         * @param subscriptionId - A subscription ID or array of subscription IDs
         * @param options - Optional parameters including timeout
         * @returns The event data
         */
        nextEvent: (
            subscriptionId: string | string[],
            options?: { timeout?: number | { seconds: number } }
        ) => {
            return eventSubscriber.nextEvent(subscriptionId, options);
        },

        /**
         * Unsubscribes from a specific subscription.
         *
         * @param subscriptionId - The subscription ID to unsubscribe from
         */
        unsubscribe: (subscriptionId: string) => {
            return eventSubscriber.unsubscribe(subscriptionId);
        }
    };

    // Set the global autokitteh object
    Object.defineProperty(global, 'autokitteh', {
        value: autokitteh,
        configurable: false,
        writable: false
    });
}
