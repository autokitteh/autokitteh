import {promises as fs} from 'fs';
import EmailClient from "./EmailClient";
import {EmailMessage} from "./EmailClient";
import ChatGPTClient from "./ChatGPTClient";
import InvoiceStorage from "./InvoiceStorage";
import {InvoiceData} from "./InvoiceStorage";

export class InvoiceProcessor {
    private gmailClient: EmailClient;
    private chatGPTClient: ChatGPTClient;
    private storage: InvoiceStorage;

    constructor(gmailClient: EmailClient, chatGPTClient: ChatGPTClient, storage: InvoiceStorage) {
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
            if (!attachment.filePath) {
                console.warn(`Attachment ${attachment.fileName} has no file path and will be skipped.`);
                continue;
            }
            const invoiceData = await this.chatGPTClient.analyzeAttachment(attachment.filePath);
            if (invoiceData && this.validateInvoice(invoiceData)) {
                this.storage.addOrUpdateInvoice({...invoiceData, date: email.date});
            } else {
                console.log(`Invalid invoice detected in attachment ${attachment.fileName}. Ignored.`);
            }
            await fs.rm(attachment.filePath, {force: true});
        }
    }

    /**
     * Processes new emails by fetching them from the Gmail client, retrieving their full message data,
     * and passing the data to a processing function.
     */
    async processNewEmails(startTimestamp:number): Promise<void> {
        let emailIds = (await this.gmailClient.fetchNewEmails(startTimestamp))?.map(email => email.id);
        emailIds = this.storage.filterProcessed(emailIds)
        await Promise.all(
            (emailIds || []).map(async (emailId) => {
                if (emailId) {
                    const emailData = await this.gmailClient.fetchMessage(emailId);
                    if (emailData) {
                        await this.processEmail(emailData);
                    }
                    this.storage.markProcessed(emailId);
                    console.log(`Processed email: ${emailId}`);
                }
            })
        );
    }
}

export default InvoiceProcessor;
