import { ak_call, Waiter } from './ak_call';
import { SysCalls } from './syscalls';

/**
 * Initializes global objects and functions that will be available to user code.
 *
 * @param waiter - The activity waiter instance
 * @param syscalls - The syscalls instance
 */
export function initializeGlobals(waiter: Waiter, syscalls: SysCalls) {
    // Initialize ak_call
    Object.defineProperty(global, 'ak_call', {
        value: ak_call(waiter),
        configurable: false,
        writable: false
    });

    // Set the global syscalls object
    Object.defineProperty(global, 'syscalls', {
        value: syscalls,
        configurable: false,
        writable: false
    });
}
