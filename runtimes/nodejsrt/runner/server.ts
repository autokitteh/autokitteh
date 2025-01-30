import * as grpc from '@grpc/grpc-js';
import {
    ExportsRequest,
    ExportsResponse,
    RuntimeService,
    RuntimeServer,
    StartRequest,
    StartResponse,
    ExecuteResponse, ExecuteRequest,
    ActivityReplyRequest, ActivityReplyResponse,
} from './pb/ak';

import {ak_call, functionsCache, resultsCache} from "./ak_call";
import {Sandbox} from "./sandbox";
import fs from "fs";
import {listSymbols, Symbol} from "./ast_utils";


export const activityReply: RuntimeServer['activityReply'] = (call: { request: ActivityReplyRequest; }, callback: (arg0: null, arg1: ActivityReplyResponse) => void) => {
    callback(null, { error: ""});
}

export const start: RuntimeServer['start'] = async (call: { request: StartRequest; }, callback: (arg0: null, arg1: StartResponse) => void) => {
    const decoder = new TextDecoder();
    const request: StartRequest = call.request;
    const args = decoder.decode(request.event?.data)
    const [fileName, functionName] = request.entryPoint.split(":")

    const sandbox = new Sandbox(ak_call)
    const codeDir = process.env["CODE_DIR"];
    await sandbox.loadFile(`${codeDir}/${fileName}`)
    const code = `${functionName}(...${args})`

    sandbox.run(code) // TODO: handle errors
    callback(null, { traceback: [], error: ""});
}

export const listExports: RuntimeServer['exports'] = async (call: { request: ExportsRequest; }, callback: (arg0: null, arg1: ExportsResponse) => void) => {
    const request: ExportsRequest = call.request;
    const codeDir = process.env["CODE_DIR"];
    const filePath = `${codeDir}/${request.fileName}`
    let symbols: Symbol[] = []

    try {
        const code = await fs.promises.readFile(filePath, "utf-8")
        symbols = await listSymbols(code)
    }
    catch (error) {
        console.log(error)
    }
    const reply: ExportsResponse = { exports: symbols, error: "" };
    callback(null, reply);
};

export const execute: RuntimeServer['execute'] = async (call: { request: ExecuteRequest; }, callback: (arg0: null, arg1: ExecuteResponse) => void) => {
    const request: ExecuteRequest = call.request;
    const decoder = new TextDecoder(); // Default is "utf-8"
    const encoder = new TextEncoder();
    const data = decoder.decode(request.data)
    const execReq = JSON.parse(data)
    let results = await functionsCache[execReq.function](...execReq.args)
    console.log("execute:", execReq)
    resultsCache[execReq.function].emit("result", results)
    const reply: ExecuteResponse = { result: encoder.encode(JSON.stringify(execReq.function)), error: "", traceback: [] };
    callback(null, reply);
}


export function runServer() {
    const server = new grpc.Server();
    server.addService(RuntimeService, { exports: listExports, start: start, execute: execute, activityReply: activityReply });

    const port = '127.0.0.1:9999';
    server.bindAsync(port, grpc.ServerCredentials.createInsecure(), (err, bindPort) => {
        if (err) {
            console.error(err);
            return;
        }
        console.log(`Server running at ${port}`);
    });
}
