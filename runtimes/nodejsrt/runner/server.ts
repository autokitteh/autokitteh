import {
    ActivityReplyRequest,
    ActivityReplyResponse, ExecuteRequest, ExecuteResponse, Export,
    ExportsRequest, ExportsResponse,
    RunnerService, StartRequest, StartResponse
} from "./pb/autokitteh/user_code/v1/runner_svc_pb";

import {ak_call, functionsCache, resultsCache} from "./ak_call";
import {Sandbox} from "./sandbox";
import fs from "fs";
import {listExports, Symbol} from "./ast_utils";

import type { ConnectRouter } from "@connectrpc/connect";
import {HealthRequest, HealthResponse} from "../../../proto/gen/ts/autokitteh/runner_manager/v1/runner_manager_svc_pb";

export default (router: ConnectRouter) => router.service(RunnerService, {
    async activityReply(req: ActivityReplyRequest) : Promise<ActivityReplyResponse> {
        console.log(req);
        return {error: "", $typeName: "autokitteh.user_code.v1.ActivityReplyResponse"}
    },

    async exports(req: ExportsRequest): Promise<ExportsResponse> {
        const codeDir = process.env["CODE_DIR"];
        const filePath = `${codeDir}/${req.fileName}`
        let exports: Export[] = []

        try {
            const code = await fs.promises.readFile(filePath, "utf-8")
            exports = await listExports(code, filePath)
        }
        catch (error) {
            console.log(error)
        }

        return {$typeName: "autokitteh.user_code.v1.ExportsResponse", exports, error:""}
    },

    async execute(req: ExecuteRequest): Promise<ExecuteResponse> {
        const decoder = new TextDecoder();
        const data = decoder.decode(req.data)
        const execReq = JSON.parse(data)
        let results = await functionsCache[execReq.function](...execReq.args)
        return {
            $typeName: "autokitteh.user_code.v1.ExecuteResponse",
            error: "",
            result: {
                $typeName:"autokitteh.values.v1.Value",
                string: {
                    $typeName:"autokitteh.values.v1.String",
                    v: JSON.stringify(results),
                }
            },
            traceback: []
        }
    },

    async health(req: HealthRequest): Promise<HealthResponse> {
        return {error: ""}
    },

    async start(req: StartRequest): Promise<StartResponse> {
        return {$typeName: "autokitteh.user_code.v1.StartResponse", error:"", traceback: []}
    }
});

//
//
// export const activityReply: RunnerService['activityReply'] = (call: { request: ActivityReplyRequest; }, callback: (arg0: null, arg1: ActivityReplyResponse) => void) => {
//     callback(null, {error: "", $typeName: "autokitteh.user_code.v1.ActivityReplyResponse"});
// }
//
// export const start: RuntimeServer['start'] = async (call: { request: StartRequest; }, callback: (arg0: null, arg1: StartResponse) => void) => {
//     const decoder = new TextDecoder();
//     const request: StartRequest = call.request;
//     const args = decoder.decode(request.event?.data)
//     const [fileName, functionName] = request.entryPoint.split(":")
//
//     const sandbox = new Sandbox(ak_call)
//     const codeDir = process.env["CODE_DIR"];
//     await sandbox.loadFile(`${codeDir}/${fileName}`)
//     const code = `${functionName}(...${args})`
//
//     sandbox.run(code) // TODO: handle errors
//     callback(null, { traceback: [], error: ""});
// }
//
// export const listExports: RuntimeServer['exports'] = async (call: { request: ExportsRequest; }, callback: (arg0: null, arg1: ExportsResponse) => void) => {
//     const request: ExportsRequest = call.request;
//     const codeDir = process.env["CODE_DIR"];
//     const filePath = `${codeDir}/${request.fileName}`
//     let symbols: Symbol[] = []
//
//     try {
//         const code = await fs.promises.readFile(filePath, "utf-8")
//         symbols = await listSymbols(code, filePath)
//     }
//     catch (error) {
//         console.log(error)
//     }
//     const reply: ExportsResponse = { exports: symbols, error: "" };
//     callback(null, reply);
// };
//
// export const execute: RuntimeServer['execute'] = async (call: { request: ExecuteRequest; }, callback: (arg0: null, arg1: ExecuteResponse) => void) => {
//     const request: ExecuteRequest = call.request;
//     const decoder = new TextDecoder(); // Default is "utf-8"
//     const encoder = new TextEncoder();
//     const data = decoder.decode(request.data)
//     const execReq = JSON.parse(data)
//     let results = await functionsCache[execReq.function](...execReq.args)
//     console.log("execute:", execReq)
//     resultsCache[execReq.function].emit("result", results)
//     const reply: ExecuteResponse = { result: encoder.encode(JSON.stringify(execReq.function)), error: "", traceback: [] };
//     callback(null, reply);
// }
//
//
// export function runServer() {
//     const server = new grpc.Server();
//     server.addService(RuntimeService, { exports: listExports, start: start, execute: execute, activityReply: activityReply });
//
//     const port = '127.0.0.1:9999';
//     server.bindAsync(port, grpc.ServerCredentials.createInsecure(), (err, bindPort) => {
//         if (err) {
//             console.error(err);
//             return;
//         }
//         console.log(`Server running at ${port}`);
//     });
// }
