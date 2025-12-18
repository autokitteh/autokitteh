/**
 * refresh google auth token
 *
 * Usage: `npx ts-node --transpile-only refresh-google-token.ts`
 */

import * as fs from "fs";
import {authenticate} from "@google-cloud/local-auth";

const SCOPES = ['https://www.googleapis.com/auth/gmail.readonly'];
const CREDENTIALS_PATH = process.env.GMAIL_CREDENTIALS_PATH || '';
const TOKEN_PATH = process.env.GMAIL_TOKEN_PATH || '';

export async function refreshToken() {
    try {
        const authClient = await authenticate({keyfilePath: CREDENTIALS_PATH, scopes: SCOPES});
        const credentials = authClient.credentials;
        if (credentials) {
            fs.writeFileSync(TOKEN_PATH, JSON.stringify(credentials));
            console.log('Token stored to', TOKEN_PATH);
        }
    } catch (error) {
        console.error('Error during authentication:', error);
    }
}

if (require.main === module) {
    refreshToken();
}
