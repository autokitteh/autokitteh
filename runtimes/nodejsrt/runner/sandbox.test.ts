import {Sandbox} from "./sandbox";

test('typescript hooking async', async () => {
    let hookCalled = false
    const hookFunc = async (...args: any): Promise<any> => {
        let f = args[0];
        if (typeof f === "function") {
            let f_args = args.slice(1)
            let out = await f(...f_args);
            console.log(out)
            return out;
        }
        else if (typeof f === "object") {
            let o = args[0];
            let m = args[1];
            let m_args = args.slice(2)
            let out = await o[m](...m_args);
            console.log(out)
            return out;
        }
    }

    const sandbox = new Sandbox("/Users/adiludmer/GolandProjects/autokitteh/runtimes/nodejsrt/runner/test_data/replay-demo", hookFunc);
    await sandbox.loadFile("/Users/adiludmer/GolandProjects/autokitteh/runtimes/nodejsrt/runner/test_data/replay-demo/main.ts");
    await sandbox.run("on_event", [1], () => {})
    expect(hookCalled).toBe(true);
});


