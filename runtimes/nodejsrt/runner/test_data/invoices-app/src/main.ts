import config from './config';
import GmailClient from './GmailClient';
import ChatGPTClient from './ChatGPTClient';
import InvoiceStorage from './InvoiceStorage';
import InvoiceProcessor from './InvoiceProcessor';
import startServer from './server';

/**
 * The main entry point for the application. This function initializes necessary components
 * such as storage, email fetcher, and AI processing client, then starts the server.
 * It continuously checks for new emails, processes them, and waits for a configured interval
 * before repeating the process.
 *
 * @return {Promise<void>} A promise that resolves when the main process completes.
 *                         Typically, this does not resolve as the function contains an infinite loop.
 */
async function main(): Promise<void> {
    const storage = new InvoiceStorage();
    const emailFetcher = new GmailClient();
    await emailFetcher.initialize();
    const chatGPTClient = new ChatGPTClient(config.chatGPT.promptTemplate);
    const processor = new InvoiceProcessor(emailFetcher, chatGPTClient, storage);

    startServer(storage, 3000);

    while (true) {
        console.log("Checking for new emails...");
        try {
            await processor.processNewEmails();
            console.log("Processing complete");
        } catch (error) {
            console.error("Error processing email:", error);
        }
        console.log(`Waiting ${config.sleepIntervalSec} seconds...`);
        await new Promise(resolve => setTimeout(resolve, config.sleepIntervalSec * 1000));
    }
}

main().catch((error: unknown) => {
    console.error("Fatal error:", error);
});
