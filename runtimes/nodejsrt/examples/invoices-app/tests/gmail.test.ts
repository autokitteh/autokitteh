import GmailClient from "../src/GmailClient";


test ('gmails',async () => {

    const client = await new GmailClient().init();
    const emails = await client.fetchNewEmails()
    console.log(emails)

})
