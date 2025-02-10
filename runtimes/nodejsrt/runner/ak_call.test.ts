import {Waiter, ak_call} from "./ak_call";
import {EventEmitter, once} from "node:events";

test('ak_call', async () => {
    class mockWaiter implements Waiter{
        event: EventEmitter;
        f: Function
        a: any

        constructor() {
            this.event = new EventEmitter();
            this.f = () => {}
        }

        async execute_signal(): Promise<any> {
            const v = await this.f(...this.a)
            this.event.emit('return', v);
            return v
        }

        async replay_signal(value: any): Promise<void> {
            this.event.emit('return', value);
        }

        async wait(f: Function, v: any): Promise<any> {
            this.f = f
            this.a = v
            const r = await once(this.event, 'return')
            return r[0]
        }
    }

    const testFunc = async (a: number, b: number) => {
        return a + b
    }

    const waiter = new mockWaiter()
    const _ak_call = ak_call(waiter)
    const p = _ak_call([testFunc, 1, 2])
    await waiter.execute_signal()
    const r = await p;
    expect(r).toBe(3)
})
