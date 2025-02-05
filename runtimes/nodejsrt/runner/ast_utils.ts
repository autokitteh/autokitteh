import traverse from "@babel/traverse";
import {parse} from "@babel/parser";
import generate from "@babel/generator";
import {isMemberExpression, identifier, isIdentifier, isAwaitExpression} from "@babel/types";
import {listFiles} from "./file_utils";
import fs from "fs"
import {Export} from "./pb/autokitteh/user_code/v1/runner_svc_pb";

export interface Symbol {
    file: string;
    name: string;
    args: string[];
    line: number;
}

export interface Symbols {
    [fileName: string]: Symbol[];
}

export async function listExports(code: string, file: string): Promise<Export[]> {
    const ast = parse(code, {sourceType: "module", plugins: ["typescript"]});

    let exports: Export[] = [];
    traverse(ast, {
        FunctionDeclaration: function (path) {
            let params: string[] = []
            path.node.params.forEach((param) => {
                if (isIdentifier(param)) {
                    params.push(param.name)
                }
            })

            let name  = ""
            if (isIdentifier(path.node.id)) {
                name = path.node.id.name;
            }


            let line = 0;

            if (path.node.loc) {
                line = path.node.loc.start.line;
            }

            exports.push({args: params, line: line, name: name, file: file, $typeName:"autokitteh.user_code.v1.Export"});
        },
    })

    return exports;
}

export async function listExportsInDirectory(dirPath: string): Promise<Export[]> {
    const files = await listFiles(dirPath);
    let exports: Export[] = []
    for (const file of files) {
        let code = fs.readFileSync(file, "utf8");
        exports = exports.concat(exports, await listExports(code, file))
    }

    return exports;
}

export async function patchCode(code: string, exclude: string[] = []): Promise<string> {
    const ast = parse(code, {sourceType: "module", plugins: ["typescript", "asyncGenerators", "topLevelAwait"]});

    traverse(ast, {
        CallExpression: function (path) {
            let originalFunc = "";

            /*
            TODO: support member expression with N levels
            ex: a.b.c.d("param")
            ATM we only support single level
            ex: a.b("param")
            */

            if (!isAwaitExpression(path.parent)) {
                return;
            }

            if (isMemberExpression(path.node.callee)) {
                if (isIdentifier(path.node.callee.object) && isIdentifier(path.node.callee.property)) {
                    originalFunc = path.node.callee.object.name + "." + path.node.callee.property.name;
                    path.node.callee = identifier("ak_call");
                }
            }
            else if (isIdentifier(path.node.callee)) {
                originalFunc = path.node.callee.name;
                path.node.callee.name =  "ak_call";
            }

            if (exclude.includes(originalFunc)) {
                return;
            }

            if (originalFunc == "") {
                return;
            }

            path.node.arguments.unshift(identifier(originalFunc));
        },
    })

    return generate(ast).code;
}
