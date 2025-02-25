import * as dotenv from 'dotenv';
import * as fs from 'fs';
import * as path from 'path';
dotenv.config();

interface GmailConfig {
    credentials: string;
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
    sleepIntervalMs: number;
    gmail: GmailConfig;
    chatGPT: ChatGPTConfig;
    server: ServerConfig;
}

const config: Config = {
    sleepIntervalMs: Number(process.env.SLEEP_INTERVAL_MS) || 60000,
    gmail: {
        credentials: process.env.GMAIL_CREDENTIALS || '',
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