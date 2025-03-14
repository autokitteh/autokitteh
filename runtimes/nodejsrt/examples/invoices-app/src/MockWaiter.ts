import {Waiter} from "./ak/ak_call";
import {EventEmitter, once} from "node:events";
import {toReaderCall} from "ts-proto/build/src/types";

export class MockWaiter implements Waiter {
    event: EventEmitter;
    f: Function
    a: any
    token: string

    constructor() {
        this.event = new EventEmitter();
        this.f = () => {
        }
        this.token = ""
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
        const r = await once(this.event, 'return')
        return r[0]
    }

    done(): void {
    }

    getRunId(): string {
        return "";
    }

    setRunId(id: string): void {
    }

    setRunnerId(id: string): void {
    }
}
