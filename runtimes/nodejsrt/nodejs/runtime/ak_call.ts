import {randomUUID} from "node:crypto";
import {EventEmitter, once} from "node:events";
import { Client } from "@connectrpc/connect";
import {HandlerService, DoneRequest} from "./pb/autokitteh/user_code/v1/handler_svc_pb";

type AnyFunction = (...args: unknown[]) => unknown;

export interface Waiter {
    wait: (f: AnyFunction, v: unknown[], token: string) => Promise<unknown>;
    execute_signal: (token: string) => Promise<unknown>;
    reply_signal: (token: string, value: unknown) => Promise<void>;
    setRunnerId: (id: string) => void;
    setRunId: (id: string) => void;
    getRunId: () => string;
    done: () => void;
}

export class ActivityWaiter implements Waiter {
    event: EventEmitter;
    private pendingActivities = new Map<string, {
        f: AnyFunction;
        a: unknown[];
    }>();
    client: Client<typeof HandlerService>;
    runnerId: string;
    runId: string;

    constructor(client: Client<typeof HandlerService>, runnerId: string) {
        this.client = client;
        this.event = new EventEmitter();
        this.runnerId = runnerId;
        this.runId = "";
    }

    setRunId(id: string): void {
        this.runId = id;
    }

    getRunId(): string {
        return this.runId;
    }

    async done(): Promise<void> {
        const encoder = new TextEncoder();
        const r: DoneRequest = {
            $typeName: "autokitteh.user_code.v1.DoneRequest",
            runnerId: this.runnerId,
            error: '',
            traceback: [],
            result: {
                $typeName: "autokitteh.values.v1.Value",
                custom: {
                    $typeName: "autokitteh.values.v1.Custom",
                    data: encoder.encode(JSON.stringify({results: "yay"})),
                    executorId: this.runId,
                }
            }
        };
        await this.client.done(r);
    }

    setRunnerId(id: string): void {
        this.runnerId = id;
    }

    async execute_signal(token: string): Promise<unknown> {
        const activity = this.pendingActivities.get(token);
        if (!activity) {
            throw new Error('tokens do not match');
        }
        try {
            const result = await activity.f(...activity.a);
            return result;
        }
        catch (err) {
            throw err;
        }

    }

    async reply_signal(token: string, value: unknown): Promise<void> {
        this.event.emit(`return:${token}`, value);
    }

    async wait(f: AnyFunction, v: unknown[], token: string): Promise<unknown> {
        this.pendingActivities.set(token, { f, a: v });
        const encoder = new TextEncoder();

        await this.client.activity({
            runnerId: this.runnerId,
            data: encoder.encode(JSON.stringify({token})),
            callInfo: {
                function: f.name,
                args: [],
            }
        });
        console.log("activity resp", "call", f.name);
        const r = (await once(this.event, `return:${token}`))[0];
        console.log("got return value", r);
        this.pendingActivities.delete(token);
        return r;
    }
}

export const ak_call = (waiter: Waiter) => {
    // Only keeping the _ak_direct_call check for functions explicitly marked as internal
    function isInternalFunction(func: AnyFunction): boolean {
        // Check if function is explicitly marked as internal
        return (func as unknown as { _ak_direct_call?: boolean })._ak_direct_call === true;
    }

    return async (...args: unknown[]): Promise<unknown> => {
        if (typeof args[0] === "object") {
            const ak_call_obj = () => {
                return async (...args: unknown[]): Promise<unknown> => {
                    const obj = args[0] as Record<string, AnyFunction>;
                    const method = args[1] as string;
                    const methodArgs = args.length > 2 ? args.slice(2) : [];

                    const func = obj[method];
                    if (!func) {
                        throw new Error(`Method ${method} not found`);
                    }

                    // Check for explicitly marked internal functions
                    if (isInternalFunction(func)) {
                        console.log("direct obj call", method, methodArgs);
                        return await func.apply(obj, methodArgs);
                    }

                    console.log("remote obj call", method, methodArgs);
                    const results = await waiter.wait(func.bind(obj), methodArgs, randomUUID());
                    console.log("got obj call results", results);
                    return results;
                };
            };
            return ak_call_obj()(...args);
        }

        const func = args[0] as AnyFunction;
        const funcArgs = args.length > 1 ? args.slice(1) : [];

        // Check for explicitly marked internal functions
        if (isInternalFunction(func)) {
            console.log("direct call", func.name, funcArgs);
            return await func(...funcArgs);
        }

        console.log("remote func call", func.name, funcArgs);
        const results = await waiter.wait(func, funcArgs, randomUUID());
        console.log("got func results", results);
        return results;
    };
};
