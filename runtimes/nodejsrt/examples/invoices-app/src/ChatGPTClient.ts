import config from "./config";
import { openaiClient } from "./autokitteh/openai";
import {InvoiceData} from "./InvoiceStorage";
import fs from 'fs';

/**
 * ChatGPTClient is a client for interacting with OpenAI's GPT models, specifically tailored
 * to analyze and extract information from PDF files using a provided prompt template.
 * The class uses openai assistant with file search tool to analyze PDF files.
 *  https://platform.openai.com/docs/assistants/tools/file-search
 */

class ChatGPTClient {
    private readonly promptTemplate: string;
    private readonly connectionName: string;
    private openai: any;
    private assistant: any;

    constructor(promptTemplate: string, connectionName?: string) {
        this.promptTemplate = promptTemplate;
        this.connectionName = connectionName || 'openai';
    }

    async init(): Promise<ChatGPTClient> {
        console.log('Initializing OpenAI client');
        // Initialize the OpenAI client using autokitteh's openaiClient
        this.openai = openaiClient(this.connectionName);
        this.assistant = await this.createAssistant();
        console.log('OpenAI client initialized successfully');
        return this;
    }

    /**
     * Analyzes the content of a given file attachment, processes its content using OpenAI's chat completions API,
     * and returns the parsed result.
     *
     * @param filePath - The file path of the attachment to be analyzed. Expected to be a PDF file.
     * @return A promise that resolves to the parsed JSON response from the API.
     * Returns null if an error occurs during processing.
     */
    async analyzeAttachment(filePath: string): Promise<InvoiceData | null> {

        const prompt: string = this.promptTemplate;
        const fileId = await this.uploadFile(filePath);
        const threadId = await this.createThread(fileId.id, prompt);
        const messages = await this.runAssistant(threadId, this.assistant.id);
        return this.processResponse(messages);
    }

    extractJSON(responseText: string) {
        const match = responseText.match(/```json([\s\S]*?)```/); // Extracts JSON from a code block
        return match ? match[1].trim() : responseText.trim(); // If no code block, return full response
    }

    processResponse(messages: any) {
        const responseText = this.extractJSON(messages.data[0].content[0].text.value);
        console.log('Response: ', responseText);
        return JSON.parse(responseText);
    }


    async uploadFile(filePath: string): Promise<any> {
        const response = await this.openai.files.create({
            file: fs.createReadStream(filePath),
            purpose: "assistants",
        });
        console.log("File uploaded successfully:", response.id);
        return response;

    }

    async createAssistant(): Promise<any> {
        const assistant = await this.openai.beta.assistants.create({
            name: "Invoice Analyst Assistant",
            instructions: "You are an expert financial analyst. Use you knowledge base to answer questions about audited financial statements.",
            model: "gpt-4o",
            tools: [{type: "file_search"}],
        });
        console.log("Assistant created:", assistant.id);
        return assistant;
    }

    async createThread(fileId: any, prompt: string): Promise<string> {

        const thread = await this.openai.beta.threads.create();
        console.log("Thread created:", thread.id);
        const message = await this.openai.beta.threads.messages.create(thread.id, {
            role: "user",
            content: prompt,
            attachments: [{file_id: fileId, tools: [{type: "file_search"}]}]
        });

        console.log("Message sent:", message.id);
        return thread.id;
    }

    async runAssistant(threadId: string, assistantId: string): Promise<any> {

        const run = await this.openai.beta.threads.runs.createAndPoll(threadId, {
            assistant_id: assistantId,
        });

        const messages = await this.openai.beta.threads.messages.list(threadId, {
            run_id: run.id,
        });

        console.log("Retrieved messages for the thread:", threadId);
        return messages;
    }
}

export default ChatGPTClient;
