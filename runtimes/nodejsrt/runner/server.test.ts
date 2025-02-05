import {functionsCache} from "./ak_call";
import { createRouterTransport, createClient } from "@connectrpc/connect";
import server from "./server";

import {RunnerService } from "./pb/autokitteh/user_code/v1/runner_svc_pb";
import { Value } from "./pb/autokitteh/values/v1/values_pb";

test('xxx', async () => {
    const mockTransport = createRouterTransport(server)
    const client = createClient(RunnerService, mockTransport);
    let v: Value = { $typeName:"autokitteh.values.v1.Value", string: {$typeName:"autokitteh.values.v1.String", v:"a"}};
    await client.activityReply({result: v, error: "a", $typeName:"autokitteh.user_code.v1.ActivityReplyRequest"});
});

test('listExports', async () => {
    const mockTransport = createRouterTransport(server)
    const client = createClient(RunnerService, mockTransport);
    process.env["CODE_DIR"] = "test_data/list_symbols"

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
    functionsCache["sum"] = async (a: number, b: number) => {
        return a + b;
    }
    const encoder = new TextEncoder();
    const mockTransport = createRouterTransport(server)
    const client = createClient(RunnerService, mockTransport);
    const req = JSON.stringify({"function": "sum", "args": [1,2]})
    const resp = await client.execute({data: encoder.encode(req)})
    const results = JSON.parse(resp.result?.string?.v ?? '')
    expect(results).toEqual(3)
});
