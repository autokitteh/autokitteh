import {createRouterTransport, createClient, ConnectRouter} from "@connectrpc/connect";
import { createService } from "./server";

import {RunnerService } from "./pb/autokitteh/user_code/v1/runner_svc_pb";
import {Sandbox} from "./sandbox";
import {ActivityWaiter} from "./ak_call";
import {HandlerService} from "./pb/autokitteh/user_code/v1/handler_svc_pb";
import {EventEmitter, once} from "node:events";

const mockSandbox = new Sandbox("", async (f: Function, args: any) => { return await f(...args)})
const doneEvent = new EventEmitter()


const mockHandlerService = (router: ConnectRouter) => router.service(HandlerService, {
    health: undefined,
    isActiveRunner: undefined,
    refreshOAuthToken: undefined,
    activity: async (req, ctx) => {
        return {};
    },
    done: async (req, ctx) => {
        doneEvent.emit("done")
        return {};
    },
    log: async (req, ctx) => {
        return {};
    },
    print: async (req, ctx) => {
        return {};
    },
    sleep: async (req, ctx) => {
        return {};
    },
    subscribe: async (req, ctx) => {
        return {};
    },
    nextEvent: async (req, ctx) => {
        return {};
    },
    unsubscribe: async (req, ctx) => {
        return {};
    },
    startSession: async (req, ctx) => {
        return {};
    },
    encodeJWT: async (req, ctx) => {
        return {};
    }
});

const mockTransport = createRouterTransport(mockHandlerService)
const mockHandlerClient  = createClient(HandlerService, mockTransport);

const waiter =  new ActivityWaiter(mockHandlerClient, "test")


test('full flow', async () => {
    const encoder = new TextEncoder();
    const mockTransport = createRouterTransport(createService("","", mockSandbox,waiter))
    const client = createClient(RunnerService, mockTransport);
    
    client.start({
        entryPoint: "main.ts:on_event",
        event: {data: encoder.encode(JSON.stringify({"a": 1}))},
    })

    await once(doneEvent, "done")
});
