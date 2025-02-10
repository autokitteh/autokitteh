import {
    ActivityReplyRequest,
    ActivityReplyResponse, ExecuteRequest, ExecuteResponse, Export,
    ExportsRequest, ExportsResponse,
    RunnerService, StartRequest, StartResponse
} from "./pb/autokitteh/user_code/v1/runner_svc_pb";

import {Sandbox} from "./sandbox";
import fs from "fs";
import {listExports} from "./ast_utils";

import type { ConnectRouter } from "@connectrpc/connect";
import {HealthRequest, HealthResponse} from "../../../proto/gen/ts/autokitteh/runner_manager/v1/runner_manager_svc_pb";

export const createService = (codeDir: string, runnerId: string, sandbox: Sandbox) => {
    const decoder = new TextDecoder();
    const encoder = new TextEncoder();

    return (router: ConnectRouter) => router.service(RunnerService, {
          async activityReply(req: ActivityReplyRequest) : Promise<ActivityReplyResponse> {
              const data = req.result?.custom?.data
              const parsedData = JSON.parse(decoder.decode(data))
              console.log("activity reply req", req, "parsed data", parsedData);
              return {error: "", $typeName: "autokitteh.user_code.v1.ActivityReplyResponse"}
          },

          async exports(req: ExportsRequest): Promise<ExportsResponse> {
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
              const data = decoder.decode(req.data)
              const execReq = JSON.parse(data)
              console.log("exec req:", execReq)
              // let results = JSON.stringify(await sandbox.run(execReq.f, execReq.f_args))
              let results = JSON.stringify("yay")
              return {
                  $typeName: "autokitteh.user_code.v1.ExecuteResponse",
                  error: "",
                  result: {
                      $typeName:"autokitteh.values.v1.Value",
                      custom: {
                          $typeName: "autokitteh.values.v1.Custom",
                          executorId: runnerId,
                          data: encoder.encode(results),
                          value: {
                              $typeName:"autokitteh.values.v1.Value",
                              string: {
                                  $typeName:"autokitteh.values.v1.String",
                                  v: "yay",
                              }
                          }

                      }
                  },
                  traceback: []
              }
          },

          async health(req: HealthRequest): Promise<HealthResponse> {
              console.log("health req", req)
              return {error: ""}
          },

          async start(req: StartRequest): Promise<StartResponse> {
              const args = decoder.decode(req.event?.data)
              const parsedArgs = JSON.parse(args)
              const [fileName, functionName] = req.entryPoint.split(":")
              await sandbox.loadFile(`${codeDir}/${fileName}`)
              sandbox.run(functionName, parsedArgs).then((results) => {
                  console.log(results);
              })
              return {$typeName: "autokitteh.user_code.v1.StartResponse", error:"", traceback: []}
          }
    });
}
