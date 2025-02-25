import config from "./config.js";
import {OpenAI} from 'openai';
import pdfParse from 'pdf-parse';
import {InvoiceData} from "./InvoiceStorage.js";

/**
 * ChatGPTClient is a client for interacting with OpenAI's GPT models, specifically tailored
 * to analyze and extract information from files such as PDFs using a provided prompt template.
 */

class ChatGPTClient {
    private readonly promptTemplate: string;
    private openai: OpenAI;

    constructor(promptTemplate: string) {
        this.promptTemplate = promptTemplate;
        const apiKey = config.chatGPT.apiKey;
        if (!apiKey) {
            throw new Error("Missing OPENAI_API_KEY. Please set it in your environment.");
        }
        this.openai = new OpenAI({apiKey: apiKey});
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
        try {
            // @ts-ignore
            const data = await pdfParse(filePath);
            const prompt: string = `${this.promptTemplate}\n\nAttachment content:\n${data.text}`;
            const response = await this.openai.chat.completions.create({
                model: config.chatGPT.model,
                messages: [{role: "user", content: prompt}],
            });

            return JSON.parse(response.choices[0].message.content);
        } catch (e) {
            console.error(e);
            return null;
        }
    }
}

export default ChatGPTClient;