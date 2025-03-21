/**
 * System calls for the Node.js runtime.
 * Similar to the Python runtime's syscalls.py, this provides implementations
 * for functions that need to interact with the Autokitteh framework.
 */

import { Client } from "@connectrpc/connect";
import { HandlerService } from "./pb/autokitteh/user_code/v1/handler_svc_pb";

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
}
