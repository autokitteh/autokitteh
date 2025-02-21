import {listExports, Symbol, listExportsInDirectory, Symbols, patchCode} from './ast_utils';
import {Export} from "./pb/autokitteh/user_code/v1/runner_svc_pb";

test('test list symbols', async () => {
    const code = `
    import {a} from './a';
    
    function some_func(a, b) {
        return a + b;
    }
    function another_func() {}
    `;

    const expected: Export[] = [
        {line: 4, name: "some_func", args: ["a", "b"], file: "test.ts", $typeName:"autokitteh.user_code.v1.Export"},
        {line: 7, name: "another_func", args: [], file: "test.ts", $typeName:"autokitteh.user_code.v1.Export"},
    ];
    expect(await listExports(code, "test.ts")).toEqual(expected);
});

test('test list symbols arrow func', async () => {
    const code = `
    import {a} from './a';
    
    const some_func = (a, b) => {
        return a + b;
    }
    const another_func = () => {}
    `;

    const expected: Export[] = [
        {line: 4, name: "some_func", args: ["a", "b"], file: "test.ts", $typeName:"autokitteh.user_code.v1.Export"},
        {line: 7, name: "another_func", args: [], file: "test.ts", $typeName:"autokitteh.user_code.v1.Export"},
    ];
    expect(await listExports(code, "test.ts")).toEqual(expected);
});

test('list symbols in directory', async () => {
    const dir = "./test_data/list_symbols";
    const actual = await listExportsInDirectory(dir)
    const expected:  Export[] = [
        {line: 1, name: "test_func", args: [], file: "test_data/list_symbols/dep.js", $typeName:"autokitteh.user_code.v1.Export"},
    ]

    expect(actual).toEqual(expected);
})


test('patch sync member call', async () => {
    const code = `
    import {client} from "slack"
    
    function some_func() {
        client.send("test");
    }
    `

    const expected = `import { client } from "slack";
function some_func() {
  client.send("test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('patch sync function call', async () => {
    const code = `
    import {f} from "lib"
    
    function some_func() {
        f("test");
    }
    `

    const expected = `import { f } from "lib";
function some_func() {
  f("test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('patch async function call', async () => {
    const code = `
    import {f} from "lib"
    
    async function some_func() {
        await f("test");
    }
    `

    const expected = `import { f } from "lib";
async function some_func() {
  await ak_call(f, "test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('patch async member call', async () => {
    const code = `
    import {a} from "lib"
    
    async function some_func() {
        await a.b("test");
    }
    `

    const expected = `import { a } from "lib";
async function some_func() {
  await ak_call(a.b, "test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})
