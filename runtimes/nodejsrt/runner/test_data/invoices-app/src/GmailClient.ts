import fs from "fs";
import os from "os";
import path from "path";
import {google, gmail_v1} from "googleapis";
import {authenticate} from "@google-cloud/local-auth";
import config from "./config";

const SCOPES = ['https://www.googleapis.com/auth/gmail.readonly'];
const CREDENTIALS_PATH = config.gmail.credentialsPath;
const TOKEN_PATH = config.gmail.tokenPath;

export interface EmailAttachment {
    id: string;
    filePath?: string;
    fileName: string;
    fileBuffer?: Buffer;
}

export interface EmailMessage {
    id: string;
    date: number;
    attachments: EmailAttachment[];
}

/**
 * Class to fetch new emails with attachments from a Gmail account using the Gmail API.
 * Handles email retrieval based on a start timestamp and processes attachments found in the emails.
 */
class GmailClient {
    private lastProcessedTime: number;
    private subjectRegex: RegExp;
    private gmail: gmail_v1.Gmail | null = null;

    constructor(
        startTimestamp: number = 0,
        subjectRegex: string = config.gmail.subjectFilter
    ) {
        this.lastProcessedTime = startTimestamp;
        this.subjectRegex = new RegExp(subjectRegex, "i");
    }

    async initialize(): Promise<void> {
        const authClient = await this.authenticate();
        this.gmail = google.gmail({version: 'v1', auth: authClient});
    }

    private async authenticate() {
        if (!fs.existsSync(CREDENTIALS_PATH)) {
            throw new Error('credentials.json file not found. Please provide it in the application directory.');
        }
        let authClient;
        if (fs.existsSync(TOKEN_PATH)) {
            const token = JSON.parse(fs.readFileSync(TOKEN_PATH, 'utf-8'));
            authClient = new google.auth.OAuth2();
            authClient.setCredentials(token);
            console.log('Using saved token for authentication.');
        } else {
            authClient = await authenticate({keyfilePath: CREDENTIALS_PATH, scopes: SCOPES});
            // Save token for future use
            const credentials = authClient.credentials;
            if (credentials) {
                fs.writeFileSync(TOKEN_PATH, JSON.stringify(credentials));
                console.log('Token stored to', TOKEN_PATH);
            }
        }
        return authClient;
    }


    /**
     * Checks if the provided message's subject matches the defined regular expression.
     *
     * @param message - The message object containing payload and headers.
     * @returns Returns true if the subject matches the regular expression; otherwise, false.
     */
    async isRelevant(message: gmail_v1.Schema$Message): Promise<boolean> {
        const subjectHeader = message.payload?.headers?.find(
            (header) => header.name === "Subject"
        );
        const subject = subjectHeader?.value || "";

        // Apply regex filtering on the subject if set
        return this.subjectRegex.test(subject);
    }

    getMessageDate(message: gmail_v1.Schema$Message): number {
        let date = parseInt(message.internalDate || "0", 10);
        return date;
        // const dateHeader = message.payload?.headers?.find(
        //     (header) => header.name === "Date"
        // );
        // return new Date(dateHeader?.value || "").getTime();
    }
    updateLastProcessedTime(message: gmail_v1.Schema$Message) {
        const date = this.getMessageDate(message);
        console.log('Message date:', date);
        this.lastProcessedTime = Math.max(this.lastProcessedTime, date);
        console.log(`Last processed time: ${this.lastProcessedTime}`);
    }

    /**
     * Fetches new emails from the Gmail inbox, optionally filtering for attachments and messages
     * received after the last processed time.
     *
     * @return A promise that resolves to an array of message objects representing the new emails.
     */
    async fetchNewEmails(): Promise<gmail_v1.Schema$Message[] | undefined> {
        const query = [
            "in:inbox",
            "has:attachment",
            ...(this.lastProcessedTime ? [`after:${5+ Math.floor(this.lastProcessedTime / 1000)}`] : []),
        ].join(" ");

        console.log(`Fetching emails with query: ${query}`);
        const listRes = await this.gmail.users.messages.list({
            userId: "me",
            q: query,
        });

        console.log('emails fetched:', listRes.data.messages?.length || '')

        return listRes.data.messages || [];
    }

    /**
     * Fetches an email message by its ID, retrieves relevant attachments,
     * and returns the message details and a list of extracted attachments.
     *
     * @param msgId - The ID of the email message to be retrieved.
     * @return Resolves with an object containing:
     *     - id: The ID of the email message.
     *     - date: The date of the email message as a Unix timestamp.
     *     - attachments: A list of objects representing attachments. Each attachment object contains:
     *         - id: The ID of the attachment.
     *         - filePath: The temporary file path of the downloaded attachment.
     *         - fileName: The filename of the attachment.
     */
    async fetchMessage(msgId: string): Promise<EmailMessage | undefined> {
        const {data: message} = await this.gmail.users.messages.get({
            userId: "me",
            id: msgId,
            format: "full",
        });


        if (!message || !(await this.isRelevant(message))) {
            this.updateLastProcessedTime(message);
            return null;
        }

        // Process attachments and metadata
        const attachments: EmailAttachment[] = [];
        const parts = message.payload?.parts || [];

        for (const part of parts) {
            if (part.filename && part.filename.endsWith('.pdf') && part.body?.attachmentId) {
                const attachId = part.body.attachmentId;
                const {data: attachmentData} = await this.gmail.users.messages.attachments.get({
                    userId: "me",
                    messageId: msgId,
                    id: attachId,
                });

                // write to temp file TODO in-memory
                const filePath = path.join(os.tmpdir(), part.filename);
                fs.writeFileSync(filePath, attachmentData.data, "base64");

                attachments.push({
                    id: attachId,
                    filePath,
                    fileName: part.filename,
                });
            }
        }

        this.updateLastProcessedTime(message);
        const date = this.getMessageDate(message);

        return {id: msgId,  date, attachments};
    }
}

export default GmailClient;
