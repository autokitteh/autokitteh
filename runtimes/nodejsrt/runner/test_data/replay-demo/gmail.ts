import path from "path";

import {authenticate} from '@google-cloud/local-auth';
import {google} from "googleapis";



const SCOPES = ['https://www.googleapis.com/auth/gmail.readonly'];
const CREDENTIALS_PATH = path.join(process.cwd(), 'credentials.json');

export async function listSnippets(): Promise<string[]> {
    const auth = await authenticate({
        scopes: SCOPES,
        keyfilePath: CREDENTIALS_PATH,
    })
    const gmail = google.gmail({version: 'v1', auth});
    const res = await gmail.users.messages.list({
        userId: "me",

    });

    let snippets: (string)[] = [];
    let ids: string[] = [];
    if (res.data?.messages !== undefined) {
        let _ids = res.data?.messages?.map(message => message.id)
        if (_ids.every(_id => typeof _id === 'string')) {
            ids = _ids
        }
    }

    for (let id of ids) {
        const msg = await gmail.users.messages.get({id, userId: "me"});
        let snippet = msg.data?.snippet
        if (typeof snippet === "string") {
            snippets.push(snippet);
        }
    }

    return snippets;
}
