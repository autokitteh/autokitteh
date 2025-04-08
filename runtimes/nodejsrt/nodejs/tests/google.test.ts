import {describe, test, expect, beforeEach} from '@jest/globals';
import {
    gmailClient,
} from '../nodejs-sdk/autokitteh/google';
import {ConnectionInitError, OAuthRefreshError} from '../nodejs-sdk/autokitteh/errors';

// Mock environment setup helper
const mockEnv = (connection: string, envVars: Record<string, string>) => {
    for (const [key, value] of Object.entries(envVars)) {
        process.env[`${connection}__${key}`] = value;
    }
};

// Mock environment cleanup helper
const cleanEnv = (connection: string, keys: string[]) => {
    for (const key of keys) {
        delete process.env[`${connection}__${key}`];
    }
};

describe('Google Integration Tests', () => {
    const testConnection = 'test_google_conn';

    beforeEach(() => {
        // Clean up any environment variables before each test
        cleanEnv(testConnection, [
            'authType',
            'oauth_AccessToken',
            'oauth_RefreshToken',
            'oauth_Expiry',
            'JSON',
            'api_key'
        ]);
    });

    describe('OAuth Client Initialization', () => {
        test('should initialize Gmail client with OAuth credentials', () => {
            mockEnv(testConnection, {
                'authType': 'oauth',
                'oauth_AccessToken': 'test_access_token',
                'oauth_RefreshToken': 'test_refresh_token',
                'oauth_Expiry': '2024-12-31 23:59:59 +0000 UTC'
            });

            const client = gmailClient(testConnection);
            expect(client).toBeDefined();
            expect(client.auth).toBeDefined();
            expect(client.auth.credentials.access_token).toBe('test_access_token');
        });

    });

    describe('Error Handling', () => {
        test('should throw ConnectionInitError when no auth method is available', () => {
            expect(() => gmailClient(testConnection)).toThrow(ConnectionInitError);
        });

        test('should throw Error for invalid connection names', () => {
            expect(() => gmailClient('')).toThrow('Invalid AutoKitteh connection name');
            expect(() => gmailClient('invalid name')).toThrow('Invalid AutoKitteh connection name');
        });

    });

    describe('API Method Stubs', () => {
        test('Gmail API methods should throw not implemented error', async () => {
            mockEnv(testConnection, {
                'authType': 'oauth',
                'oauth_AccessToken': 'test_access_token',
                'oauth_RefreshToken': 'test_refresh_token',
                'oauth_Expiry': '2024-12-31 23:59:59 +0000 UTC'
            });

            const client = gmailClient(testConnection);
            await expect(client.users.messages.list({})).rejects.toThrow('Gmail API not yet implemented');
            await expect(client.users.messages.get({})).rejects.toThrow('Gmail API not yet implemented');
            await expect(client.users.messages.send({})).rejects.toThrow('Gmail API not yet implemented');
        });

    });
});
