import { listExports, patchCode } from '../runtime/common/ast_utils';
import { Export } from "../runtime/pb/autokitteh/user_code/v1/runner_svc_pb";

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
  await (global as any).ak_call(f, "test");
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
  await (global as any).ak_call(a.b, "test");
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
  await (global as any).ak_call(a.b.c.d, "test");
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
  const result1 = await (global as any).ak_call(service1.method1);
  const result2 = await (global as any).ak_call(service2.method2);
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
  return await (global as any).ak_call(api.getData);
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
  return await (global as any).ak_call<Data>(api.fetchData);
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('should not patch method calls on objects from relative imports', async () => {
    const code = `
    import {localService} from "./local-service"
    
    async function some_func() {
        await localService.method("test");
    }
    `

    const expected = `import { localService } from "./local-service";
async function some_func() {
  await localService.method("test");
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('should not patch calls to native objects', async () => {
    const code = `
    async function some_func() {
        await console.log("test");
        await Promise.resolve("value");
        await JSON.parse('{"key": "value"}');
    }
    `

    const expected = `async function some_func() {
  await console.log("test");
  await Promise.resolve("value");
  await JSON.parse('{"key": "value"}');
}`
    const patch = await patchCode(code)
    expect(patch).toEqual(expected);
})

test('should handle locally declared function calls', async () => {
    const code = `
    async function helper() {
        return "helper result";
    }
    
    async function main() {
        return await helper();
    }
    `;

    const expected = `async function helper() {
  return "helper result";
}
async function main() {
  return await helper();
}`; // Should NOT wrap local function calls
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should handle objects created from local constructors', async () => {
    const code = `
    class LocalService {
        async getData() {
            return "data";
        }
    }
    
    async function main() {
        const service = new LocalService();
        return await service.getData();
    }
    `;

    const expected = `class LocalService {
  async getData() {
    return "data";
  }
}
async function main() {
  const service = new LocalService();
  return await service.getData();
}`; // Should NOT wrap methods on local class instances
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should handle objects created from imported constructors', async () => {
    const code = `
    import { ExternalService } from "external-package";
    import { LocalService } from "./local-service";
    
    async function main() {
        const externalService = new ExternalService();
        const localService = new LocalService();
        
        const externalResult = await externalService.getData();
        const localResult = await localService.getData();
        
        return { externalResult, localResult };
    }
    `;

    const expected = `import { ExternalService } from "external-package";
import { LocalService } from "./local-service";
async function main() {
  const externalService = new ExternalService();
  const localService = new LocalService();
  const externalResult = await (global as any).ak_call(externalService.getData);
  const localResult = await localService.getData();
  return {
    externalResult,
    localResult
  };
}`; // Should wrap external but not local service
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should handle nested member expressions', async () => {
    const code = `
    import { client } from "external-api";
    import { localClient } from "./local-client";
    
    async function main() {
        const externalResult = await client.api.methods.call();
        const localResult = await localClient.api.methods.call();
        return { externalResult, localResult };
    }
    `;

    const expected = `import { client } from "external-api";
import { localClient } from "./local-client";
async function main() {
  const externalResult = await (global as any).ak_call(client.api.methods.call);
  const localResult = await localClient.api.methods.call();
  return {
    externalResult,
    localResult
  };
}`; // Should wrap nested methods on external but not local objects
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should properly handle axios.get method call', async () => {
    const code = `
    import axios from "axios";
    
    async function fetchUserData(userId) {
        const response = await axios.get(\`\${this.baseUrl}/users/\${userId}\`);
        return response.data;
    }
    `;

    const expected = `import axios from "axios";
async function fetchUserData(userId) {
  const response = await (global as any).ak_call(axios.get, \`\${this.baseUrl}/users/\${userId}\`);
  return response.data;
}`;
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should wrap axios.get properly if passed as reference', async () => {
    const code = `
    import axios from "axios";
    
    async function fetchUserData(userId) {
        const getFunc = axios.get;
        const response = await getFunc(\`\${this.baseUrl}/users/\${userId}\`);
        return response.data;
    }
    `;

    const expected = `import axios from "axios";
async function fetchUserData(userId) {
  const getFunc = axios.get;
  const response = await (global as any).ak_call(getFunc, \`\${this.baseUrl}/users/\${userId}\`);
  return response.data;
}`;
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should transform autokitteh.subscribe calls', async () => {
    const code = `
    import { autokitteh } from 'autokitteh';
    
    async function handleEvents() {
        const subId = await autokitteh.subscribe('my-source', 'event.type == "test"');
        return subId;
    }
    `;

    const expected = `/* import { autokitteh } from 'autokitteh'; - commented out by autokitteh build process*/
async function handleEvents() {
  const subId = await (global as any).syscalls.subscribe('my-source', 'event.type == "test"');
  return subId;
}`;
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should transform autokitteh.nextEvent calls', async () => {
    const code = `
    import { autokitteh } from 'autokitteh';
    
    async function waitForEvent(subId) {
        const event = await autokitteh.nextEvent(subId, {
            timeout: 30
        });
        return event;
    }
    `;

    const expected = `/* import { autokitteh } from 'autokitteh'; - commented out by autokitteh build process*/
async function waitForEvent(subId) {
  const event = await (global as any).syscalls.nextEvent(subId, {
    timeout: 30
  });
  return event;
}`;
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should transform autokitteh.unsubscribe calls', async () => {
    const code = `
    import { autokitteh } from 'autokitteh';
    
    async function cleanup(subId) {
        await autokitteh.unsubscribe(subId);
    }
    `;

    const expected = `/* import { autokitteh } from 'autokitteh'; - commented out by autokitteh build process*/
async function cleanup(subId) {
  await (global as any).syscalls.unsubscribe(subId);
}`;
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should handle multiple autokitteh method calls in one function', async () => {
    const code = `
    import { autokitteh } from 'autokitteh';
    
    async function handleEventFlow() {
        const subId = await autokitteh.subscribe('source', 'filter');
        const event = await autokitteh.nextEvent(subId);
        await autokitteh.unsubscribe(subId);
        return event;
    }
    `;

    const expected = `/* import { autokitteh } from 'autokitteh'; - commented out by autokitteh build process*/
async function handleEventFlow() {
  const subId = await (global as any).syscalls.subscribe('source', 'filter');
  const event = await (global as any).syscalls.nextEvent(subId);
  await (global as any).syscalls.unsubscribe(subId);
  return event;
}`;
    const patch = await patchCode(code);
    expect(patch).toEqual(expected);
});

test('should only patch direct autokitteh event methods', async () => {
    const code = `
    import { autokitteh } from 'autokitteh';
    import { google } from 'autokitteh/google';
    
    async function handleEvents() {
        // These should be patched
        const subId = await autokitteh.subscribe('my-source', 'filter');
        const event = await autokitteh.nextEvent(subId);
        await autokitteh.unsubscribe(subId);

        // These should NOT be patched
        const gmail = await autokitteh.google.gmail_client('my_gmail');
        const calendar = await autokitteh.google.calendar_client('my_calendar');
    }
    `;

    const patch = await patchCode(code);
    
    // Check that standalone imports are gone (we only want them in comments)
    expect(patch.match(/^import\s+{.*?}\s+from\s+['"]autokitteh.*?['"]/m)).toBe(null);
    expect(patch.match(/^import\s+{.*?}\s+from\s+['"]autokitteh\/google.*?['"]/m)).toBe(null);
    
    // Check that comments about removed imports exist
    expect(patch).toContain('import { autokitteh } from \'autokitteh\'; - commented out by autokitteh build process');
    expect(patch).toContain('import { google } from \'autokitteh/google\'; - commented out by autokitteh build process');
    
    // Check that syscalls are properly transformed
    expect(patch).toContain('await (global as any).syscalls.subscribe(\'my-source\', \'filter\')');
    expect(patch).toContain('await (global as any).syscalls.nextEvent(subId)');
    expect(patch).toContain('await (global as any).syscalls.unsubscribe(subId)');
    
    // Check that other method calls are properly transformed
    expect(patch).toContain('await (global as any).ak_call(autokitteh.google.gmail_client, \'my_gmail\')');
    expect(patch).toContain('await (global as any).ak_call(autokitteh.google.calendar_client, \'my_calendar\')');
});

test('should comment out autokitteh imports', async () => {
    const code = `
    import { autokitteh } from 'autokitteh';
    
    // Function that doesn't use autokitteh
    function processData(data) {
        return data;
    }
    `;

    const patched = await patchCode(code);
    
    // The import should be converted to a comment
    expect(patched).toContain('commented out by autokitteh build process');
    
    // Import statement should not appear as a standalone import
    expect(patched).not.toContain('import "__autokitteh_dummy__"');
    
    // The function should still be present and unchanged
    expect(patched).toContain('function processData(data)');
});
