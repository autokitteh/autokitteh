import traverse, {NodePath} from "@babel/traverse";
import {parse} from "@babel/parser";
import generate from "@babel/generator";
import {
    isMemberExpression,
    identifier,
    isIdentifier,
    isAwaitExpression,
    isVariableDeclarator,
    CallExpression
} from "@babel/types";

import {listFiles} from "./file_utils";
import * as fs from "fs";
import {Export} from "../pb/autokitteh/user_code/v1/runner_svc_pb";

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

// Helper function to check if an identifier is from a relative import or local
function isFromRelativeImport(path: NodePath, identifierName: string): boolean {
    const binding = path.scope.getBinding(identifierName);

    // Check for import from relative path
    if (binding?.path.parent?.type === 'ImportDeclaration') {
        const importSource = binding.path.parent.source.value;
        return importSource.startsWith('.');
    }

    // Special case: For variables initialized with a constructor
    if (binding?.path.node.type === 'VariableDeclarator' &&
        binding.path.node.init?.type === 'NewExpression' &&
        binding.path.node.init.callee.type === 'Identifier') {

        // Get the class name
        const className = binding.path.node.init.callee.name;

        // Look up where the class comes from
        const classBinding = path.scope.getBinding(className);

        // If the class is imported, check if it's from a relative path
        if (classBinding?.path.parent?.type === 'ImportDeclaration') {
            const importSource = classBinding.path.parent.source.value;
            return importSource.startsWith('.');
        }
    }

    // If binding exists but wasn't imported, it's defined in this file
    if (binding && !binding.path.findParent(p => p.isImportDeclaration())) {
        return true;
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

// Helper function to check if a call is to a direct autokitteh method
function isAutokittehMethod(path: NodePath<CallExpression>): { isAutokitteh: boolean; methodName: string | null } {
    const callee = path.node.callee;

    // Must be a member expression (e.g., autokitteh.something)
    if (!isMemberExpression(callee)) {
        return { isAutokitteh: false, methodName: null };
    }

    // Object must be the identifier 'autokitteh'
    if (!isIdentifier(callee.object) || callee.object.name !== 'autokitteh') {
        return { isAutokitteh: false, methodName: null };
    }

    // Property must be a direct identifier (not another member expression)
    if (!isIdentifier(callee.property)) {
        return { isAutokitteh: false, methodName: null };
    }

    // Only patch known event methods
    const methodName = callee.property.name;
    const knownMethods = new Set(['subscribe', 'unsubscribe', 'nextEvent']);
    if (!knownMethods.has(methodName)) {
        return { isAutokitteh: false, methodName: null };
    }

    return {
        isAutokitteh: true,
        methodName
    };
}

export async function patchCode(code: string): Promise<string> {
    const ast = parse(code, {sourceType: "module", plugins: ["typescript"]});

    traverse(ast, {
        // ImportDeclaration(path) {
        //     const source = path.node.source.value;
        //     if ( source === 'autokitteh' || source.startsWith('autokitteh/')) {
        //         // Get the original import code for the comment
        //         const importCode = generate(path.node).code;
        //         // Add a comment indicating this was removed by the build process
        //         const comment = ` ${importCode} - commented out by autokitteh build process`;
        //
        //         // Instead of replacing with a dummy import, we're going to completely remove it
        //         // Add the comment as a standalone comment node
        //         const parentPath = path.parentPath;
        //
        //         // Remove the import node completely
        //         path.remove();
        //
        //         // Add a standalone comment where the import was
        //         if (parentPath && parentPath.node) {
        //             parentPath.addComment('leading', comment);
        //         }
        //     }
        // },
        CallExpression(path) {
            // Check for autokitteh method calls first
            const { isAutokitteh, methodName } = isAutokittehMethod(path);
            if (isAutokitteh && methodName) {
                // Transform the call to use global syscalls with same method name
                path.node.callee = identifier(`(global as any).syscalls.${methodName}`);
                return;
            }

            // Handle other async calls as before
            if (!isAwaitExpression(path.parent)) {
                return;
            }

            // For direct function calls
            if (isIdentifier(path.node.callee)) {
                const identifierName = path.node.callee.name;
                const binding = path.scope.getBinding(identifierName);

                // Check if it's a variable that references an external method
                if (binding?.path.node.type === 'VariableDeclarator' &&
                    isMemberExpression(binding.path.node.init)) {

                    const objExpr = binding.path.node.init.object;
                    if (isIdentifier(objExpr)) {
                        const objectName = objExpr.name;

                        // If the object is from an external module, wrap the function call
                        if (!isFromRelativeImport(binding.path, objectName) && !isStandardBuiltIn(objectName)) {
                            path.node.callee = identifier("(global as any).ak_call");
                            path.node.arguments.unshift(identifier(identifierName));
                            return;
                        }
                    }
                }

                // Check if it's a direct external function
                if (isFromRelativeImport(path, identifierName) || isStandardBuiltIn(identifierName)) {
                    return;
                }

                // Wrap the direct function call
                path.node.callee = identifier("(global as any).ak_call");
                path.node.arguments.unshift(identifier(identifierName));
            }
            // For method calls (obj.method())
            else if (isMemberExpression(path.node.callee)) {
                // Get the root object of the member expression
                let rootObject = path.node.callee.object;

                // Traverse to the root object in case of nested member expressions
                while (isMemberExpression(rootObject)) {
                    rootObject = rootObject.object;
                }

                // Only proceed if the root is an identifier
                if (isIdentifier(rootObject)) {
                    const rootName = rootObject.name;

                    // Skip wrapping if the root object is from a relative import, local, or a standard built-in
                    if (isFromRelativeImport(path, rootName) || isStandardBuiltIn(rootName)) {
                        return;
                    }

                    // Wrap the entire method expression rather than passing object and method separately
                    const originalCallee = path.node.callee;
                    path.node.callee = identifier("(global as any).ak_call");
                    path.node.arguments.unshift(originalCallee);
                }
            }
        }
    });

    return generate(ast).code;
}
