import {listSymbols, Symbol, listSymbolsInDirectory, Symbols, patchCode} from './ast_utils';

test('test list symbols', async () => {
    const code = `
    import {a} from './a';
    
    function some_func(a, b) {
        return a + b;
    }
    function another_func() {}
    `;

    const expected: Symbol[] = [
        {line: 4, name: "some_func", args: ["a", "b"]},
        {line: 7, name: "another_func", args: []},
    ];
    expect(await listSymbols(code)).toEqual(expected);
});

test('list symbols in directory', async () => {
    const dir = "./test_data/list_symbols";
    const actual = await listSymbolsInDirectory(dir)
    const expected:  Symbols = {
        "test_data/list_symbols/a.js": [],
        "test_data/list_symbols/dep.js": [{line: 1, name: "test_func", args: []}],
    }

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
