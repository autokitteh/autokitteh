import {Waiter, ak_call} from "../runtime/ak_call";
import {EventEmitter, once} from "node:events";

// Define a type for the function
type AnyFunction = (...args: unknown[]) => unknown;

class mockWaiter implements Waiter {
    event: EventEmitter;
    f: AnyFunction;
    a: unknown[];
    token: string;

    constructor() {
        this.event = new EventEmitter();
        this.f = () => {};
        this.a = [];
        this.token = "";
    }

    async execute_signal(token: string): Promise<unknown> {
        if (token != this.token) {
            throw new Error('tokens do not match');
        }

        return await this.f(...this.a);
    }

    async reply_signal(token: string, value: unknown): Promise<void> {
        if (token != this.token) {
            throw new Error('tokens do not match');
        }

        this.event.emit('return', value);
    }

    async wait(f: AnyFunction, v: unknown[], token: string): Promise<unknown> {
        this.f = f;
        this.a = v;
        this.token = token;
        const r = await once(this.event, 'return');
        return r[0];
    }

    done(): void {
    }

    getRunId(): string {
        return "";
    }

    setRunId(_id: string): void {
    }

    setRunnerId(_id: string): void {
    }
}

test('ak_call execute and reply', async () => {
    let realFuncExecuted = false;
    const testFunc = async (a: number, b: number) => {
        realFuncExecuted = true;
        return a + b;
    };

    const waiter = new mockWaiter();
    const _ak_call = ak_call(waiter, process.cwd());
    const p = _ak_call(testFunc, 1, 2);

    const v = await waiter.execute_signal(waiter.token);
    await waiter.reply_signal(waiter.token, v);

    expect(v).toEqual(3);
    expect(await p).toBe(3);
    expect(realFuncExecuted).toBeTruthy();
});

test('ak_call reply only', async () => {
    let realFuncExecuted = false;
    const testFunc = async (a: number, b: number) => {
        realFuncExecuted = true;
        return a + b
    }

    testFunc.ak_call = true;

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter, process.cwd())
    const p = _ak_call(testFunc, 1, 2)
    await waiter.reply_signal(waiter.token, 3)

    expect(await p).toBe(3)
    expect(realFuncExecuted).toBeFalsy()
})

test('ak_call wrong token', async () => {
    const testFunc = async (a: number, b: number) => {
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter, process.cwd())
    _ak_call([testFunc, 1, 2])
    try {
        await waiter.reply_signal('wrong token', 3)
    } catch (e) {
        expect(e).toStrictEqual(new Error('tokens do not match'))
    }

})
