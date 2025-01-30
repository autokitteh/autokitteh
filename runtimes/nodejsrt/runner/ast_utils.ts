import traverse from "@babel/traverse";
import {parse} from "@babel/parser";
import generate from "@babel/generator";
import {isMemberExpression, identifier, isIdentifier, isAwaitExpression} from "@babel/types";
import {listFiles} from "./file_utils";
import fs from "fs"

export interface Symbol {
    file: string;
    name: string;
    args: string[];
    line: number;
}

export interface Symbols {
    [fileName: string]: Symbol[];
}

export async function listSymbols(code: string, file: string): Promise<Symbol[]> {
    const ast = parse(code, {sourceType: "module", plugins: ["typescript"]});

    let symbols: Symbol[] = [];
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

            symbols.push({args: params, line: line, name: name, file: file});
        },
    })

    return symbols;
}

export async function listSymbolsInDirectory(dirPath: string): Promise<Symbol[]> {
    const files = await listFiles(dirPath);
    let symbols: Symbol[] = []
    for (const file of files) {
        let code = fs.readFileSync(file, "utf8");
        symbols = symbols.concat(symbols, await listSymbols(code, file))
    }

    return symbols;
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
