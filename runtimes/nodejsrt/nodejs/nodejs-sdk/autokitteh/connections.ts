/**
 * AutoKitteh connection-related utilities.
 */

/**
 * Validates an AutoKitteh connection name.
 *
 * @param connection - The connection name to validate
 * @throws Error if the connection name is invalid
 */
export function checkConnectionName(connection: string): void {
  if (!/^[A-Za-z_]\w*$/.test(connection)) {
    throw new Error(`Invalid AutoKitteh connection name: '${connection}'`);
  }
}

/**
 * Object that holds functions that can be patched at runtime
 */
export const connectionUtils = {
  /**
   * Generates a JWT with the given payload.
   * This is a mock function that will be overridden by the AutoKitteh runner.
   *
   * @param payload - The JWT payload
   * @param connection - The connection name
   * @param algorithm - The signing algorithm
   * @returns A JWT token
   */
  encodeJwt: function(payload: Record<string, number>, connection: string, algorithm: string): string {
    console.warn('!!!!!!!!!! SDK\'s encodeJwt() not overridden !!!!!!!!!!');
    return 'DUMMY JWT';
  },

  /**
   * Refreshes an OAuth token for a connection.
   * This is a mock function that will be overridden by the AutoKitteh runner.
   *
   * @param integration - The integration name
   * @param connection - The connection name
   * @returns A tuple containing the new access token and expiry date
   */
  refreshOAuth: async function(integration: string, connection: string): Promise<[string, Date]> {
    console.warn('!!!!!!!!!! SDK\'s refreshOAuth() not overridden !!!!!!!!!!');
    return ['DUMMY TOKEN', new Date()];
  }
};
