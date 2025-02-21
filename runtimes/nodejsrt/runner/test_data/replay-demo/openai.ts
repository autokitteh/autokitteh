import OpenAI from 'openai';

const client = new OpenAI({
    apiKey: process.env['OPENAI_API_KEY'], // This is the default and can be omitted
});

async function main() {
    const chatCompletion = await client.chat.completions.create({
        messages: [{ role: 'user', content: 'Say this is a test' }],
        model: 'gpt-4o',
    });

    console.log(chatCompletion);
}

main();


//  https://platform.openai.com/docs/api-reference/files/create?lang=node.js
