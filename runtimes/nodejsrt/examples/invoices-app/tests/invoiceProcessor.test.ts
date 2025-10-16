import InvoiceProcessor from "../src/InvoiceProcessor";
import InvoiceStorage, {InvoiceData} from "../src/InvoiceStorage";
import EmailClient from "../src/EmailClient";
import ChatGPTClient from "../src/ChatGPTClient";


class FakeGmailClient implements Partial<EmailClient> {

    constructor(private readonly msgs: any) {
        this.msgs = msgs;
    }

    async fetchNewEmails(): Promise<any[]> {
        return this.msgs;
    }

    async fetchMessage(msgId: string): Promise<any | undefined> {
        return this.msgs.find((msg: { id: string }) => msg.id === msgId);
    }
}

class FakeChatGPTClient implements Partial<ChatGPTClient>{
    private msgs: any;

    constructor(msgs: any) {
        this.msgs = msgs;
    }

    async analyzeAttachment(filePath: string): Promise<InvoiceData | null> {
        const result = this.msgs.map((msg: { attachments: any; }) => msg.attachments || []).flat().find(
            (attachment: { filePath: string; invoice?: any }) => attachment.filePath === filePath
        );
        return result?.invoice;
    }
}

describe('InvoiceProcessor testing', () => {
    beforeEach(() => {
        // Mock the method to prevent actual implementation from being called
        // jest.spyOn(gmailClient, 'fetchNewEmails').mockResolvedValue([]);
    });

    it('main flow', async () => {
        const msgs: any[] = [
            {
                id: "1",
                date: 1740334461,
                attachments: [
                    {
                        id: "att-1",
                        filePath: "./inv-1.pdf",
                        fileName: "inv-1.pdf",
                        invoice: {
                            isInvoice: true,
                            invoiceId: "inv-1",
                            items: [
                                { description: "Item 1", amount: 20 },
                                { description: "Item 2", amount: 10 }
                            ],
                            vat: 0,
                            total: 30
                        }
                    }
                ]
            },
            {
                id: "2",
                date: 1740334461,
                attachments: [
                    {
                        id: "att-2",
                        filePath: "./inv-2.pdf",
                        fileName: "inv-2.pdf",
                        invoice: {
                            isInvoice: true,
                            invoiceId: "inv-2",
                            items: [
                                { description: "Item 1", amount: 30 },
                                { description: "Item 2", amount: 10 }
                            ],
                            vat: 0,
                            total: 40
                        }
                    }
                ]
            }
        ];

        const storage = new InvoiceStorage();
        const emailFetcher:any = new FakeGmailClient(msgs);
        const chatGPTClient:any = new FakeChatGPTClient(msgs);
        const processor = new InvoiceProcessor(emailFetcher, chatGPTClient, storage);

        await processor.processNewEmails();
        expect(storage.getTotalAmount()).toEqual(70);
    });

    it('check storage', async () => {
        const storage = new InvoiceStorage();
        const inv:any = {
            isInvoice: true,
            invoiceId: "inv-1",
            items: [
                { description: "Item 1", amount: 20 },
                { description: "Item 2", amount: 10 }
            ],
            vat: 0,
            total: 30,
            date: 1740334461,
        };

        // Add the first invoice
        storage.addOrUpdateInvoice({ ...inv });
        expect(storage.getTotalAmount()).toEqual(30);

        // Add the same invoice with an older date
        storage.addOrUpdateInvoice({ ...inv, total: 50, date: 1740334460 });
        expect(storage.getTotalAmount()).toEqual(30);

        // Add the same invoice with a more recent date
        const inv1 = { ...inv, total: 40, date: 1740334462 };
        storage.addOrUpdateInvoice(inv1);
        expect(storage.getTotalAmount()).toEqual(40);

        // Add another invoice
        storage.addOrUpdateInvoice({ ...inv, invoiceId: "inv-2", total: 50.25 });
        expect(storage.getTotalAmount()).toEqual(90.25);

        // Get the latest updated invoice
        expect(storage.getInvoice("inv-1")).toEqual(inv1);
    });
});
