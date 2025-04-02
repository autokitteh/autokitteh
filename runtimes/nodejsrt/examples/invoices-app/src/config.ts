import * as dotenv from 'dotenv';
dotenv.config();
import * as fs from 'fs';
import * as path from 'path';

interface GmailConfig {
    connectionName: string;
    subjectFilter: string;
}

interface ChatGPTConfig {
    connectionName: string;
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
        connectionName: process.env.GMAIL_CONNECTION_NAME || 'gmail',
        subjectFilter: process.env.SUBJECT_FILTER || '.*invoice.*',
    },
    chatGPT: {
        connectionName: process.env.OPENAI_CONNECTION_NAME || 'openai',
        promptTemplate: fs.readFileSync(path.join(__dirname,'../src/chatgpt_prompt.txt'), 'utf8'),
        model: 'gpt-4o',
    },
    server: {
        port: Number(process.env.SERVER_PORT) || 3000,
    },
};

export default config;
