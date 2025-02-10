import * as vm from 'vm';
import * as fs from 'fs';
import {transformAsync} from "@babel/core";
import {patchCode} from "./ast_utils";
import {listFiles} from "./file_utils";

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

    initContext() {
        this.context.exports = {}
        this.context.ak_call = this.ak_call
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

    async run(f: string, args: any): Promise<any> {
        let code = `(async () => {
            return await ${f}(...[${JSON.stringify(args)}]);
        })();`
        return await vm.runInContext(code, this.context)
    }
}
