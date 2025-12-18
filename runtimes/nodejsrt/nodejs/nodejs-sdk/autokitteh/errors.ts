/**
 * AutoKitteh SDK errors.
 */

/**
 * Generic base class for all errors in the AutoKitteh SDK.
 */
export class AutoKittehError extends Error {
  constructor(...args: any[]) {
    super(args.join(' '));
    this.name = 'AutoKittehError';
  }
}

/**
 * A required AutoKitteh connection was not initialized yet.
 */
export class ConnectionInitError extends AutoKittehError {
  constructor(connection: string) {
    super(`AutoKitteh connection '${connection}' not initialized`);
    this.name = 'ConnectionInitError';
  }
}

/**
 * A required environment variable is missing or invalid.
 */
export class EnvVarError extends AutoKittehError {
  constructor(envVar: string, desc: string) {
    super(`Environment variable '${envVar}' is ${desc}`);
    this.name = 'EnvVarError';
  }
}

/**
 * OAuth token refresh failed.
 */
export class OAuthRefreshError extends AutoKittehError {
  constructor(connection: string, error?: Error) {
    super(`OAuth refresh failed for '${connection}' connection: ${error?.message || 'Unknown error'}`);
    this.name = 'OAuthRefreshError';
  }
}

/**
 * API calls not supported by OAuth-based Atlassian connections.
 */
export class AtlassianOAuthError extends AutoKittehError {
  constructor(connection: string) {
    super(
      `API calls not supported by '${connection}', use a token-based connection instead`
    );
    this.name = 'AtlassianOAuthError';
  }
} 