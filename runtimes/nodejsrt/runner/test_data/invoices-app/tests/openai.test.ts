import ChatGPTClient from "../src/ChatGPTClient";
import config from "../src/config";
import path from "path";

test("ChatGPTClient - analyzeAttachment", async () => {
    const client = new ChatGPTClient(config.chatGPT.promptTemplate, config.chatGPT.apiKey);
    const filePath = path.join(__dirname, "./invoices/inv-1.pdf");
    const response = await client.analyzeAttachment(filePath);
    expect(response).toEqual({
        "isInvoice": true,
        "invoiceId": "inv-1",
        "items": [
            {
                "description": "Item 1",
                "amount": 20
            },
            {
                "description": "Item 2",
                "amount": 10
            }
        ],
        "vat": 0,
        "total": 30
    })
}, 20000)
