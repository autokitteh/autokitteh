import { EventEmitter } from "events";
import { Waiter } from "../../runtime/ak_call";
import { TextEncoder } from "util";

// Forward declaration to avoid circular import
export interface IDirectHandlerClient {
    activity: (request: any) => Promise<any>;
}

/**
 * DirectActivityWaiter implements the Waiter interface to simulate the activity flow
 * without going through gRPC. This implementation automatically triggers activity execution.
 */
export class DirectActivityWaiter implements Waiter {
    private runnerId: string;
    private runId: string = "";
    private event = new EventEmitter();
    private originalConsoleLog: typeof console.log;
    private client: IDirectHandlerClient | null = null;
    
    // Storage for pending activities
    private pendingActivities: Map<string, {
        f: (...args: any[]) => unknown,
        args: unknown[]
    }> = new Map();
    
    constructor(runnerId: string) {
        this.runnerId = runnerId;
        this.originalConsoleLog = console.log;
    }

    /**
     * Set the client reference after initialization
     */
    setClient(client: IDirectHandlerClient): void {
        this.client = client;
    }

    /**
     * Register a function call and automatically trigger activity execution
     */
    async wait(f: (...args: any[]) => unknown, args: unknown[], token: string): Promise<unknown> {
        this.originalConsoleLog(`[DirectWaiter] Registering function: ${f.name} with token: ${token}`);
        
        // Store the function and arguments for later execution
        this.pendingActivities.set(token, { f, args });
        
        // If we have a client, trigger the activity request - this will lead to automatic execution
        if (this.client) {
            this.originalConsoleLog(`[DirectWaiter] Triggering activity request for token: ${token}`);
            await this.client.activity({
                runnerId: this.runnerId,
                data: new TextEncoder().encode(JSON.stringify({ token }))
            });
        } else {
            this.originalConsoleLog(`[DirectWaiter] No client set, activity request not triggered for token: ${token}`);
        }
        
        // Return a promise that will be resolved when reply_signal is called with this token
        return new Promise((resolve) => {
            this.event.once(`return:${token}`, resolve);
        });
    }

    /**
     * Execute a pending function with the given token
     */
    async execute_signal(token: string): Promise<unknown> {
        this.originalConsoleLog(`[DirectWaiter] execute_signal called with token: ${token}`);
        
        const activity = this.pendingActivities.get(token);
        if (!activity) {
            throw new Error(`No activity found for token: ${token}`);
        }
        
        try {
            // THIS is where the function actually gets executed
            const result = await activity.f(...activity.args);
            this.originalConsoleLog(`[DirectWaiter] Function executed successfully, result:`, result);
            return result;
        } catch (error) {
            this.originalConsoleLog(`[DirectWaiter] Function execution failed:`, error);
            throw error;
        }
    }

    /**
     * Deliver the result back to the waiting caller
     */
    async reply_signal(token: string, value: unknown): Promise<void> {
        this.originalConsoleLog(`[DirectWaiter] reply_signal called with token: ${token}, value:`, value);
        this.event.emit(`return:${token}`, value);
        // Clean up
        this.pendingActivities.delete(token);
    }

    /**
     * Get all pending activity tokens
     */
    getPendingActivityTokens(): string[] {
        return Array.from(this.pendingActivities.keys());
    }

    setRunnerId(id: string): void {
        this.originalConsoleLog(`[DirectWaiter] Setting runnerId: ${id}`);
        this.runnerId = id;
    }

    setRunId(id: string): void {
        this.originalConsoleLog(`[DirectWaiter] Setting runId: ${id}`);
        this.runId = id;
    }

    getRunId(): string {
        return this.runId;
    }

    async done(): Promise<void> {
        this.originalConsoleLog(`[DirectWaiter] done called`);
    }
} 