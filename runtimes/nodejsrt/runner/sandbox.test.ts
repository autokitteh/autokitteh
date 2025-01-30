import {Sandbox} from "./sandbox";

test('typescript hooking async', async () => {
    let hookCalled = false
    const hookFunc = async (...args: any): Promise<string> => {
        let f = args[0];
        let f_args = args.slice(1)

        if (f.ak_call === true) {
            hookCalled = true;
            return "yay"
        }

        return await f(...f_args);
    }

    const sandbox = new Sandbox(hookFunc);
    await sandbox.loadFile("test_data/demo_async/main.ts");
    await sandbox.run("on_event()")
    expect(hookCalled).toBe(true);
});


