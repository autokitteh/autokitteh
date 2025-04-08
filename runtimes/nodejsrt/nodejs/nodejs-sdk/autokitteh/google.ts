/**
 * Initialize Google API clients, based on AutoKitteh connections.
 */

import {checkConnectionName} from './connections';
import {connectionUtils} from './connections';
import {ConnectionInitError, OAuthRefreshError} from './errors';

// Using any types to avoid complex type issues with the Google API libraries
type OAuth2Client = any;
type JWT = any;

// Common API client options
interface GoogleClientOptions {
    [key: string]: any;
}

/**
 * Base structure for creating Google API clients
 */
function createGoogleClient(
    integration: string,
    connection: string,
    scopes: string[],
    options: GoogleClientOptions = {}
): any {
    try {
        const auth = googleAuth(integration, connection, scopes, options);

        // In a real implementation, this would return the actual Google API client
        // This is a placeholder that follows the same interface pattern
        return {
            auth,
            _integration: integration,
            _connection: connection,
            _options: options
        };
    } catch (error: any) {
        console.error(`Error creating ${integration} client:`, error);
        throw error;
    }
}

/**
 * Initialize a Gmail client, based on an AutoKitteh connection.
 *
 * API documentation:
 * https://docs.autokitteh.com/integrations/google/gmail/nodejs
 *
 * @param connection - AutoKitteh connection name.
 * @param options - Additional options to pass to the Gmail client.
 * @returns Gmail client.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OAuthRefreshError if OAuth token refresh failed.
 */
export function gmailClient(connection: string, options: GoogleClientOptions = {}) {
    const defaultScopes = [
        'https://www.googleapis.com/auth/gmail.modify',
        'https://www.googleapis.com/auth/gmail.settings.basic',
    ];

    const client = createGoogleClient('gmail', connection, defaultScopes, options);

    // Return a placeholder that mimics the Gmail API structure
    return {
        // Core properties
        ...client,

        // Users resource
        users: {
            // Messages resource
            messages: {
                list: async (params: any) => {
                    console.log('Gmail API - listing messages', params);
                    throw new Error('Gmail API not yet implemented');
                },
                get: async (params: any) => {
                    console.log('Gmail API - getting message', params);
                    throw new Error('Gmail API not yet implemented');
                },
                send: async (params: any) => {
                    console.log('Gmail API - sending message', params);
                    throw new Error('Gmail API not yet implemented');
                }
            },
            // Labels resource
            labels: {
                list: async (params: any) => {
                    console.log('Gmail API - listing labels', params);
                    throw new Error('Gmail API not yet implemented');
                }
            },
            // Thread resource
            threads: {
                list: async (params: any) => {
                    console.log('Gmail API - listing threads', params);
                    throw new Error('Gmail API not yet implemented');
                }
            }
        }
    };
}

/**
 * Initialize a Google Calendar client, based on an AutoKitteh connection.
 *
 * API documentation:
 * https://docs.autokitteh.com/integrations/google/calendar/nodejs
 *
 * @param connection - AutoKitteh connection name.
 * @param options - Additional options to pass to the Calendar client.
 * @returns Google Calendar client.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OAuthRefreshError if OAuth token refresh failed.
 */
export function googleCalendarClient(connection: string, options: GoogleClientOptions = {}) {
    const defaultScopes = [
        'https://www.googleapis.com/auth/calendar',
        'https://www.googleapis.com/auth/calendar.events',
    ];

    const client = createGoogleClient('googlecalendar', connection, defaultScopes, options);

    // Return a placeholder that mimics the Calendar API structure
    return {
        // Core properties
        ...client,

        // Calendar resources
        calendars: {
            get: async (params: any) => {
                console.log('Calendar API - getting calendar', params);
                throw new Error('Calendar API not yet implemented');
            },
            list: async (params: any) => {
                console.log('Calendar API - listing calendars', params);
                throw new Error('Calendar API not yet implemented');
            }
        },
        // Events resource
        events: {
            get: async (params: any) => {
                console.log('Calendar API - getting event', params);
                throw new Error('Calendar API not yet implemented');
            },
            list: async (params: any) => {
                console.log('Calendar API - listing events', params);
                throw new Error('Calendar API not yet implemented');
            },
            insert: async (params: any) => {
                console.log('Calendar API - inserting event', params);
                throw new Error('Calendar API not yet implemented');
            },
            update: async (params: any) => {
                console.log('Calendar API - updating event', params);
                throw new Error('Calendar API not yet implemented');
            },
            delete: async (params: any) => {
                console.log('Calendar API - deleting event', params);
                throw new Error('Calendar API not yet implemented');
            }
        }
    };
}

/**
 * Initialize a Google Drive client, based on an AutoKitteh connection.
 *
 * API documentation:
 * https://docs.autokitteh.com/integrations/google/drive/nodejs
 *
 * @param connection - AutoKitteh connection name.
 * @param options - Additional options to pass to the Drive client.
 * @returns Google Drive client.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OAuthRefreshError if OAuth token refresh failed.
 */
export function googleDriveClient(connection: string, options: GoogleClientOptions = {}) {
    const defaultScopes = [
        'https://www.googleapis.com/auth/drive.file',
    ];

    const client = createGoogleClient('googledrive', connection, defaultScopes, options);

    // Return a placeholder that mimics the Drive API structure
    return {
        // Core properties
        ...client,

        // Files resource
        files: {
            get: async (params: any) => {
                console.log('Drive API - getting file', params);
                throw new Error('Drive API not yet implemented');
            },
            list: async (params: any) => {
                console.log('Drive API - listing files', params);
                throw new Error('Drive API not yet implemented');
            },
            create: async (params: any) => {
                console.log('Drive API - creating file', params);
                throw new Error('Drive API not yet implemented');
            },
            update: async (params: any) => {
                console.log('Drive API - updating file', params);
                throw new Error('Drive API not yet implemented');
            },
            delete: async (params: any) => {
                console.log('Drive API - deleting file', params);
                throw new Error('Drive API not yet implemented');
            }
        }
    };
}

/**
 * Initialize a Google Forms client, based on an AutoKitteh connection.
 *
 * API documentation:
 * https://docs.autokitteh.com/integrations/google/forms/nodejs
 *
 * @param connection - AutoKitteh connection name.
 * @param options - Additional options to pass to the Forms client.
 * @returns Google Forms client.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OAuthRefreshError if OAuth token refresh failed.
 */
export function googleFormsClient(connection: string, options: GoogleClientOptions = {}) {
    const defaultScopes = [
        'https://www.googleapis.com/auth/forms.body',
        'https://www.googleapis.com/auth/forms.responses.readonly',
    ];

    const client = createGoogleClient('googleforms', connection, defaultScopes, options);

    // Return a placeholder that mimics the Forms API structure
    return {
        // Core properties
        ...client,

        // Forms resource
        forms: {
            get: async (params: any) => {
                console.log('Forms API - getting form', params);
                throw new Error('Forms API not yet implemented');
            },
            create: async (params: any) => {
                console.log('Forms API - creating form', params);
                throw new Error('Forms API not yet implemented');
            },
            update: async (params: any) => {
                console.log('Forms API - updating form', params);
                throw new Error('Forms API not yet implemented');
            }
        },
        // Responses resource
        responses: {
            list: async (params: any) => {
                console.log('Forms API - listing responses', params);
                throw new Error('Forms API not yet implemented');
            },
            get: async (params: any) => {
                console.log('Forms API - getting response', params);
                throw new Error('Forms API not yet implemented');
            }
        }
    };
}

/**
 * Initialize a Google Sheets client, based on an AutoKitteh connection.
 *
 * API documentation:
 * https://docs.autokitteh.com/integrations/google/sheets/nodejs
 *
 * @param connection - AutoKitteh connection name.
 * @param options - Additional options to pass to the Sheets client.
 * @returns Google Sheets client.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OAuthRefreshError if OAuth token refresh failed.
 */
export function googleSheetsClient(connection: string, options: GoogleClientOptions = {}) {
    const defaultScopes = ['https://www.googleapis.com/auth/spreadsheets'];

    const client = createGoogleClient('googlesheets', connection, defaultScopes, options);

    // Return a placeholder that mimics the Sheets API structure
    return {
        // Core properties
        ...client,

        // Spreadsheets resource
        spreadsheets: {
            get: async (params: any) => {
                console.log('Sheets API - getting spreadsheet', params);
                throw new Error('Sheets API not yet implemented');
            },
            create: async (params: any) => {
                console.log('Sheets API - creating spreadsheet', params);
                throw new Error('Sheets API not yet implemented');
            },
            // Values sub-resource
            values: {
                get: async (params: any) => {
                    console.log('Sheets API - getting values', params);
                    throw new Error('Sheets API not yet implemented');
                },
                update: async (params: any) => {
                    console.log('Sheets API - updating values', params);
                    throw new Error('Sheets API not yet implemented');
                },
                append: async (params: any) => {
                    console.log('Sheets API - appending values', params);
                    throw new Error('Sheets API not yet implemented');
                },
                clear: async (params: any) => {
                    console.log('Sheets API - clearing values', params);
                    throw new Error('Sheets API not yet implemented');
                }
            },
            // Sheets sub-resource
            sheets: {
                copyTo: async (params: any) => {
                    console.log('Sheets API - copying sheet', params);
                    throw new Error('Sheets API not yet implemented');
                }
            }
        }
    };
}

/**
 * Initialize a Gemini generative AI client, based on an AutoKitteh connection.
 *
 * API reference:
 * - https://ai.google.dev/gemini-api/docs
 *
 * @param connection - AutoKitteh connection name.
 * @param options - Additional options for the Gemini API.
 * @returns A Gemini API client instance.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 */
export function geminiClient(connection: string, options: Record<string, any> = {}) {
    checkConnectionName(connection);

    // Set the API key, if possible.
    const apiKey = process.env[`${connection}__api_key`];
    if (!apiKey) {
        throw new ConnectionInitError(connection);
    }

    // Return a placeholder that mimics the Gemini API structure
    return {
        // Core methods
        generateContent: async (prompt: string | any) => {
            console.log('Gemini API - generating content', typeof prompt === 'string' ? prompt : JSON.stringify(prompt));
            throw new Error('Gemini API not yet implemented in Node.js');
        },

        // Model configuration
        safety: {
            setSettings: (settings: any) => {
                console.log('Gemini API - setting safety settings', settings);
                throw new Error('Gemini API not yet implemented in Node.js');
            }
        },

        // Additional options
        ...options
    };
}

/**
 * Initialize credentials for a Google APIs client.
 *
 * This function supports both AutoKitteh connection modes:
 * users (with OAuth 2.0), and GCP service accounts (with a JSON key).
 *
 * @param integration - AutoKitteh integration name.
 * @param connection - AutoKitteh connection name.
 * @param scopes - List of OAuth permission scopes.
 * @param options - Additional options for the auth client.
 * @returns Google API credentials.
 * @throws Error if the connection name is invalid.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OAuthRefreshError if OAuth token refresh failed.
 */
function googleAuth(integration: string, connection: string, scopes: string[], options: GoogleClientOptions = {}): OAuth2Client | JWT {
    checkConnectionName(connection);

    // Case 1: OAuth authentication
    if (process.env[`${connection}__authType`] === 'oauth') {
        return googleAuthOAuth2(integration, connection, scopes);
    }

    // Case 2: Service Account authentication
    const jsonKey = process.env[`${connection}__JSON`];
    if (jsonKey) {
        try {
            const keyData = JSON.parse(jsonKey);
            // In a real implementation, this would create a JWT auth client
            return {
                type: 'service_account',
                email: keyData.client_email,
                key: keyData.private_key,
                scopes,
                ...options
            };
        } catch (error: any) {
            throw new Error(`Invalid JSON key for connection ${connection}: ${error instanceof Error ? error.message : String(error)}`);
        }
    }

    // Case 3: No authentication method found
    throw new ConnectionInitError(connection);
}

/**
 * Initialize user credentials for Google APIs using OAuth 2.0.
 *
 * @param integration - AutoKitteh integration name.
 * @param connection - AutoKitteh connection name.
 * @param scopes - List of OAuth permission scopes.
 * @returns Google API OAuth2 client.
 * @throws ConnectionInitError if the connection was not initialized yet.
 * @throws OAuthRefreshError if OAuth token refresh failed.
 */
function googleAuthOAuth2(integration: string, connection: string, scopes: string[]): OAuth2Client {
    // Get expiry time from environment
    const expiry = process.env[`${connection}__oauth_Expiry`];
    if (!expiry) {
        throw new ConnectionInitError(connection);
    }

    // Convert Go's time string (e.g. "2024-06-20 19:18:17 -0700 PDT") to
    // an ISO-8601 string that JavaScript can parse with timezone awareness
    const expiryIso = expiry
        .replace(/ [A-Z]+.*$/, '')
        .replace(/\.\d+/, '');

    // Get tokens from environment
    const token = process.env[`${connection}__oauth_AccessToken`];
    const refreshToken = process.env[`${connection}__oauth_RefreshToken`];
    const clientId = process.env.GOOGLE_CLIENT_ID;
    const clientSecret = process.env.GOOGLE_CLIENT_SECRET || 'NOT AVAILABLE';

    // Create OAuth2 client simulation
    const oauth2Client = {
        credentials: {
            access_token: token,
            refresh_token: refreshToken,
            expiry_date: new Date(expiryIso).getTime(),
            scope: scopes.join(' ')
        },

        // Refresh token implementation
        refreshAccessToken: async () => {
            try {
                const [newToken, expiryDate] = await connectionUtils.refreshOAuth(integration, connection);
                oauth2Client.credentials.access_token = newToken;
                oauth2Client.credentials.expiry_date = expiryDate.getTime();
                return {credentials: oauth2Client.credentials};
            } catch (error: any) {
                throw new OAuthRefreshError(connection, error instanceof Error ? error : new Error(String(error)));
            }
        }
    };

    // Check if token is expired and refresh if necessary
    const now = Date.now();
    if (oauth2Client.credentials.expiry_date && oauth2Client.credentials.expiry_date <= now) {
        try {
            // Fire and forget since we don't need to wait for the refresh to complete
            oauth2Client.refreshAccessToken().catch((error: any) => {
                console.error(`Failed to refresh token for ${connection}:`, error);
            });
        } catch (error: any) {
            console.error(`Error initializing token refresh for ${connection}:`, error);
            // We don't throw here to allow initialization even if token refresh fails initially
        }
    }

    return oauth2Client;
}

/**
 * Extract the Google Doc/Form/Sheet ID from a URL. This function is idempotent.
 *
 * Example: 'https://docs.google.com/.../d/1a2b3c4d5e6f/edit' --> '1a2b3c4d5e6f'
 *
 * @param url - The Google URL or ID
 * @returns The extracted Google ID
 * @throws Error if the URL does not contain a valid Google ID
 */
export function googleId(url: string): string {
    const match = url.match(/(.+\/d\/(e\/)?)?([a-zA-Z0-9_-]{20,})/);
    if (match) {
        return match[3];
    } else {
        throw new Error(`Invalid Google ID in "${url}"`);
    }
}
