import express, { Request, Response } from 'express';
import config from './config.js';
import Storage from './InvoiceStorage.js';

/**
 * Interface representing the storage with methods for retrieving invoice data and total amount.
 */

/**
 * Starts the server and sets up the necessary routes for handling API requests.
 *
 * @param {Storage} storage - An object that provides methods to retrieve invoice data and total amount.
 * @param {number} [port=config.server.port] - The port number on which the server will listen. Defaults to the configured server port.
 * @return {void} Does not return a value.
 */
function startServer(storage: Storage, port: number = config.server.port): void {
    const app = express();

    app.get('/total', (req: Request, res: Response) => {
        res.json({ total: storage.getTotalAmount() });
    });

    app.get('/list', (req: Request, res: Response) => {
        res.json({ total: storage.getInvoices() });
    });

    app.get('/invoice/:id', (req: Request, res: Response) => {
        const invoice = storage.getInvoice(req.params.id);
        if (invoice) {
            res.json(invoice);
        } else {
            res.status(404).json({ error: 'Invoice not found' });
        }
    });

    app.listen(port, () => {
        console.log(`Query API listening on port ${port}`);
    });
}

export default startServer;