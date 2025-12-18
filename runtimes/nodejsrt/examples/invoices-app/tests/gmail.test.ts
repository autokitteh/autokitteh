import * as dotenv from 'dotenv';

dotenv.config();
import EmailClient from "../src/EmailClient";

describe('GmailClient Tests', () => {
    let client: EmailClient;

    beforeAll(async () => {
        // Initialize the GmailClient before running tests
        client = await new EmailClient().init();
    });

    it('should fetch new emails', async () => {
        const emails = await client.fetchNewEmails(0);
        // Make assertions to verify the emails array
        expect(emails).toBeInstanceOf(Array);
        if (emails.length > 0) {
            expect(emails[0]).toHaveProperty('id');
        }
    });

    it('should fetch details of a specific email', async () => {
        const emails = await client.fetchNewEmails(0);
        if (emails.length > 0) {
            const message = await client.fetchMessage(emails[0].id);
            // Make assertions to check the fetched email details
            expect(message).toHaveProperty('id');
            expect(message).toHaveProperty('snippet');
        } else {
            console.warn('No emails to fetch messages from.');
        }
    });
});
dotenv.config();
import EmailClient from "../src/EmailClient";

test ('gmails',async () => {

    const client = await new EmailClient().init();
    const emails = await client.fetchNewEmails(0)
    console.log(emails)
    const message = await client.fetchMessage(emails[0].id)
    console.log(message)
})
