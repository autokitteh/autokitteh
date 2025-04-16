/**
 * Initialize Google API clients, based on AutoKitteh connections.
 */
import {checkConnectionName} from './connections';
import {connectionUtils} from './connections';
import {ConnectionInitError, OAuthRefreshError} from './errors';
import {google} from 'googleapis';
import {OAuth2Client} from 'google-auth-library';
import {JWT} from 'google-auth-library';
import {RefreshAccessTokenResponse} from "google-auth-library/build/src/auth/oauth2client";

// Common API client options
interface GoogleClientOptions {
    [key: string]: any;
}
/**
 * Base structure for creating Google API clients
 */
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

    const auth = googleAuth('gmail', connection, defaultScopes, options);
    const client = google.gmail({version: 'v1', auth: auth});
    return client;

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

    // Validate connection pattern
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
            return new google.auth.JWT({
                email: keyData.client_email,
                key: keyData.private_key,
                scopes: scopes,
                ...options
            });
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
    // Get tokens from environment
    const token = process.env[`${connection}__oauth_AccessToken`];
    const refreshToken = process.env[`${connection}__oauth_RefreshToken`];
    const clientId = process.env.GOOGLE_CLIENT_ID;
    const clientSecret = process.env.GOOGLE_CLIENT_SECRET || 'NOT AVAILABLE';
    const expiry = Number(process.env[`${connection}__oauth_Expiry`]);
    if (!expiry) {
        throw new ConnectionInitError(connection);
    }

    // Create OAuth2 client
    const oauth2Client = new google.auth.OAuth2(clientId, clientSecret);
    oauth2Client.setCredentials({
        access_token: token,
        refresh_token: refreshToken,
        expiry_date: expiry,
        scope: scopes.join(' ')
    });

    // Override the refresh token function to use AutoKitteh's refresh mechanism
    oauth2Client.refreshAccessToken = async () : Promise<RefreshAccessTokenResponse>=> {
        try {
            const [newToken, expiryDate] = await connectionUtils.refreshOAuth(integration, connection);
            const credentials = {
                access_token: newToken,
                refresh_token: refreshToken,
                expiry_date: expiryDate.getTime(),
                scope: scopes.join(' ')
            };
            oauth2Client.setCredentials(credentials);
            return {credentials, res: null};
        } catch (error: any) {
            throw new OAuthRefreshError(connection, error instanceof Error ? error : new Error(String(error)));
        }
    };

    // Check if token is expired and refresh if necessary
    if (expiry <= Date.now()) {
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
