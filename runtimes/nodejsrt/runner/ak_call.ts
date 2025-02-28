import {randomUUID} from "node:crypto";
import {EventEmitter, once} from "node:events";
import { Client } from "@connectrpc/connect";
import {HandlerService, DoneRequest} from "./pb/autokitteh/user_code/v1/handler_svc_pb";


export interface Waiter {
    wait:  (f: Function, v: any, token: string) => Promise<any>
    execute_signal: (token: string) => Promise<any>
    reply_signal: (token: string, value: any) => Promise<void>
    setRunnerId: (id: string) => void
    setRunId: (id: string) => void
    getRunId: () => string
    done: () => void
}

export class ActivityWaiter implements Waiter{
    event: EventEmitter;
    f: Function
    a: any
    token: string
    client: Client<typeof HandlerService>
    runnerId: string
    runId: string

    constructor(client: Client<typeof HandlerService>, runnerId: string) {
        this.client = client;
        this.event = new EventEmitter();
        this.f = () => {}
        this.token = ""
        this.runnerId = runnerId;
        this.runId = ""
    }

    setRunId(id: string): void {
        this.runId = id;
    }

    getRunId(): string {
        return this.runId
    }

    async done(): Promise<void> {
        const encoder = new TextEncoder();
        const r: DoneRequest = {
            $typeName:"autokitteh.user_code.v1.DoneRequest",
            runnerId: this.runnerId,
            error: '',
            traceback: [],
            result: {
                $typeName: "autokitteh.values.v1.Value",
                custom: {
                    $typeName:"autokitteh.values.v1.Custom",
                    data: encoder.encode(JSON.stringify({results: "yay"})),
                    executorId: this.runId,
                }
            }
        }
        await this.client.done(r);
    }

    setRunnerId(id: string): void {
        this.runnerId = id;
    }

    async execute_signal(token: string): Promise<any> {
        if (token != this.token) {
            throw new Error('tokens do not match')
        }

        return await this.f(...this.a)
    }

    async reply_signal(token: string, value: any): Promise<void> {
        if (token != this.token) {
            throw new Error('tokens do not match')
        }

        this.event.emit('return', value);
    }

    async wait(f: Function, v: any, token: string): Promise<any> {
        this.f = f
        this.a = v
        this.token = token
        const encoder = new TextEncoder()

        await this.client.activity({
            runnerId: this.runnerId,
            data: encoder.encode(JSON.stringify({token}))
        })


        const resp = await this.client.activity({
            runnerId: this.runnerId,
            data: encoder.encode(JSON.stringify({token: token})),
            callInfo: {
                function: f.name,
                args: [],
            }
        });
        console.log("activity resp", resp, "call", f.name)
        const r = (await once(this.event, 'return'))[0]
        console.log("got return value", r)
        return r
    }
}

export const ak_call = (waiter: Waiter) => {
    return async (...args: any) => {
        if (typeof args[0] === "object") {
            const ak_call_obj = () => {
                return async (...args: any) => {
                    let o = args[0];
                    let m = args[1];
                    let m_args: any = []
                    if (args.length > 2) {
                        m_args = args.slice(1);
                    }


                    if (o.ak_call === undefined) {
                        console.log("direct obj call", o.name, m ,m_args)
                        return await o[m](...m_args);
                    }

                    console.log("remote obj call", f.name, f_args);
                    const results = await waiter.wait(f, f_args, randomUUID());
                    console.log("got obj call results", results)
                    return results;
                }
            }
            return ak_call_obj()(...args);
        }
        let f = args[0];
        let f_args: any = []
        if (args.length > 1) {
            f_args = args.slice(1);
        }


        if (f.ak_call === undefined || f.name == "authenticate") {
            console.log("direct call", f.name, f_args)
            return await f(...f_args);
        }

        console.log("remote func call", f.name, f_args);
        const results = await waiter.wait(f, f_args, randomUUID());
        console.log("got func results", results)
        return results;
    }
}
