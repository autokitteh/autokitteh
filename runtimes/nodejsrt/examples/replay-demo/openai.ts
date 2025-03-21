import OpenAI from 'openai';

const client = new OpenAI({
    apiKey: process.env['OPENAI_API_KEY'], // This is the default and can be omitted
});

export async function isInvoice(subject: string) {
    const chatCompletion = await client.chat.completions.create({
        messages: [{ role: 'user', content: `i wanna know if email with the following subject might contain an invoice. answer just in yes or no. subject: ${subject}` }],
        model: 'gpt-4o',
    });

    console.log(chatCompletion.choices[0].message.content);
}


//  https://platform.openai.com/docs/api-reference/files/create?lang=node.js
