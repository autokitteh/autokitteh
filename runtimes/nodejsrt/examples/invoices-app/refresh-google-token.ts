/**
 * refresh google auth token
 *
 * Usage: `npx ts-node --transpile-only refresh-google-token.ts`
 */

import fs from "fs";
import {authenticate} from "@google-cloud/local-auth";
import config from "./src/config";

const SCOPES = ['https://www.googleapis.com/auth/gmail.readonly'];
const CREDENTIALS_PATH = config.gmail.credentialsPath;
const TOKEN_PATH = config.gmail.tokenPath;

(async () => {
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
})();
