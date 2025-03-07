import * as dotenv from 'dotenv';
dotenv.config();
import * as fs from 'fs';
import * as path from 'path';


interface GmailConfig {
    credentialsPath: string;
    tokenPath: string;
    subjectFilter: string;
}

interface ChatGPTConfig {
    apiKey: string;
    promptTemplate: string;
    model: string;
}

interface ServerConfig {
    port: number;
}

interface Config {
    sleepIntervalSec: number;
    gmail: GmailConfig;
    chatGPT: ChatGPTConfig;
    server: ServerConfig;
}

const config: Config = {
    sleepIntervalSec: Number(process.env.SLEEP_INTERVAL_SEC) || 60,
    gmail: {
        credentialsPath: process.env.GMAIL_CREDENTIALS_PATH || '',
        tokenPath: process.env.GMAIL_TOKEN_PATH || '',
        subjectFilter: process.env.SUBJECT_FILTER || '.*invoice.*',
    },
    chatGPT: {
        apiKey: process.env.OPENAI_API_KEY || '',
        promptTemplate: fs.readFileSync(path.join(__dirname,'../src/chatgpt_prompt.txt'), 'utf8'),
        model: 'gpt-4',
    },
    server: {
        port: Number(process.env.SERVER_PORT) || 3000,
    },
};

export default config;
