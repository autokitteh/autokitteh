export interface Waiter {
    wait:  (f: Function, v: any) => Promise<any>
    execute_signal: () => Promise<any>
    replay_signal: (value: any) => Promise<void>
}

export const ak_call = (waiter: Waiter) => {
    return async (args: any) => {
        let f = args[0];
        let f_args = args.slice(1)

        if (f.ak_call == false) {
            return await f(...f_args);
        }

        const results = await waiter.wait(f, f_args);
        console.log("got results", results)
        return results;
    }
}
