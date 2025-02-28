import traverse, {NodePath} from "@babel/traverse";
import {parse} from "@babel/parser";
import generate from "@babel/generator";
import { Expression, isMemberExpression, identifier, isIdentifier, isAwaitExpression, isVariableDeclarator, stringLiteral, MemberExpression, Identifier, CallExpression } from "@babel/types";

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
        ArrowFunctionExpression: function (path) {
            let params: string[] = []
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

            exports.push({args: params, line: line, name: name, file: file, $typeName:"autokitteh.user_code.v1.Export"});
        },
    })

    return exports;
}

export async function listExportsInDirectory(dirPath: string): Promise<Export[]> {
    const files = await listFiles(dirPath);
    let exports: Export[] = []
    for (const file of files) {
        if (!file.endsWith(".js") && !file.endsWith(".ts")) {
            continue;
        }

        let code = fs.readFileSync(file, "utf8");
        exports = exports.concat(exports, await listExports(code, file))
    }

    return exports;
}

export async function patchCode(code: string, exclude: string[] = []): Promise<string> {
    const parserOptions: ParserOptions = { sourceType: "module", plugins: ["typescript", "asyncGenerators", "topLevelAwait"] };
    const ast = parse(code, parserOptions);

    traverse(ast, {
        CallExpression: function (path: NodePath<CallExpression>) {
            let originalFunc: string = "";

            if (!isAwaitExpression(path.parent)) {
                return;
            }

            function getFullMemberExpressionParts(memberExpr: Expression): { object: Identifier; method: string } | null {
                let parts: string[] = [];
                while (isMemberExpression(memberExpr)) {
                    if (isIdentifier(memberExpr.property)) {
                        parts.unshift(memberExpr.property.name);
                    } else {
                        return null;
                    }
                    memberExpr = memberExpr.object as MemberExpression;
                }
                if (isIdentifier(memberExpr)) {
                    parts.unshift(memberExpr.name);
                    const method = parts.pop()!
                    const obj = parts.join(".")
                    return { object: identifier(obj), method: method };
                }
                return null;
            }

            if (isMemberExpression(path.node.callee)) {
                const parts = getFullMemberExpressionParts(path.node.callee);
                if (parts) {
                    originalFunc = parts.object.name + "." + parts.method;
                    path.node.callee = identifier("ak_call");
                    path.node.arguments = [parts.object, stringLiteral(parts.method), ...path.node.arguments];
                }
            } else if (isIdentifier(path.node.callee)) {
                originalFunc = path.node.callee.name;
                path.node.callee.name = "ak_call";
                path.node.arguments.unshift(identifier(originalFunc));
            }

            if (exclude.includes(originalFunc) || originalFunc === "") {
                return;
            }
        },
    });

    return generate(ast).code;
}
