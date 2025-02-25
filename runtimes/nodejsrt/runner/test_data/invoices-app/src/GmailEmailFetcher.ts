import config from "./config.js";
import path from "path";
import os from "os";
import fs from "fs";

import {google, gmail_v1} from "googleapis";
import {OAuth2Client} from "google-auth-library";

export interface EmailAttachment {
    id: string;
    filePath: string;
    fileName: string;
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
class GmailEmailFetcher {
    private oauth2Client!: OAuth2Client;
    private gmail!: gmail_v1.Gmail;
    private lastProcessedTime: number;
    private subjectRegex: RegExp;

    constructor(
        startTimestamp: number = 0,
        subjectRegex: string = config.gmail.subjectFilter
    ) {
        this.initConnection();
        this.lastProcessedTime = startTimestamp;
        this.subjectRegex = new RegExp(subjectRegex, "i");
    }

    /**
     * Initializes a connection to the Gmail API using OAuth2 credentials.
     * @throws {Error} If required environment variables are missing.
     */
    private initConnection(): void {
        const clientId = process.env.GOOGLE_CLIENT_ID;
        const clientSecret = process.env.GOOGLE_CLIENT_SECRET;
        const refreshToken = process.env.GOOGLE_REFRESH_TOKEN;
        const user = process.env.GMAIL_USER;

        if (!clientId || !clientSecret || !refreshToken || !user) {
            throw new Error(
                "Missing Gmail API credentials. Please set GMAIL_CLIENT_ID, GMAIL_CLIENT_SECRET, GMAIL_REFRESH_TOKEN, and GMAIL_USER in your environment."
            );
        }
        this.oauth2Client = new OAuth2Client(clientId, clientSecret);
        this.oauth2Client.setCredentials({refresh_token: refreshToken});
        this.gmail = google.gmail({version: "v1", auth: this.oauth2Client});
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
            ...(this.lastProcessedTime ? [`after:${this.lastProcessedTime}`] : []),
        ].join(" ");

        console.log(`Fetching emails with query: ${query}`);
        const listRes = await this.gmail.users.messages.list({
            userId: "me",
            q: query,
        });

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
            return undefined;
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

                const filePath = path.join(os.tmpdir(), part.filename);
                fs.writeFileSync(filePath, attachmentData.data, "base64");

                attachments.push({
                    id: attachId,
                    filePath,
                    fileName: part.filename,
                });
            }
        }

        const dateHeader = message.payload?.headers?.find(
            (header) => header.name === "Date"
        );
        // const date = new Date(dateHeader?.value || "").getTime();
        let date = parseInt(message.internalDate || "0", 10);

        // Update last process time
        console.log('internalDate', message.internalDate);
        console.log(`Processing email ${msgId} with date ${date}`);
        this.lastProcessedTime = Math.max(this.lastProcessedTime, date);
        console.log(`Last processed time: ${this.lastProcessedTime}`);

        return {
            id: msgId,
            date,
            attachments,
        };
    }
}

export default GmailEmailFetcher;