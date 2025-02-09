import * as vm from 'vm';
import * as fs from 'fs';
import {transformAsync} from "@babel/core";
import {EventEmitter, once} from "node:events";
import {patchCode} from "./ast_utils";
import {listFiles} from "./file_utils";
import path from "path";
import {createConnectTransport, createGrpcTransport} from "@connectrpc/connect-node";
import {createClient, Client} from "@connectrpc/connect";
import {HandlerService} from "./pb/autokitteh/user_code/v1/handler_svc_pb";
import {functionsCache, resultsCache} from "./ak_call";
import type {DescService} from "@bufbuild/protobuf";
import {randomUUID} from "node:crypto";

async function transpile(code: string, filename: string): Promise<string> {
    const result = await transformAsync(code, {
        presets: ['@babel/preset-env', '@babel/preset-typescript'],
        filename: filename,
        minified: false,
        compact: false,
    });

    if (!result?.code) {
        throw new Error('Failed to transpile TypeScript code.');
    }

    return result.code;
}

async function patchDir(dir: string, outDir: string): Promise<void> {
    if (fs.existsSync(outDir)) {
        fs.rmSync(outDir, {recursive: true});
    }

    fs.mkdirSync(outDir);

    const files = await listFiles(dir)
    for (const file of files) {
        if (!file.endsWith(".js") && !file.endsWith(".ts")) {
            continue
        }
        let code = fs.readFileSync(file, "utf8");
        const bob = file
        code = await patchCode(code)
        if (file.endsWith(".ts")) {
            code = await transpile(code, file);
        }
        let newFile = bob.replace(dir, outDir).replace(".ts", ".js")
        fs.appendFileSync(newFile, code, 'utf-8');
    }
}


export interface Context {
    [key: string]: any; // Allow dynamic properties
}

export interface Hook {
    module: string;
    function: string;
}

const defaultHook = (f: Function, args: any): any => {
    return true;
}

interface OriginalFunctions {
    [key: string]: Function
}

export class Sandbox {
    context: Context
    hookFunc: Function
    emitter: EventEmitter
    runnerId: string
    workerAddress: string
    codeDir: string
    client: Client<typeof HandlerService>

    constructor(hookFunc: Function, runnerId: string, workerAddress: string, codeDir: string) {
        this.runnerId = runnerId;
        this.workerAddress = workerAddress;
        this.codeDir = codeDir;
        this.context = {};
        this.hookFunc = hookFunc;
        this.emitter = new EventEmitter();
        this.initContext()
        this.client = this.initClient()
    }

    initContext() {
        this.context.exports = {}
        this.context.results = ""
        this.context.emmiter = this.emitter
        this.context.ak_call = this.ak_call
        this.context.originalFunction = {}
        this.context.workingDir = "."
        this.context.console = {
            log: console.log,
        }

        this.context.require = (moduleName: string) => {

            if (moduleName.startsWith(".")) {
                moduleName =  this.codeDir + "/dist/" + moduleName + ".js"
                const code = fs.readFileSync(moduleName, "utf8")
                vm.runInContext(code, this.context)
                return this.context
            }

            let mod =  require(moduleName)
            for (let k in mod) {
                if (typeof mod[k] === "function") {
                    mod[k].ak_call = true
                }

                if (typeof mod[k] === "object") {
                    for (let m in mod[k]) {
                        if (typeof mod[k][m] === "function") {
                            mod[k][m].ak_call = true
                        }
                    }
                }
            }

            return mod
        }
        vm.createContext(this.context);
    }

    async loadFile(filePath: string): Promise<void> {
        let parts = filePath.split("/");
        parts.pop();
        let dir = parts.join("/")
        let out = dir + "/dist"
        await patchDir(dir, out)

        filePath = filePath.replace(dir, out).replace(".ts", ".js");
        let code = fs.readFileSync(filePath, 'utf8');
        vm.runInContext(code, this.context);
    }

    async run(code: string): Promise<void> {
        vm.runInContext(code, this.context);
        // await this.client.done({runnerId: this.runnerId, result: {$typeName:"autokitteh.values.v1.Value", string: {$typeName:"autokitteh.values.v1.String", v:"yay"}}});
    }

    initClient() {
        const transport = createGrpcTransport({
            baseUrl: `http://${this.workerAddress}`,
        });
        return createClient(HandlerService, transport);
    }

    ak_call = async (...args: any) => {
        let f = args[0];
        let f_args = args.slice(1)

        if (f.ak_call !== true) {
            return await f(...f_args);
        }

        const uuid = randomUUID()
        functionsCache[uuid] = f

        let data = {
            f: f.name,
            f_args: ""
        }

        if (f_args) {
            data.f_args = f_args
        }

        const serializedData = JSON.stringify(data)

        const encoder = new TextEncoder()
        await this.client.activity({runnerId: this.runnerId, data: encoder.encode(serializedData)});
        return "aa"
    }
}
