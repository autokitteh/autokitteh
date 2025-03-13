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
    f: AnyFunction;
    a: unknown[];
    token: string;
    client: Client<typeof HandlerService>;
    runnerId: string;
    runId: string;

    constructor(client: Client<typeof HandlerService>, runnerId: string) {
        this.client = client;
        this.event = new EventEmitter();
        this.f = () => {};
        this.a = [];
        this.token = "";
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
        if (token !== this.token) {
            throw new Error('tokens do not match');
        }
        return await this.f(...this.a);
    }

    async reply_signal(token: string, value: unknown): Promise<void> {
        if (token !== this.token) {
            throw new Error('tokens do not match');
        }
        this.event.emit('return', value);
    }

    async wait(f: AnyFunction, v: unknown[], token: string): Promise<unknown> {
        this.f = f;
        this.a = v;
        this.token = token;
        const encoder = new TextEncoder();

        await this.client.activity({
            runnerId: this.runnerId,
            data: encoder.encode(JSON.stringify({token}))
        });

        await this.client.activity({
            runnerId: this.runnerId,
            data: encoder.encode(JSON.stringify({token})),
            callInfo: {
                function: f.name,
                args: [],
            }
        });
        console.log("activity resp", "call", f.name);
        const r = (await once(this.event, 'return'))[0];
        console.log("got return value", r);
        return r;
    }
}

export const ak_call = (waiter: Waiter, projectRoot: string) => {
    function isInternalFunction(func: AnyFunction): boolean {
        // Case 1: Already marked as internal
        if ((func as unknown as { _ak_direct_call?: boolean })._ak_direct_call === true) {
            return true;
        }

        // Case 2: Native methods
        if (func.toString().includes('[native code]')) {
            return false;
        }

        // Case 3: Check module path of the function
        const moduleFile = module.filename;
        if (moduleFile && 
            moduleFile.startsWith(projectRoot) && 
            !moduleFile.includes('node_modules')) {
            return true;
        }

        return false;
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
