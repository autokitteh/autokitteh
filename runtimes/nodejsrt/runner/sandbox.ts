import * as vm from 'vm';
import * as fs from 'fs';
import {transformAsync} from "@babel/core";
import {EventEmitter, once} from "node:events";
import {patchCode} from "./ast_utils";
import {listFiles} from "./file_utils";
import {createGrpcTransport} from "@connectrpc/connect-node";
import {createClient, Client} from "@connectrpc/connect";
import {HandlerService} from "./pb/autokitteh/user_code/v1/handler_svc_pb";
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

interface FunctionsCache {
    [key: string]: Function
}

interface ResultsCache {
    [key: string]: EventEmitter;
}

export class Sandbox {
    context: Context
    emitter: EventEmitter
    runnerId: string
    workerAddress: string
    codeDir: string
    client: Client<typeof HandlerService>
    functionsCache: FunctionsCache
    resultsCache: ResultsCache

    constructor(runnerId: string, workerAddress: string, codeDir: string) {
        this.runnerId = runnerId;
        this.workerAddress = workerAddress;
        this.codeDir = codeDir;
        this.context = {};
        this.emitter = new EventEmitter();
        this.initContext()
        this.client = this.initClient()
        this.functionsCache = {}
        this.resultsCache = {}
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
        this.context.functionsCache = this.functionsCache
        this.context.resultsCache = this.resultsCache

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

    async run(f: string, args: any, callDone: boolean = false): Promise<any> {
        let code: string;

        if (callDone) {
            this.runnerId = args.runnerId;
            code = `(async () => {return await ${f}(...[${JSON.stringify(args)}]);})();`
            const p = vm.runInContext(code, this.context);
            const r = await p;
            this.client.done({runnerId: this.runnerId, result: {$typeName:"autokitteh.values.v1.Value", string: {$typeName:"autokitteh.values.v1.String", v:"yay"}}});
            return r
        } else {
            code = `(async () => {
                const results = await functionsCache[${f}](...[${JSON.stringify(args)}]);
                resultsCache[${f}].emit('return', results);
                })();`
            try {
                const p = vm.runInContext(code, this.context)
                const a = await p;
                console.log(a)
            } catch (error) {
                console.log(error)
            }

            // const results = await once(this.context.resultsCache[f], 'return');
            // console.log(results)
            // return results;
        }
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

        const uuid = "f_" + randomUUID().toString().replace(/-/g, "");
        this.functionsCache[uuid] = f
        this.resultsCache[uuid] = new EventEmitter()
        console.log("adding function:", f.name, uuid)

        let data = {
            f: uuid,
            f_args: []
        }

        if (f_args) {
            data.f_args = f_args
        }

        const serializedData = JSON.stringify(data)

        const encoder = new TextEncoder()
        const resp = await this.client.activity({runnerId: this.runnerId, data: encoder.encode(serializedData), callInfo: {
            function: uuid,
            args: [],
        }});
        console.log("activity call resp", resp, "args", args);
        // const results = await once(this.resultsCache[uuid], 'return');
        // console.log("got results", results)
        // return results;
        return "yay";
    }
}
