/**
 * Event subscription functionality for the Node.js runtime.
 * Provides methods to subscribe to events, wait for events, and unsubscribe.
 */

import { HandlerService } from "../../pb/autokitteh/user_code/v1/handler_svc_pb";
import { createClient } from "@connectrpc/connect";

/**
 * Error thrown when an event subscription operation fails.
 */
export class EventSubscriptionError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'EventSubscriptionError';
  }
}

/**
 * Converts timeout to milliseconds
 * @param timeout Timeout in seconds or as an object with a seconds property
 * @returns Timeout in milliseconds
 */
function timeoutToMs(timeout: number | { seconds: number }): number {
  if (!timeout) return 0;

  if (typeof timeout === 'number') {
    return Math.floor(timeout * 1000);
  }

  if (typeof timeout === 'object' && 'seconds' in timeout) {
    return Math.floor(timeout.seconds * 1000);
  }

  throw new TypeError(`timeout ${timeout} should be a number of seconds or an object with a seconds property`);
}

/**
 * Type for the Handler service client
 */
type HandlerClient = ReturnType<typeof createClient<typeof HandlerService>>;

/**
 * Handles event subscriptions by communicating with the Autokitteh framework.
 */
export class EventSubscriber {
  /**
   * Creates a new EventSubscriber.
   *
   * @param client - The client for communicating with the handler service.
   * @param runnerId - The runner ID for this runtime instance.
   */
  constructor(
    private readonly client: HandlerClient,
    private readonly runnerId: string
  ) {}

  /**
   * Subscribes to events from a specific source with an optional filter.
   *
   * @param source - The connection name or event source.
   * @param filter - Optional CEL filter expression.
   * @returns A subscription ID string.
   * @throws EventSubscriptionError if the subscription fails.
   */
  async subscribe(source: string, filter: string = ""): Promise<string> {
    console.log(`EventSubscriber.subscribe: ${source} ${filter}`);

    if (!source) {
      throw new EventSubscriptionError("Missing source parameter");
    }

    try {
      const response = await this.client.subscribe({
        runnerId: this.runnerId,
        connection: source,
        filter: filter
      });

      if (response.error) {
        throw new EventSubscriptionError(`Subscribe failed: ${response.error}`);
      }

      return response.signalId;
    } catch (error) {
      if (error instanceof EventSubscriptionError) {
        throw error;
      }
      throw new EventSubscriptionError(`Subscribe failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Waits for the next event from one or more subscriptions.
   *
   * @param subscriptionId - A subscription ID string or array of subscription IDs.
   * @param options - Optional parameters including timeout.
   * @returns The event data.
   * @throws EventSubscriptionError if the operation fails.
   */
  async nextEvent(
    subscriptionId: string | string[],
    options?: { timeout?: number | { seconds: number } }
  ): Promise<any> {
    const ids = Array.isArray(subscriptionId) ? subscriptionId : [subscriptionId];
    const timeoutMs = options?.timeout ? timeoutToMs(options.timeout) : 0;

    console.log(`EventSubscriber.nextEvent: ${ids} (timeout: ${timeoutMs}ms)`);

    try {
      const response = await this.client.nextEvent({
        runnerId: this.runnerId,
        signalIds: ids,
        timeoutMs: BigInt(timeoutMs)
      });

      if (response.error) {
        throw new EventSubscriptionError(`NextEvent failed: ${response.error}`);
      }

      if (!response.event?.data) {
        return null;
      }

      try {
        // Parse the event data from the buffer
        const dataText = new TextDecoder().decode(response.event.data);
        const data = JSON.parse(dataText);

        // Return as an object with attribute-style access if it's an object
        return typeof data === 'object' ? Object.assign(Object.create(null), data) : data;
      } catch (error) {
        throw new EventSubscriptionError(`Invalid event data: ${error instanceof Error ? error.message : String(error)}`);
      }
    } catch (error) {
      if (error instanceof EventSubscriptionError) {
        throw error;
      }
      throw new EventSubscriptionError(`NextEvent failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Unsubscribes from a specific subscription.
   *
   * @param subscriptionId - The subscription ID to unsubscribe from.
   * @throws EventSubscriptionError if the unsubscribe operation fails.
   */
  async unsubscribe(subscriptionId: string): Promise<void> {
    console.log(`EventSubscriber.unsubscribe: ${subscriptionId}`);

    try {
      const response = await this.client.unsubscribe({
        runnerId: this.runnerId,
        signalId: subscriptionId
      });

      if (response.error) {
        throw new EventSubscriptionError(`Unsubscribe failed: ${response.error}`);
      }
    } catch (error) {
      if (error instanceof EventSubscriptionError) {
        throw error;
      }
      throw new EventSubscriptionError(`Unsubscribe failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }
}
