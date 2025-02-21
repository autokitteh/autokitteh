import fs from "fs";
import path from "path";
import { authenticate } from "@google-cloud/local-auth";
import { google } from "googleapis";

const SCOPES = ["https://www.googleapis.com/auth/gmail.readonly"];
const TOKEN_PATH = path.join(process.cwd(), "token.json");
const CREDENTIALS_PATH = path.join(process.cwd(), "credentials.json");

async function loadSavedCredentials() {
    try {
        const content = fs.readFileSync(TOKEN_PATH, "utf-8");
        return JSON.parse(content);
    } catch (err) {
        return null;
    }
}

async function saveCredentials(client: any) {
    const payload = JSON.stringify({
        type: "authorized_user",
        client_id: client.credentials.client_id,
        client_secret: client.credentials.client_secret,
        refresh_token: client.credentials.refresh_token,
    });
    fs.writeFileSync(TOKEN_PATH, payload);
}

async function authorize() {
    let client = await loadSavedCredentials();
    if (!client) {
        client = await authenticate({
            scopes: SCOPES,
            keyfilePath: CREDENTIALS_PATH,
        });
        if (client.credentials) {
            await saveCredentials(client);
        }
    }
    return google.auth.fromJSON(client);
}

export default authorize;
