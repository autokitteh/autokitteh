import {HandlerService} from "./pb/autokitteh/user_code/v1/handler_svc_pb";
import { createConnectTransport } from "@connectrpc/connect-node";
import { createClient } from "@connectrpc/connect";
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
    const transport = createConnectTransport({
        baseUrl: "http://localhost:8080",
        httpVersion: "1.1"
    });

    resultsCache[f] = new EventEmitter()
    const client = createClient(HandlerService, transport);
    await client.activity({})
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
