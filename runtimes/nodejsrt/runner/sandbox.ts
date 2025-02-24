import * as vm from 'vm';
import * as fs from 'fs';
import {transformAsync} from "@babel/core";
import {patchCode} from "./ast_utils";
import {listFiles} from "./file_utils";
import {execSync} from "node:child_process";

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

async function patchDir(dir: string): Promise<void> {
    const files = await listFiles(dir)
    for (const file of files) {
        if (file.includes("/node_modules/")) {
            continue
        }

        if (!file.endsWith(".js") && !file.endsWith(".ts")) {
            continue
        }
        let code = fs.readFileSync(file, "utf8");
        try {
            code = await patchCode(code)
            if (file.endsWith(".ts")) {
                code = await transpile(code, file);
            }
            let newFile = file.replace(".ts", ".js")
            fs.appendFileSync(newFile, code, 'utf-8');
        } catch (err) {
            console.error(err);
        }
    }
}


export interface Context {
    [key: string]: any; // Allow dynamic properties
}

export class Sandbox {
    context: Context
    codeDir: string
    ak_call: Function

    constructor(codeDir: string, ak_call: Function) {
        this.codeDir = codeDir;
        this.context = {};
        this.ak_call = ak_call;
        this.initContext()
    }

    setCodeDir(codeDir: string): void {
        this.codeDir = codeDir;
        let output = execSync(`cd ${this.codeDir}; npm install`).toString();
        console.log(output);
    }

    initContext() {
        this.context.exports = {}
        this.context.ak_call = this.ak_call
        this.context.workingDir = "."
        this.context.console = {
            log: console.log,
        }
        this.context.process = {
            env: process.env,
            cwd: () => this.codeDir
        }

        this.context.require = (moduleName: string) => {

            if (moduleName.startsWith(".")) {
                moduleName =  this.codeDir + "/" + moduleName + ".js"
                const code = fs.readFileSync(moduleName, "utf8")
                vm.runInContext(code, this.context)
                return this.context
            }

            let mod: any
            console.log("resolving " + moduleName)
            try {
                mod =  require(moduleName)
            } catch (e) {
                mod =  require(`${this.codeDir}/node_modules/${moduleName}`)
            }

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
        await patchDir(dir)

        filePath = filePath.replace(".ts", ".js");
        let code = fs.readFileSync(filePath, 'utf8');
        vm.runInContext(code, this.context);
    }

    async run(f: string, args: any, callback: Function): Promise<any> {
        let code = `(async () => {
            return await ${f}(...[${JSON.stringify(args)}]);
        })();`
        const results = await vm.runInContext(code, this.context)
        callback(results)
    }
}
