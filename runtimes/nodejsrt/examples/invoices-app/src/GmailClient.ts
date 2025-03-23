import { gmailClient } from './autokitteh/google';
import config from './config';
import fs from 'fs';
import path from 'path';
import * as os from "node:os";
import { gmail_v1 } from 'googleapis';

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
    private lastProcessedTime: number;
    private connectionName: string;
    private subjectRegex: RegExp;
    private client: any;

    constructor(
        connectionName?: string,
        startTimestamp: number = 0,
        subjectRegex: string = config.gmail.subjectFilter
    ) {
        this.connectionName = connectionName || 'gmail';
        this.lastProcessedTime = startTimestamp;
        this.subjectRegex = new RegExp(subjectRegex, "i");
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
     * Fetches new emails with attachments that have been received since the last processed time.
     *
     * @return A promise that resolves to an array of email messages with their attachments.
     */
    async fetchNewEmails(): Promise<EmailMessage[]> {

        // Create query to find emails with attachments after the lastProcessedTime
        const query = [
            "in:inbox",
            "has:attachment",
            ...(this.lastProcessedTime ? [`after:${5+ Math.floor(this.lastProcessedTime / 1000)}`] : []),
        ].join(" ");

        console.log(`Fetching emails with query: ${query}`);

        // List messages matching the query
        const response = await this.client.users.messages.list({
            userId: 'me',
            q: query
        });

        console.log('emails fetched:', response.data.messages?.length || '')

        return response.data.messages || [];
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
        const {data: message} = await this.client.users.messages.get({
            userId: "me",
            id: msgId,
            format: "full",
        });


        if (!message || !(await this.isRelevant(message))) {
            this.updateLastProcessedTime(message);
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

        this.updateLastProcessedTime(message);
        const date = this.getMessageDate(message);

        return {id: msgId,  date, attachments};
    }
}

export default GmailClient;
