import { listExports, patchCode } from '../ast_utils';
import { Export } from "../pb/autokitteh/user_code/v1/runner_svc_pb";

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
  await ak_call(a, "b", "test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('patch async nested member call', async () => {
    const code = `
    import {a} from "lib"
    
    async function some_func() {
        await a.b.c.d("test");
    }
    `

    const expected = `import { a } from "lib";
async function some_func() {
  await ak_call(a.b.c, "d", "test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('should not patch relative imports', async () => {
    const code = `
    import {localFunc} from "./local"
    
    async function some_func() {
        await localFunc("test");
    }
    `

    const expected = `import { localFunc } from "./local";
async function some_func() {
  await localFunc("test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('should handle multiple await calls in one function', async () => {
    const code = `
    import {service1, service2} from "external"
    
    async function some_func() {
        const result1 = await service1.method1();
        const result2 = await service2.method2();
        return result1 + result2;
    }
    `

    const expected = `import { service1, service2 } from "external";
async function some_func() {
  const result1 = await ak_call(service1, "method1");
  const result2 = await ak_call(service2, "method2");
  return result1 + result2;
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('should handle async arrow functions', async () => {
    const code = `
    import {api} from "external"
    
    const some_func = async () => {
        return await api.getData();
    }
    `

    const expected = `import { api } from "external";
const some_func = async () => {
  return await ak_call(api, "getData");
};`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('should handle TypeScript types', async () => {
    const code = `
    import {api} from "external"
    
    interface Data {
        id: string;
        value: number;
    }
    
    async function getData(): Promise<Data> {
        return await api.fetchData<Data>();
    }
    `

    const expected = `import { api } from "external";
interface Data {
  id: string;
  value: number;
}
async function getData(): Promise<Data> {
  return await ak_call<Data>(api, "fetchData");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})
