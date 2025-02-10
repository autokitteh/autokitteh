import {randomUUID} from "node:crypto";
import {EventEmitter, once} from "node:events";
import {Client} from "@connectrpc/connect";
import {HandlerService} from "./pb/autokitteh/user_code/v1/handler_svc_pb";

interface FunctionsCache {
    [key: string]: Function
}

interface ResultsCache {
    [key: string]: EventEmitter;
}

export const ak_call = (client: Client<typeof HandlerService>, runnerId: string, functionsCache: FunctionsCache, resultsCache: ResultsCache) => {
    return async (...args: any) => {
        let f = args[0];
        let f_args = args.slice(1)

        if (f.ak_call !== true) {
            return await f(...f_args);
        }

        const uuid = "f_" + randomUUID().toString().replace(/-/g, "");
        functionsCache[uuid] = f
        resultsCache[uuid] = new EventEmitter()
        console.log("adding function:", f.name, uuid)

        let data = {
            f: uuid,
            f_args: []
        }

        if (f_args) {
            data.f_args = f_args
        }

        const serializedData = JSON.stringify(data)

        const encoder = new TextEncoder()
        const resp = await client.activity({runnerId: runnerId, data: encoder.encode(serializedData), callInfo: {
                function: uuid,
                args: [],
            }});

        console.log("activity call resp", resp, "args", args);
        const results = await once(resultsCache[uuid], 'return');
        console.log("got results", results)
        return results;
    }
}
