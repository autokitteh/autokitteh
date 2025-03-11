import traverse, {NodePath} from "@babel/traverse";
import {parse} from "@babel/parser";
import generate from "@babel/generator";
import { isMemberExpression, identifier, isIdentifier, isAwaitExpression, isVariableDeclarator, stringLiteral, CallExpression } from "@babel/types";

import {listFiles} from "./file_utils";
import fs from "fs"
import {Export} from "./pb/autokitteh/user_code/v1/runner_svc_pb";
import {ParserOptions} from "@babel/core";

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

    const exports: Export[] = [];
    traverse(ast, {
        FunctionDeclaration: function (path) {
            const params: string[] = []
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

            if (name === "") {
                return;
            }

            exports.push({args: params, line: line, name: name, file: file, $typeName:"autokitteh.user_code.v1.Export"});
        },
        ArrowFunctionExpression: function (path) {
            const params: string[] = []
            path.node.params.forEach((param) => {
                if (isIdentifier(param)) {
                    params.push(param.name)
                }
            })

            let name  = ""
            if (isVariableDeclarator(path.parent)) {
                if (isIdentifier(path.parent.id)) {
                    name = path.parent.id.name;
                }
            }

            let line = 0;
            if (path.node.loc) {
                line = path.node.loc.start.line;
            }

            if (name === "") {
                return
            }

            exports.push({args: params, line: line, name: name, file: file, $typeName:"autokitteh.user_code.v1.Export"});
        },
    })

    return exports;
}

export async function listExportsInDirectory(dirPath: string): Promise<Export[]> {
    const files = await listFiles(dirPath);
    const exports: Export[] = []
    for (const file of files) {
        if (file.includes("node_modules")) {
            continue
        }

        if (!file.endsWith(".ts")) {
            continue;
        }

        const code = fs.readFileSync(file, "utf8");
        const new_exports = await listExports(code, file)
        exports.push(...new_exports)
    }

    return exports;
}

export async function patchCode(code: string): Promise<string> {
    const ast = parse(code, {sourceType: "module", plugins: ["typescript"]});

    traverse(ast, {
        CallExpression(path) {
            if (!isAwaitExpression(path.parent)) {
                return;
            }

            // Skip wrapping if it's a relative import
            if (isIdentifier(path.node.callee)) {
                const binding = path.scope.getBinding(path.node.callee.name);
                if (binding?.path.parent?.type === 'ImportDeclaration') {
                    const importSource = binding.path.parent.source.value;
                    if (importSource.startsWith('.')) {
                        return;
                    }
                }
            }

            if (isMemberExpression(path.node.callee)) {
                const parts = {
                    object: path.node.callee.object,
                    method: isIdentifier(path.node.callee.property) ? path.node.callee.property.name : ''
                };
                if (parts.method) {
                    path.node.callee = identifier("ak_call");
                    path.node.arguments = [parts.object, stringLiteral(parts.method), ...path.node.arguments];
                }
            } else if (isIdentifier(path.node.callee)) {
                const name = path.node.callee.name;
                path.node.callee = identifier("ak_call");
                path.node.arguments.unshift(identifier(name));
            }
        }
    });

    return generate(ast).code;
}
