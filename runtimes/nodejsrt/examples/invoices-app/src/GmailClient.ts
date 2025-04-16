import fs from 'fs';
import path from 'path';
import * as os from "node:os";
import {gmail_v1} from 'googleapis';
import {gmailClient} from 'autokitteh/google';

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
 * GmailClient is a client for interacting with the Gmail API, specifically tailored
 * to fetch emails with PDF attachments.
 */
class GmailClient {
    private connectionName: string;
    private subjectRegex?: RegExp;
    private client: any;

    constructor(
        connectionName?: string,
        subjectRegex?: string
    ) {
        this.connectionName = connectionName || 'gmail';
        this.subjectRegex = subjectRegex ? new RegExp(subjectRegex, "i") : undefined;
    }

    async init(): Promise<GmailClient> {

        // Initialize the Gmail client using autokitteh's gmailClient
        this.client = gmailClient(this.connectionName);
        console.log("Gmail client initialized successfully");
        return this;
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
        if (this.subjectRegex)
            return this.subjectRegex.test(subject);
        else
            return true;
    }

    getMessageDate(message: gmail_v1.Schema$Message): number {
        return parseInt(message.internalDate || "0", 10);

    }

    /**
     * Fetches new emails with attachments that have been received since the last processed time.
     *
     * @return A promise that resolves to an array of email messages with their attachments.
     */
    async fetchNewEmails(startTimestamp: number): Promise<EmailMessage[]> {

        // Create query to find emails with attachments after the startTimestamp
        let query = ["in:inbox", "has:attachment"]
        if (startTimestamp){
            query.push(`after:${Math.floor(Date.now() / 1000)}`)
        }

        console.log(`Fetching emails with query: ${query.join(" ")}`);

        // List messages matching the query
        const response = await this.client.users.messages.list({
            userId: 'me',
            q: query
        });

        console.log('emails fetched:', response.data.messages?.length || '')

        return response.data.messages || [];
    }

    async fetchMessage(msgId: string): Promise<EmailMessage | undefined> {
        const {data: message} = await this.client.users.messages.get({
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
                const {data: attachmentData} = await this.client.users.messages.attachments.get({
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
        const date = this.getMessageDate(message);
        return {id: msgId, date, attachments};
    }
}

export default GmailClient;
