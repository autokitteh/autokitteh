import {ActivityResponse, AKClient} from "./pb/ak";
import {credentials, type ServiceError} from "@grpc/grpc-js";
import {EventEmitter, once} from "node:events";
import { randomUUID } from "node:crypto";

interface FunctionsCache {
    [key: string]: Function
}

interface ResultsCache {
    [key: string]: EventEmitter
}

export const functionsCache: FunctionsCache = {}
export const resultsCache: ResultsCache = {}

const sendActivityAndGetResults = async (f: string, args: any[]) => {
    const client = new AKClient(
        'localhost:9998',
        credentials.createInsecure()
    )

    const encoder = new TextEncoder();
    const data = encoder.encode(JSON.stringify({"function": f, "args": args}));
    resultsCache[f] = new EventEmitter()
    client.activity(
        {callInfo: {
                function: f,
                args: args,
                kwargs: {}
            }, data: data},
        (error: ServiceError | null, response: ActivityResponse) => {
            console.log("activity reply", response, error);
        })

    client.done({ result: f, error: "", traceback: []}, (err, result) => {
        console.log("done reply", result, err)
    })

    return (await once(resultsCache[f], "result"))[0]
}

export const ak_call = async (...args: any) => {
    let f = args[0];
    let f_args = args.slice(1)

    if (f.ak_call !== true) {
        return await f(...f_args);
    }

    const uuid = randomUUID()
    functionsCache[uuid] = f

    return sendActivityAndGetResults(uuid, f_args)
}
