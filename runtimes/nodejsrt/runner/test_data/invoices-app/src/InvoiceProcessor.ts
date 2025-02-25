import { promises as fs } from 'fs';
import GmailEmailFetcher from "./GmailEmailFetcher.js";
import {EmailMessage} from "./GmailEmailFetcher.js";
import ChatGPTClient from "./ChatGPTClient.js";
import InvoiceStorage from "./InvoiceStorage.js";
import {InvoiceData} from "./InvoiceStorage.js";

export class InvoiceProcessor {
    private gmailClient: GmailEmailFetcher;
    private chatGPTClient: ChatGPTClient;
    private storage: InvoiceStorage;

    constructor(gmailClient: GmailEmailFetcher, chatGPTClient: ChatGPTClient, storage: InvoiceStorage) {
        this.gmailClient = gmailClient;
        this.chatGPTClient = chatGPTClient;
        this.storage = storage;
    }

    /**
     * Validates the invoice data by checking if the computed total matches the total provided in the data.
     *
     * @param data - The invoice data object.
     * @return Returns `true` if the computed total matches the provided total within a precision of 0.01, otherwise `false`.
     */
    validateInvoice(data: InvoiceData): boolean {
        if (!data.invoiceId) return false;
        const computedTotal = data.items.reduce((sum, item) => sum + item.amount, 0) + data.vat;
        return Math.abs(computedTotal - data.total) < 0.01;
    }

    /**
     * Processes an email to analyze its attachments, validate invoice data, and store valid invoices.
     *
     * @param email - The email object containing details and attachments.
     */
    async processEmail(email: EmailMessage): Promise<void> {
        for (const attachment of email.attachments) {
            const invoiceData = await this.chatGPTClient.analyzeAttachment(attachment.filePath);
            if (invoiceData && this.validateInvoice(invoiceData)) {
                this.storage.addOrUpdateInvoice({ ...invoiceData, date: email.date });
            } else {
                console.log(`Invalid invoice detected in attachment ${attachment.fileName}. Ignored.`);
            }
            await fs.rm(attachment.filePath, { force: true });
        }
    }

    /**
     * Processes new emails by fetching them from the Gmail client, retrieving their full message data,
     * and passing the data to a processing function.
     */
    async processNewEmails(): Promise<void> {
        const emails = await this.gmailClient.fetchNewEmails();
        await Promise.all(
            (emails || []).map(async (email) => {
                console.log(`Processing email: ${email.id}`);
                if (email.id) {
                    const emailData = await this.gmailClient.fetchMessage(email.id);
                    if (emailData) {
                        await this.processEmail(emailData);
                    }
                }
            })
        );
    }
}

export default InvoiceProcessor;