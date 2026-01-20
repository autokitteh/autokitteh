import * as fs from "fs";
import * as path from "path";
import * as dotenv from 'dotenv';
dotenv.config({ path: path.resolve(__dirname, '.env') });
import * as process from "node:process";
import * as autokitteh from 'autokitteh';
import EmailClient from './EmailClient';
import ChatGPTClient from './ChatGPTClient';
import InvoiceStorage from './InvoiceStorage';
import InvoiceProcessor from './InvoiceProcessor';


// Execution parameters
const sleepIntervalSec = Number(process.env.SLEEP_INTERVAL_SEC) || 60;

// Gmail parameters
const subjectFilter = process.env.SUBJECT_FILTER || '.*invoice.*';
const gmailConnectionName = process.env.GMAIL_CONNECTION_NAME || 'gmail';

// chatGPT parameters
const chatConnectionName = process.env.OPENAI_CONNECTION_NAME || 'openai';
const promptTemplate = fs.readFileSync(path.join(__dirname, 'chatgpt_prompt.txt'), 'utf8');
const gptModel = process.env.GPTMODEL || 'gpt-4o';

// Global storage
const storage = new InvoiceStorage();

/**
 * Main entry point - processes emails for a specific month and waits for new ones
 * @param event Event with yearMonth parameter (YYYY-MM format)
 * @returns Promise that resolves when the month ends
 */
export async function on_event(event: any = {}): Promise<any> {
    const year = parseInt(event.year, 10) || new Date().getFullYear();
    const month = parseInt(event.month) || new Date().getMonth() + 1;
    const startTime = new Date(year, month - 1, 1, 0, 0, 0, 0).getTime();
    const endTime = new Date(year, month, 1, 0, 0, 0, 0).getTime();

    console.log(`Processing invoices for ${year}-${month}`);

    // Subscribe to Gmail events
    const filter = `event_type == 'mailbox_change' && data.subject.matches('${subjectFilter}')`;
    const subscriptionId = await autokitteh.subscribe(gmailConnectionName, filter);
    console.log(`Subscribed to Gmail events with ID: ${subscriptionId}`);

    await processAndPrintTotal(startTime);
    // Wait and process new emails until the end of month
    while (Date.now() < endTime) {
        const event = await autokitteh.nextEvent(subscriptionId, {timeout: {seconds: sleepIntervalSec}});
        if (event) {
            console.log("New email received:", event.data?.subject);
            await processAndPrintTotal(startTime);
        }
    }

    // Cleanup
    console.log(`Unsubscribing from Gmail events with ID: ${subscriptionId}...`);
    await autokitteh.unsubscribe(subscriptionId);

    // Execution summary
    console.log(`Total invoices: ${storage.getInvoices().length}, Total amount: ${storage.getTotalAmount()}`);

    // Return final state
    return {
        status: "completed",
        total: storage.getTotalAmount(),
        invoiceCount: storage.getInvoices().length
    };
}

/**
 * Process new emails and print the total
 */
async function processAndPrintTotal(startTimestamp: number): Promise<number> {

    // Initialize Gmail client
    const gmailClient = await new EmailClient(gmailConnectionName).init();

    // Initialize OpenAI client
    const chatGPTClient = await new ChatGPTClient(promptTemplate, chatConnectionName, gptModel).init();

    // Initialize InvoiceProcessor
    const processor = new InvoiceProcessor(gmailClient, chatGPTClient, storage);

    // Process new emails and print total amount
    await processor.processNewEmails(startTimestamp);
    const total = storage.getTotalAmount();
    console.log(`Total invoice amount: ${total}`);

    return total;
}

