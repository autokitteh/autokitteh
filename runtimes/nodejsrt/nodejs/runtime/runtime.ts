import { ak_call, Waiter } from './ak_call';

export function initializeGlobals(waiter: Waiter, projectRoot: string) {
    Object.defineProperty(global, 'ak_call', {
        value: ak_call(waiter, projectRoot),
        configurable: false,
        writable: false
    });
} 