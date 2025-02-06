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
        const decoder = new TextDecoder();
        const args = [decoder.decode(req.event?.data)]
        const [fileName, functionName] = req.entryPoint.split(":")

        const sandbox = new Sandbox(ak_call)
        const codeDir = process.env["CODE_DIR"];
        await sandbox.loadFile(`${codeDir}/${fileName}`)
        const code = `${functionName}()`

        sandbox.run(code)
        return {$typeName: "autokitteh.user_code.v1.StartResponse", error:"", traceback: []}
    }
});
