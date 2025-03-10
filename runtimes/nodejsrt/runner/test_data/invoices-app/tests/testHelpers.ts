import fs from 'fs';
import path from 'path';
import PDFDocument from 'pdfkit';

/**
 * Generates a dummy PDF invoice and stores it to a temporary file.
 * @param invoiceNumber - The invoice number (default: '12345').
 * @param invoiceDate - The invoice date (default: '2023-10-15').
 * @returns A promise that resolves to the path of the generated PDF file.
 */
async function generateDummyPDFInvoice(
    invoiceNumber: string = '12345',
    invoiceDate: string = '2023-10-15',
): Promise<string> {
    return new Promise((resolve, reject) => {
        const doc = new PDFDocument();
        const filePath = path.join(__dirname, 'invoices', `${invoiceNumber}.pdf`);

        try {
            // Ensure the directory exists
            fs.mkdirSync(path.dirname(filePath), { recursive: true });

            const writeStream = fs.createWriteStream(filePath);
            doc.pipe(writeStream);

            // Add content to the PDF
            doc.fontSize(20).text('Invoice', { align: 'center' });
            doc.moveDown();
            doc.fontSize(12).text(`Invoice Number: ${invoiceNumber}`);
            doc.text('Date: ' + invoiceDate);
            doc.moveDown();
            doc.text('Bill To:');
            doc.text('Customer Name');
            doc.text('12345 Example Street');
            doc.text('City, State, ZIP');
            doc.moveDown();
            doc.text('Item Description    Quantity    Price');
            doc.text('------------------------------------------------');
            doc.text('Item 1               2            $20.00');
            doc.text('Item 2               1            $10.00');
            doc.text('------------------------------------------------');
            doc.text('Total: $30.00', { align: 'right' });

            doc.end();

            writeStream.on('finish', () => resolve(filePath));
            writeStream.on('error', reject);

            console.log('Generated dummy PDF invoice at', filePath);
        } catch (err) {
            reject(err);
        }
    });
}

// Promise.all([
//     generateDummyPDFInvoice('inv-1', '2023-10-15' ),
//     generateDummyPDFInvoice('inv-2', '2023-10-16' ),
//     generateDummyPDFInvoice('inv-3', '2023-10-17' ),
// ])



export { generateDummyPDFInvoice };