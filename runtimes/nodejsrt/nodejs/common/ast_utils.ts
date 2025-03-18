import traverse, {NodePath} from "@babel/traverse";
import {parse} from "@babel/parser";
import generate from "@babel/generator";
import {isMemberExpression, identifier, isIdentifier, isAwaitExpression, isVariableDeclarator, stringLiteral} from "@babel/types";

import {listFiles} from "./file_utils";
import fs from "fs";
import {Export} from "../runtime/pb/autokitteh/user_code/v1/runner_svc_pb";

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

// Helper function to check if an identifier is from a relative import
function isFromRelativeImport(path: NodePath, identifierName: string): boolean {
    const binding = path.scope.getBinding(identifierName);
    if (binding?.path.parent?.type === 'ImportDeclaration') {
        const importSource = binding.path.parent.source.value;
        return importSource.startsWith('.');
    }
    return false;
}

// List of standard JavaScript built-in objects that don't need wrapping
const standardBuiltIns = new Set([
    'console', 'Promise', 'JSON'
    // Removed objects that don't have async methods: Math, Object, Array, String,
    // Number, Date, RegExp, Map, Set
]);

// Check if an identifier refers to a standard JavaScript built-in
function isStandardBuiltIn(name: string): boolean {
    return standardBuiltIns.has(name);
}

export async function patchCode(code: string): Promise<string> {
    const ast = parse(code, {sourceType: "module", plugins: ["typescript"]});

    traverse(ast, {
        CallExpression(path) {
            if (!isAwaitExpression(path.parent)) {
                return;
            }

            // For direct function calls
            if (isIdentifier(path.node.callee)) {
                const identifierName = path.node.callee.name;

                // Skip wrapping if it's a relative import or a standard built-in
                if (isFromRelativeImport(path, identifierName) || isStandardBuiltIn(identifierName)) {
                    return;
                }

                // Wrap the direct function call
                path.node.callee = identifier("(global as any).ak_call");
                path.node.arguments.unshift(identifier(identifierName));
            }
            // For method calls (obj.method())
            else if (isMemberExpression(path.node.callee)) {
                const object = path.node.callee.object;
                const method = isIdentifier(path.node.callee.property) ? path.node.callee.property.name : '';

                // Skip wrapping if the object is from a relative import or is a standard built-in
                if (isIdentifier(object)) {
                    if (isFromRelativeImport(path, object.name) || isStandardBuiltIn(object.name)) {
                        return;
                    }
                }

                // Only wrap if we have a valid method name
                if (method) {
                    path.node.callee = identifier("(global as any).ak_call");
                    path.node.arguments = [object, stringLiteral(method), ...path.node.arguments];
                }
            }
        }
    });

    return generate(ast).code;
}
