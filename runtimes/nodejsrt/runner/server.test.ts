import { createRouterTransport, createClient } from "@connectrpc/connect";
import { createService } from "./server";

import {RunnerService } from "./pb/autokitteh/user_code/v1/runner_svc_pb";
import {Sandbox} from "./sandbox";

const mockSandbox = new Sandbox("", async (f: Function, args: any) => { return await f(...args)})

test('activity reply', async () => {
    const mockTransport = createRouterTransport(createService("","",mockSandbox))
    const client = createClient(RunnerService, mockTransport);
    const encoder = new TextEncoder()
    await client.activityReply({
        result: {
            $typeName:"autokitteh.values.v1.Value",
            custom: {
                $typeName: "autokitteh.values.v1.Custom",
                executorId: "runnerId",
                data: encoder.encode(JSON.stringify({test: "test"})),
                value: {
                    $typeName:"autokitteh.values.v1.Value",
                    string: {
                        $typeName:"autokitteh.values.v1.String",
                        v: "yay",
                    }
                }
            }
        }
    });
});

test('listExports', async () => {
    const mockTransport = createRouterTransport(createService("test_data/list_symbols","", mockSandbox))
    const client = createClient(RunnerService, mockTransport);
    const resp = await client.exports({fileName: "dep.js"})
    expect(resp.exports).toEqual([{
        "$typeName": "autokitteh.user_code.v1.Export",
        "args": [],
        "line": 1,
        "name": "test_func",
        "file": "test_data/list_symbols/dep.js",
    }]);
});


test('execute', async () => {
    const encoder = new TextEncoder();
    const decoder = new TextDecoder();
    const mockTransport = createRouterTransport(createService("","", mockSandbox))
    const client = createClient(RunnerService, mockTransport);
    const req = JSON.stringify({"function": "sum", "args": [1,2]})
    const resp = await client.execute({data: encoder.encode(req)})
    const results = JSON.parse(decoder.decode(resp.result?.custom?.data))
    expect(results).toEqual("yay")
});
