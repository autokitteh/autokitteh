import {Sandbox} from "./sandbox";

test('typescript hooking async', async () => {
    let hookCalled = false
    const hookFunc = async (...args: any): Promise<string> => {
        let f = args[0];
        let f_args = args.slice(1)
        return await f(...f_args);
    }

    const sandbox = new Sandbox("/Users/adiludmer/GolandProjects/autokitteh/runtimes/nodejsrt/runner/test_data/replay-demo", hookFunc);
    await sandbox.loadFile("/Users/adiludmer/GolandProjects/autokitteh/runtimes/nodejsrt/runner/test_data/replay-demo/main.ts");
    await sandbox.run("on_event", [1], () => {})
    expect(hookCalled).toBe(true);
});


