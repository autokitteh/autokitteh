import GmailClient from "../src/GmailClient";


test ('gmails',async () => {

    const client = new GmailClient();
    await client.initialize();
    const emails = await client.fetchNewEmails()
    console.log(emails)

})
