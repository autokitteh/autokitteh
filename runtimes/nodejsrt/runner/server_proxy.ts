import {
    ActivityReplyRequest,
    ActivityReplyResponse, ExecuteRequest, ExecuteResponse,
    ExportsRequest, ExportsResponse,
    RunnerService, StartRequest, StartResponse
} from "./pb/autokitteh/user_code/v1/runner_svc_pb";


import {ConnectRouter, createClient} from "@connectrpc/connect";
import {HealthRequest, HealthResponse} from "./pb/autokitteh/runner_manager/v1/runner_manager_svc_pb";
import {createGrpcTransport} from "@connectrpc/connect-node";


export const createService = (codeDir: string, runnerId: string, workerAddress: string) => {
    console.log("--code-dir=", codeDir, "--runner-id=", runnerId, "--worker-address=", workerAddress);
    const transport = createGrpcTransport({
        baseUrl: `http://localhost:1111`,
    });

    const client = createClient(RunnerService, transport);

  return (router: ConnectRouter) => router.service(RunnerService, {
      async activityReply(req: ActivityReplyRequest) : Promise<ActivityReplyResponse> {
          return await client.activityReply(req)
      },

      async exports(req: ExportsRequest): Promise<ExportsResponse> {
          return await client.exports(req)
      },

      async execute(req: ExecuteRequest): Promise<ExecuteResponse> {
          return await client.execute(req)
      },

      async health(req: HealthRequest): Promise<HealthResponse> {
          return client.health(req)
      },

      async start(req: StartRequest): Promise<StartResponse> {
          const decoder = new TextDecoder();
          const encoder = new TextEncoder();
          const args = JSON.parse(decoder.decode(req.event?.data))
          args.runnerId = runnerId
          args.codeDir = codeDir

          if (req.event) {
              req.event.data = encoder.encode(JSON.stringify(args))
          }

          return await client.start(req)
      }
  });
}
