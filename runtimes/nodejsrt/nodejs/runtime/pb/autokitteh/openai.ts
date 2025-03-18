/**
 * Initialize an OpenAI client, based on an AutoKitteh connection.
 */

import { OpenAI } from 'openai';
import { checkConnectionName } from './connections';
import { ConnectionInitError } from './errors';

// Define the shape of options we support
interface OpenAIClientOptions {
  organization?: string;
  apiKey?: string; 
  baseURL?: string;
  timeout?: number;
  maxRetries?: number;
  defaultHeaders?: Record<string, string>;
  defaultQuery?: Record<string, string>;
}

/**
 * Initialize an OpenAI client, based on an AutoKitteh connection.
 * 
 * API reference:
 * https://platform.openai.com/docs/api-reference/
 * https://github.com/openai/openai-node
 * 
 * @param connection - AutoKitteh connection name.
 * @param options - Additional OpenAI client options.
 * @returns OpenAI API client.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OpenAIError from the OpenAI library if connection attempt fails or is unauthorized.
 */
export function openaiClient(connection: string, options: Partial<OpenAIClientOptions> = {}): OpenAI {
  try {
    // Validate connection name
    checkConnectionName(connection);

    // Get API key from environment variables
    const apiKey = process.env[`${connection}__apiKey`];

    // Check if API key is available
    if (!apiKey) {
      throw new ConnectionInitError(connection);
    }

    // Optional: Get organization ID if available
    const organization = process.env[`${connection}__organization`];

    // Create and return the OpenAI client
    return new OpenAI({ 
      apiKey,
      organization,
      ...options
    });
  } catch (error) {
    // If error is already a ConnectionInitError, rethrow it
    if (error instanceof ConnectionInitError) {
      throw error;
    }
    
    // For other errors, provide context about the connection
    throw new Error(`Failed to initialize OpenAI client for connection '${connection}': ${error instanceof Error ? error.message : String(error)}`);
  }
} 