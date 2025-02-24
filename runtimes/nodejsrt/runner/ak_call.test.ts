import {Waiter, ak_call} from "./ak_call";
import {EventEmitter, once} from "node:events";

class mockWaiter implements Waiter{
    event: EventEmitter;
    f: Function
    a: any
    token: string

    constructor() {
        this.event = new EventEmitter();
        this.f = () => {}
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
}

test('ak_call execute and reply', async () => {
    let realFuncExecuted = false;
    const testFunc = async (a: number, b: number) => {
        realFuncExecuted = true;
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter)
    let p = _ak_call(testFunc, 1, 2)

    let v = await waiter.execute_signal(waiter.token)
    await waiter.reply_signal(waiter.token, v)

    expect(v).toEqual(3)
    expect(await p).toBe(3)
    expect(realFuncExecuted).toBeTruthy()
})


test('ak_call reply only', async () => {
    let realFuncExecuted = false;
    const testFunc = async (a: number, b: number) => {
        realFuncExecuted = true;
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter)
    let p = _ak_call(testFunc, 1, 2)
    await waiter.reply_signal(waiter.token, 3)

    expect(await p).toBe(3)
    expect(realFuncExecuted).toBeFalsy()
})

test('ak_call wrong token', async () => {
    const testFunc = async (a: number, b: number) => {
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter)
    _ak_call([testFunc, 1, 2])
    try {
        await waiter.reply_signal('wrong token', 3)
    } catch (e) {
        expect(e).toStrictEqual(new Error('tokens do not match'))
    }

})
