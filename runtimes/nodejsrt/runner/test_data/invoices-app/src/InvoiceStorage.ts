
/**
 * Interface defining the structure of an invoice object.
 */
export interface InvoiceData {
    invoiceId: string;
    items: InvoiceItem[];
    vat: number;
    total: number;
    date?: number;
}

export interface InvoiceItem {
    description: string;
    amount: number;
}

/**
 * A class for storing and managing invoices. The InvoiceStorage class provides
 * functionality for adding or updating invoices, retrieving a specific invoice,
 * and calculating the total amount of all stored invoices.
 */
class InvoiceStorage {
    private invoices: Map<string, InvoiceData>;
    private processedEmails: Set<string> = new Set();

    constructor() {
        this.invoices = new Map<string, InvoiceData>();
    }

    /**
     * Adds a new invoice or updates an existing invoice in the collection.
     * If an invoice with the same ID exists and the provided invoice has a
     * date that is later than the existing invoice's date, it will be updated.
     *
     * @param invoice - The invoice object to be added or updated.
     * @return This method does not return a value.
     */
    addOrUpdateInvoice(invoice: InvoiceData): void {
        const existing = this.invoices.get(invoice.invoiceId);
        if (!existing || invoice.date > existing.date) {
            this.invoices.set(invoice.invoiceId, invoice);
        }
    }

    markProcessed(emailId: string) {
        this.processedEmails.add(emailId);
    }
    isProcessed(emailId: string) {
        return this.processedEmails.has(emailId);
    }

    filterProcessed(emailIds: string[]): string[] {
        return emailIds.filter(emailId => !this.isProcessed(emailId));
    }

    /**
     * Retrieves the invoice object associated with the given invoice ID.
     *
     * @param invoiceId - The unique identifier of the invoice to retrieve.
     * @return The invoice object associated with the specified ID, or undefined if not found.
     */
    getInvoice(invoiceId: string): InvoiceData | undefined {
        return this.invoices.get(invoiceId);
    }

    /**
     * Calculates the total amount from all invoices.
     *
     * @return The sum of the 'total' field from all invoices.
     */
    getTotalAmount(): number {
        return Array.from(this.invoices.values()).reduce(
            (total, inv) => total + inv.total,
            0
        );
    }

    getInvoices(): InvoiceData[] {
        return Array.from(this.invoices.values());
    }
}

export default InvoiceStorage;
