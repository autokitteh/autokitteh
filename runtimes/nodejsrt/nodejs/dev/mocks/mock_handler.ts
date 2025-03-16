import {HandlerService} from "../../runtime/pb/autokitteh/user_code/v1/handler_svc_pb";
import {ConnectRouter, createClient} from "@connectrpc/connect";
import {createGrpcTransport} from "@connectrpc/connect-node";


export const createService = (codeDir: string, runnerId: string, workerAddress: string) => {
    console.log("--code-dir=", codeDir, "--runner-id=", runnerId, "--worker-address=", workerAddress);
    const transport = createGrpcTransport({
        baseUrl: `http://localhost:1111`,
    });

    const client = createClient(HandlerService, transport);

    return (router: ConnectRouter) => router.service(HandlerService, {
        health: undefined,
        isActiveRunner: undefined,
        refreshOAuthToken: undefined,
        activity: async (req, ctx) => {
            return {};
        },
        done: async (req, ctx) => {
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
}
