export async function helloWorld(name: string): Promise<string> {
    return `Hello, ${name}!`;
}

export async function add(a: number, b: number): Promise<number> {
    return a + b;
}

export async function fetchData(url: string): Promise<unknown> {
    const axios = await import('axios');
    const response = await axios.default.get(url);
    return response.data;
}

// Function with ak_call flag for testing remote execution
export async function remoteFunction(input: string): Promise<string> {
    return `Remote: ${input}`;
}
remoteFunction.ak_call = true; 