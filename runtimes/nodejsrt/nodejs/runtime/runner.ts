import { createClient } from "@connectrpc/connect";
import { ConnectRouter } from "@connectrpc/connect";
import { TextDecoder, TextEncoder } from "util";
import { EventEmitter } from "events";
import * as path from "path";
import * as fs from "fs";

import { HandlerService } from "./pb/autokitteh/user_code/v1/handler_svc_pb";
import { RunnerService, Export } from "./pb/autokitteh/user_code/v1/runner_svc_pb";
import { ActivityWaiter } from "./ak_call";
import { listExports } from "../common/ast_utils";
import { initializeGlobals } from "./runtime";
import { safeSerialize } from "../common/serializer";

type AkCallFunction = (...args: unknown[]) => Promise<unknown>;

// Use module augmentation instead of namespace
declare global {
    interface Global {
        ak_call: AkCallFunction;
    }
}

const START_TIMEOUT = 10000; // 10 seconds
const HEALTH_CHECK_INTERVAL = 10000; // 10 seconds
const SERVER_GRACE_TIMEOUT = 3000; // 3 seconds, matching Python's SERVER_GRACE_TIMEOUT

class ActivityError extends Error {
    constructor(message: string) {
        super(message);
        this.name = 'ActivityError';
    }
}

// Similar to Python's Result namedtuple
interface Result {
    value: unknown;
    error: Error | null;
    traceback: { name: string; filename: string; lineno: number; code: string }[];
}

function formatError(err: unknown): { message: string, traceback: { name: string; filename: string; lineno: number; code: string }[] } {
    const error = err instanceof Error ? err : new Error(String(err));
    const stack = error.stack || '';
    const frames = stack.split('\n').map(line => {
        const match = line.match(/at\s+(.+?)\s+\((.+?):(\d+):\d+\)/);
        if (!match) return null;
        const [, name, filename, linenoStr] = match;
        return {
            name,
            filename,
            lineno: parseInt(linenoStr, 10),
            code: line.trim()
        };
    }).filter((f): f is { name: string; filename: string; lineno: number; code: string } => f !== null);

    return {
        message: error.message,
        traceback: frames
    };
}

function createResult(value: unknown = null, error: unknown = null): Result {
    if (error) {
        const { message, traceback } = formatError(error);
        return { value: null, error: new Error(message), traceback };
    }
    return { value, error: null, traceback: [] };
}

function fixHttpBody(data: unknown): void {
    if (typeof data !== 'object' || data === null) return;

    const eventData = data as Record<string, unknown>;
    const body = eventData.body;

    if (typeof body !== 'object' || body === null) return;

    const bodyData = body as Record<string, unknown>;
    if (typeof bodyData.bytes === 'string') {
        try {
            bodyData.bytes = Buffer.from(bodyData.bytes, 'base64');
        } catch (err) {
            console.warn('Failed to decode base64 body:', err);
        }
    }
}

export default class Runner {
    private readonly id: string;
    private readonly codeDir: string;
    private readonly client: ReturnType<typeof createClient<typeof HandlerService>>;
    private readonly waiter: ActivityWaiter;
    private readonly events = new EventEmitter();
    private readonly encoder = new TextEncoder();
    private readonly decoder = new TextDecoder();
    private isStarted = false;
    private healthcheckTimer?: NodeJS.Timeout;
    private startTimer?: NodeJS.Timeout;
    private originalConsoleLog: typeof console.log;
    private isShuttingDown = false;
    private _isRunningHealthCheck: boolean = false;

    constructor(id: string, codeDir: string, client: ReturnType<typeof createClient<typeof HandlerService>>) {
        this.id = id;
        this.codeDir = codeDir;
        this.client = client;
        this.waiter = new ActivityWaiter(client, id);
        this.originalConsoleLog = console.log;

        // Setup start timeout
        this.startTimer = setTimeout(() => this.stopIfStartNotCalled(), START_TIMEOUT);

        // Intercept console.log with typed parameters
        console.log = (...args: unknown[]) => {
            this.akPrint(...args);
        };

        // Handle graceful shutdown
        this.setupSignalHandlers();
    }

    private setupSignalHandlers(): void {
        const signals: NodeJS.Signals[] = ['SIGTERM', 'SIGINT'];
        signals.forEach(signal => {
            process.on(signal, () => {
                console.log(`Received ${signal}, shutting down...`);
                this.gracefulShutdown();
            });
        });

        // Handle uncaught errors
        process.on('uncaughtException', (err) => {
            console.error('Uncaught exception:', err);
            this.gracefulShutdown(1);
        });

        process.on('unhandledRejection', (reason) => {
            console.error('Unhandled rejection:', reason);
            this.gracefulShutdown(1);
        });
    }

    private async gracefulShutdown(exitCode = 0): Promise<void> {
        if (this.isShuttingDown) return;
        this.isShuttingDown = true;

        console.log('Starting graceful shutdown...');

        try {
            // Stop the runner
            this.stop();

            // Give some time for cleanup
            await new Promise(resolve => setTimeout(resolve, SERVER_GRACE_TIMEOUT));

            console.log('Shutdown complete');
        } catch (err) {
            console.error('Error during shutdown:', err);
            exitCode = 1;
        } finally {
            process.exit(exitCode);
        }
    }

    private async akPrint(...args: unknown[]) {
        const message = args.map(arg => String(arg)).join(' ');
        this.originalConsoleLog(message);

        try {
            await this.client.print({
                runnerId: this.id,
                message
            });
        } catch (err) {
            this.originalConsoleLog('Failed to send print message:', err);
        }
    }

    private stopIfStartNotCalled() {
        if (!this.isStarted) {
            console.error(`Start not called after ${START_TIMEOUT}ms, terminating`);
            this.stop();
        }
    }

    private async startHealthCheck() {
        const maxRetries = 10;
        const retryDelay = 1000; // 1 second
        this._isRunningHealthCheck = true;

        // Initial active runner check with retry logic
        while (this._isRunningHealthCheck) {
            let retries = 0;
            let success = false;

            while (retries < maxRetries && !success) {
                try {
                    const response = await this.client.isActiveRunner({
                        runnerId: this.id
                    });

                    // Handle response error
                    if (response.error) {
                        console.error('Active runner check failed:', response.error);
                        retries++;

                        if (retries >= maxRetries) {
                            console.error(`Maximum retries (${maxRetries}) reached. Stopping runner.`);
                            this.stop();
                            return; // Exit the method completely
                        }

                        console.log(`Retrying active runner check (${retries}/${maxRetries})...`);
                        await new Promise(resolve => setTimeout(resolve, retryDelay));
                        continue;
                    }

                    // Check if runner is active
                    if (!response.isActive) {
                        console.log('Runner is no longer active, stopping...');
                        this.stop();
                        return; // Exit the method completely
                    }

                    // Success case
                    success = true;
                    console.log('Active runner check successful');

                } catch (err) {
                    console.error('Active runner check error:', err);
                    retries++;

                    if (retries >= maxRetries) {
                        console.error(`Maximum retries (${maxRetries}) reached. Stopping runner.`);
                        this.stop();
                        return; // Exit the method completely
                    }

                    console.log(`Retrying active runner check (${retries}/${maxRetries})...`);
                    await new Promise(resolve => setTimeout(resolve, retryDelay));
                }
            }

            // Wait for next check interval before running another active check
            await new Promise(resolve => setTimeout(resolve, HEALTH_CHECK_INTERVAL));
        }
    }

    async start() {
        if (this.isStarted) {
            throw new Error("Runner already started");
        }

        console.log("Starting runner...");

        // Clear start timeout
        if (this.startTimer) {
            clearTimeout(this.startTimer);
        }

        // Start health checks
        console.log("Starting health checks...");
        this.startHealthCheck();

        // Start regular health check timer (separate from the active runner check)
        this.healthcheckTimer = setInterval(async () => {
            try {
                console.log("Sending health check...");
                await this.client.health({});
                console.log("Health check successful");
            } catch (err) {
                console.error("Health check failed:", err);
                this.stop();
            }
        }, HEALTH_CHECK_INTERVAL);

        // Make sure the timer doesn't prevent Node.js from exiting
        this.healthcheckTimer.unref();

        console.log("Setting up started event listener...");
        this.events.once("started", () => {
            console.log("Runner started event received");
            this.isStarted = true;
        });

        console.log("Runner initialization complete");
    }

    stop() {
        this._isRunningHealthCheck = false;

        if (this.healthcheckTimer) {
            clearInterval(this.healthcheckTimer);
            this.healthcheckTimer = undefined;
        }

        // Restore original console.log
        console.log = this.originalConsoleLog;

        this.events.emit("stop");
    }

    createService() {
        return (router: ConnectRouter) => router.service(RunnerService, {
            activityReply: async (req) => {
                if (!req.error) {
                    const data = req.result?.custom?.data;
                    if (data) {
                        const parsedData = JSON.parse(this.decoder.decode(data));
                        if (req.result?.custom?.executorId) {
                            this.waiter.setRunId(req.result.custom.executorId);
                        }
                        await this.waiter.reply_signal(parsedData.token, parsedData.results);
                        console.log("Activity reply:", parsedData);
                        return { error: "" };
                    }
                }
                throw new ActivityError(req.error || "Invalid activity reply");
            },

            execute: async (req) => {
                const data = this.decoder.decode(req.data);
                const execReq = JSON.parse(data);
                console.log("Execute request:", execReq);

                let result: Result;
                try {
                    const value = await this.waiter.execute_signal(execReq.token);
                    result = createResult(value);
                } catch (err) {
                    if (err instanceof ActivityError) {
                        result = createResult(null, err);
                    } else {
                        result = createResult(null, new ActivityError(String(err)));
                    }
                    return {
                        error: result.error?.message || "Unknown error",
                        traceback: result.traceback
                    };
                }

                const serialized = JSON.stringify({
                    token: execReq.token,
                    results: safeSerialize(result.value)
                });

                return {
                    error: "",
                    result: {
                        custom: {
                            executorId: this.waiter.getRunId(),
                            data: this.encoder.encode(serialized),
                            value: { string: { v: serialized } }
                        }
                    },
                    traceback: result.traceback
                };
            },

            start: async (req) => {
                this.events.emit("started");
                const data = req.event?.data;
                if (!data) {
                    return { error: "No event data provided", traceback: [] };
                }

                try {
                    const eventData = JSON.parse(this.decoder.decode(data));
                    fixHttpBody(eventData.data);

                    const [fileName, funcName] = (req.entryPoint || "").split(":");

                    if (!fileName || !funcName) {
                        return { error: "Invalid entry point format", traceback: [] };
                    }

                    // Initialize global ak_call
                    initializeGlobals(this.waiter);

                    // Import and execute user code
                    const modulePath = path.resolve(path.join(this.codeDir, fileName));
                    const module = await import(modulePath);

                    // Get the function and call it
                    const func = module[funcName];
                    if (typeof func !== 'function') {
                        return { error: `Function ${funcName} not found in ${fileName}`, traceback: [] };
                    }

                    // Call the function directly - ak_call is now globally available
                    const result = createResult();
                    try {
                        result.value = await func();
                    } catch (err) {
                        const { message, traceback } = formatError(err);
                        return {
                            error: message,
                            traceback
                        };
                    }

                    return { error: "", traceback: [] };
                } catch (err) {
                    const { message, traceback } = formatError(err);
                    return {
                        error: message,
                        traceback
                    };
                }
            },

            health: async () => {
                return { error: "" };
            },

            exports: async (req) => {
                if (!req.fileName) {
                    throw new Error("missing file name");
                }

                try {
                    const filePath = path.join(this.codeDir, req.fileName);
                    const exports = await this.discoverExports(filePath);
                    return { exports };
                } catch (err) {
                    throw new Error(`Failed to get exports: ${err}`);
                }
            },
        });
    }

    private async discoverExports(filePath: string): Promise<Export[]> {
        try {
            const code = await fs.promises.readFile(filePath, "utf-8");
            return await listExports(code, filePath);
        } catch (error) {
            console.error("Failed to discover exports:", error);
            return [];
        }
    }

    // Expose the emit method
    emit(event: string, ...args: unknown[]) {
        this.events.emit(event, ...args);
    }
}

