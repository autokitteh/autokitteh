import config from './config.js';
import GmailEmailFetcher from './GmailEmailFetcher.js';
import ChatGPTClient from './ChatGPTClient.js';
import InvoiceStorage from './InvoiceStorage.js';
import InvoiceProcessor from './InvoiceProcessor.js';
import startServer from './server.js';

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
    const emailFetcher = new GmailEmailFetcher();
    const chatGPTClient = new ChatGPTClient(config.chatGPT.promptTemplate);
    const processor = new InvoiceProcessor(emailFetcher, chatGPTClient, storage);

    startServer(storage, 3000);

    while (true) {
        console.log("Checking for new emails...");
        await processor.processNewEmails();
        console.log(`Waiting ${config.sleepIntervalMs / 1000} seconds...`);
        await new Promise(resolve => setTimeout(resolve, config.sleepIntervalMs));
    }
}

main().catch((error: unknown) => {
    console.error("Fatal error:", error);
});