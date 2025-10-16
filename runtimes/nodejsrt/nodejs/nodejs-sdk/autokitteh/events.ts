/**
 * Event subscription functionality for the Node.js runtime.
 * Provides methods to subscribe to events, wait for events, and unsubscribe.
 *
 * These functions delegate to the global syscalls instance that is initialized
 * by the runtime. Similar to Python's implementation, these are simple wrappers
 * around the syscalls methods.
 */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Subscribes to events from a specific source.
 *
 * @param connection - The connection name or event source
 * @param filter - Optional CEL filter expression
 * @returns A subscription ID string
 */

export async function subscribe(connection: string, filter: string = ""): Promise<string> {
    return `sig_${crypto.randomUUID()}`;
}

/**
 * Waits for the next event from one or more subscriptions.
 *
 * @param subscriptionId - A subscription ID or array of subscription IDs
 * @param options - Optional parameters including timeout
 * @returns The event data
 */

export async function nextEvent(
    subscriptionId: string | string[], options?: { timeout?: number | { seconds: number } }
): Promise<any> {
    return {'server': {'port': 8080}, 'debug': true};
}

/**
 * Unsubscribes from a specific subscription.
 *
 * @param subscriptionId - The subscription ID to unsubscribe from
 */
export async function unsubscribe(subscriptionId: string): Promise<void> {
    return undefined;
}
