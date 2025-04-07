/**
 * System calls for the Node.js runtime.
 * Similar to the Python runtime's syscalls.py, this provides implementations
 * for functions that need to interact with the Autokitteh framework.
 */

import { Client } from "@connectrpc/connect";
import { HandlerService } from "../pb/autokitteh/user_code/v1/handler_svc_pb";

/**
 * Error thrown when a system call fails.
 */
export class SyscallError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'SyscallError';
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
 * Wrapper for system calls that communicate with the Autokitteh framework.
 */
export class SysCalls {
  private runnerId: string;
  private client: Client<typeof HandlerService>;

  /**
   * Creates a new SysCalls instance.
   *
   * @param runnerId - The ID of the runner.
   * @param client - The gRPC client for communicating with the handler service.
   */
  constructor(runnerId: string, client: Client<typeof HandlerService>) {
    this.runnerId = runnerId;
    this.client = client;
  }

  /**
   * Refreshes an OAuth token for a connection.
   *
   * @param integration - The integration name.
   * @param connection - The connection name.
   * @returns A tuple containing the new access token and expiry date.
   * @throws SyscallError if the token refresh fails.
   */
  async akRefreshOAuth(integration: string, connection: string): Promise<[string, Date]> {
    try {
      const response = await this.client.refreshOAuthToken({
        runnerId: this.runnerId,
        integration,
        connection
      });

      if (response.error) {
        throw new SyscallError(`refresh_oauth: ${response.error}`);
      }

      const token = response.token;
      const expiryDate = response.expires ? new Date(Number(response.expires.seconds) * 1000) : new Date();

      return [token, expiryDate];
    } catch (error) {
      if (error instanceof SyscallError) {
        throw error;
      }
      throw new SyscallError(`refresh_oauth failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Encodes a JWT with the given payload.
   *
   * @param payload - The JWT payload.
   * @param connection - The connection name.
   * @param algorithm - The signing algorithm.
   * @returns A JWT token.
   * @throws SyscallError if the JWT encoding fails.
   */
  akEncodeJwt(payload: Record<string, number>, connection: string, algorithm: string): string {
    // Since the signature requires a synchronous return, we'll provide a fallback
    // that makes it clear this should be replaced with actual implementation
    console.warn('Synchronous JWT encoding is not implemented. This is a placeholder.');
    return 'DUMMY JWT';
  }

  /**
   * Subscribes to events from a specific source with an optional filter.
   *
   * @param source - The connection name or event source.
   * @param filter - Optional CEL filter expression.
   * @returns A subscription ID string.
   * @throws SyscallError if the subscription fails.
   */
  async subscribe(source: string, filter: string = ""): Promise<string> {
    console.log(`subscribe: ${source} ${filter}`);

    if (!source) {
      throw new Error("missing source");
    }

    try {
      const response = await this.client.subscribe({
        runnerId: this.runnerId,
        connection: source,
        filter: filter
      });

      if (response.error) {
        throw new SyscallError(`subscribe: ${response.error}`);
      }

      return response.signalId;
    } catch (error) {
      if (error instanceof SyscallError) {
        throw error;
      }
      throw new SyscallError(`subscribe failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Waits for the next event from one or more subscriptions.
   *
   * @param subscriptionId - A subscription ID string or array of subscription IDs.
   * @param options - Optional parameters including timeout.
   * @returns The event data.
   * @throws SyscallError if the operation fails.
   */
  async nextEvent(
    subscriptionId: string | string[],
    options?: { timeout?: number | { seconds: number } }
  ): Promise<any> {
    const ids = Array.isArray(subscriptionId) ? subscriptionId : [subscriptionId];
    const timeoutMs = options?.timeout ? timeoutToMs(options.timeout) : 0;

    console.log(`nextEvent: ${ids} (timeout: ${timeoutMs}ms)`);

    try {
      const response = await this.client.nextEvent({
        runnerId: this.runnerId,
        signalIds: ids,
        timeoutMs: BigInt(timeoutMs)
      });

      if (response.error) {
        throw new SyscallError(`next_event: ${response.error}`);
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
        throw new SyscallError(`next_event: invalid event data: ${error instanceof Error ? error.message : String(error)}`);
      }
    } catch (error) {
      if (error instanceof SyscallError) {
        throw error;
      }
      throw new SyscallError(`next_event failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Unsubscribes from a specific subscription.
   *
   * @param subscriptionId - The subscription ID to unsubscribe from.
   * @throws SyscallError if the unsubscribe operation fails.
   */
  async unsubscribe(subscriptionId: string): Promise<void> {
    console.log(`unsubscribe: ${subscriptionId}`);

    try {
      const response = await this.client.unsubscribe({
        runnerId: this.runnerId,
        signalId: subscriptionId
      });

      if (response.error) {
        throw new SyscallError(`unsubscribe: ${response.error}`);
      }
    } catch (error) {
      if (error instanceof SyscallError) {
        throw error;
      }
      throw new SyscallError(`unsubscribe failed: ${error instanceof Error ? error.message : String(error)}`);
    }
  }
}
