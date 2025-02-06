import * as vm from 'vm';
import * as fs from 'fs';
import {transformAsync} from "@babel/core";
import {EventEmitter, once} from "node:events";
import {patchCode} from "./ast_utils";
import {listFiles} from "./file_utils";
import path from "path";
import {createGrpcTransport} from "@connectrpc/connect-node";
import {createClient} from "@connectrpc/connect";
import {HandlerService} from "./pb/autokitteh/user_code/v1/handler_svc_pb";
import {resultsCache} from "./ak_call";

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
        if (!file.endsWith(".js") || file.endsWith(".ts")) {
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

const done = async () => {
    const transport = createGrpcTransport({
        baseUrl: "http://localhost:9980",
    });
    const client = createClient(HandlerService, transport);
    await client.done({runnerId: "runner_01jkd8ryv5eq3bqnq7m45c8edr", error: "", traceback: []});
}

export class Sandbox {
    context: Context
    hookFunc: Function
    emitter: EventEmitter

    constructor(hookFunc: Function) {
        this.context = {};
        this.hookFunc = hookFunc;
        this.emitter = new EventEmitter();
        this.initContext()
    }

    initContext() {
        this.context.exports = {}
        this.context.results = ""
        this.context.emmiter = this.emitter
        this.context.ak_call = this.hookFunc
        this.context.originalFunction = {}
        this.context.workingDir = "."
        this.context.console = {
            log: console.log,
        }

        this.context.require = (moduleName: string) => {

            if (moduleName.startsWith(".")) {
                moduleName =  "./dist/" + moduleName + ".js"
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
        await done()
    }

    async runOriginalFunction(name: string, args: any[]): Promise<any> {
        const code = `originalFunction['${name}'](...${JSON.stringify(args)})`;
        const out = await vm.runInContext(code, this.context);
        this.emitter.emit("return", out);
        return out
    }
}
