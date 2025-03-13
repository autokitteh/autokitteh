import config from './config';
import GmailClient from './GmailClient';
import ChatGPTClient from './ChatGPTClient';
import InvoiceStorage from './InvoiceStorage';
import InvoiceProcessor from './InvoiceProcessor';


async function total(event: any): Promise<number> {
    const storage = new InvoiceStorage();
    const gmailClient = await new GmailClient().init();
    const chatGPTClient = await new ChatGPTClient(config.chatGPT.promptTemplate).init();
    const processor = new InvoiceProcessor(gmailClient, chatGPTClient, storage);
    await processor.processNewEmails();
    return storage.getTotalAmount()
}


export default total;
