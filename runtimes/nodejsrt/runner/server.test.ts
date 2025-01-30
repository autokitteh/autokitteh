import { listExports, execute } from './server';
import { ServerUnaryCall, sendUnaryData } from '@grpc/grpc-js';
import { ExportsRequest, ExportsResponse, ExecuteRequest, ExecuteResponse } from './pb/ak';
import {functionsCache, resultsCache} from "./ak_call";
import {EventEmitter, once} from "node:events";

test('listExports', async () => {
    const mockRequest: Partial<ServerUnaryCall<ExportsRequest, ExportsResponse>> = {
        request: { fileName: "dep.js" },
    };

    process.env["CODE_DIR"] = "test_data/list_symbols"

    const mockCallback: sendUnaryData<ExportsResponse> = async (error, response) => {
        expect(response?.exports).toEqual([{"args": [], "line": 1, "name": "test_func"}]);
    };

    // Invoke the service method
    await listExports(mockRequest as ServerUnaryCall<ExportsRequest, ExportsResponse>, mockCallback);
});


test('execute', async () => {
    const encoder = new TextEncoder();
    const mockRequest: Partial<ServerUnaryCall<ExecuteRequest, ExecuteResponse>> = {
        request: {data: encoder.encode(JSON.stringify({"function": "sum", "args": [1,2]}))}
    };

    functionsCache["sum"] = async (a: number, b: number) => {
        return a + b;
    }

    resultsCache["sum"] = new EventEmitter()

    const mockCallback: sendUnaryData<ExecuteResponse> = async (error, response) => {
        const decoder = new TextDecoder();
        const resultsCacheKey = JSON.parse(decoder.decode(response?.result))
        expect(resultsCacheKey).toEqual("sum");
        const results = (await once(resultsCache["sum"], "result"))[0];
        expect(results).toEqual(3)
    };

    // Invoke the service method
    await execute(mockRequest as ServerUnaryCall<ExecuteRequest, ExecuteResponse>, mockCallback);
});
