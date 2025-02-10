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

        const v = await this.f(...this.a)
        this.event.emit('return', v);
        return v
    }

    async replay_signal(token: string, value: any): Promise<void> {
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

test('ak_call execute signal', async () => {
    let realFuncExecuted = false;
    const testFunc = async (a: number, b: number) => {
        realFuncExecuted = true;
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter)
    let p = _ak_call([testFunc, 1, 2])
    let v = await waiter.execute_signal(waiter.token)
    expect(v).toEqual(3)
    const r = await p;
    expect(r).toBe(3)
    expect(realFuncExecuted).toBeTruthy()
})

test('ak_call reply signal', async () => {
    let realFuncExecuted = false;
    const testFunc = async (a: number, b: number) => {
        realFuncExecuted = true;
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter)
    realFuncExecuted = false;
    let p = _ak_call([testFunc, 1, 2])
    await waiter.replay_signal(waiter.token, 3)
    expect(await p).toBe(3)
    expect(realFuncExecuted).toBeFalsy()
})


test('ak_call wrong token', async () => {
    let realFuncExecuted = false;
    const testFunc = async (a: number, b: number) => {
        realFuncExecuted = true;
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter)
    let p = _ak_call([testFunc, 1, 2])
    try {
        await waiter.replay_signal('wrong token', 3)
    } catch (e) {
        expect(e).toStrictEqual(new Error('tokens do not match'))
    }

})
