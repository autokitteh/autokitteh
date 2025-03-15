import {Sandbox} from "../runtime/sandbox";

test('typescript hooking async', async () => {
    let hookCalled = false
    const hookFunc = async (...args: any): Promise<any> => {
        hookCalled = true
        let f = args[0];
        if (f._ak_direct_call === true) {
            console.log("direct call", f.name)
        }
        if (typeof f === "function") {
            let f_args = args.slice(1)
            let out = await f(...f_args);
            let s = JSON.stringify(out)
            console.log("out:", f.name, s)
            return out;
        }
        else if (typeof f === "object") {
            let o = args[0];
            let m = args[1];
            let m_args = args.slice(2)
            let out = await o[m](...m_args);
            let s = JSON.stringify(out)


            console.log("out:", o.constructor.name.toString(), m, s)
            return out;
        }
    }

    const sandbox = new Sandbox("test_data/invoices-app", hookFunc);
    await sandbox.loadFile("src/main.ts");
    await sandbox.run("on_event", [1], () => {})
    expect(hookCalled).toBe(true);
});


