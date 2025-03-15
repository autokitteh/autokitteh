declare global {
    interface Global {
        ak_call: (...args: unknown[]) => Promise<unknown>;
    }
} 