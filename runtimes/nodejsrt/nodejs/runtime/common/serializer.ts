/**
 * Utility functions for safe serialization of objects, handling special cases like
 * circular references, non-serializable values, and Axios responses.
 */

/**
 * Checks if a value can be safely serialized to JSON
 */
function isSerializable(value: unknown): boolean {
    if (value === undefined) return false;
    if (value === null) return true;
    if (typeof value === 'function') return false;
    if (typeof value === 'symbol') return false;
    if (typeof value !== 'object') return true;
    return true;
}

/**
 * List of Axios internal properties that should be excluded from serialization
 */
const AXIOS_INTERNAL_PROPS = new Set([
    'transformRequest',
    'transformResponse',
    'adapter',
    'env',
    'transitional',
    'maxBodyLength',
    'maxContentLength',
    'maxRedirects',
    'beforeRedirect',
    'transport',
    'withCredentials',
    'xsrfCookieName',
    'xsrfHeaderName',
    'onUploadProgress',
    'onDownloadProgress',
    'decompress',
    'maxRate',
    'validateStatus'
]);

/**
 * Safely serializes any object, handling:
 * - Circular references (replaced with '[Circular]')
 * - Non-serializable values (functions, symbols, undefined)
 * - Special cases like Axios responses and errors
 * - Arrays and nested objects
 */
export function safeSerialize(obj: unknown): unknown {
    const seen = new WeakSet();

    function serializeValue(value: unknown): unknown {
        // Handle primitive types
        if (!value || typeof value !== 'object') {
            return isSerializable(value) ? value : null;
        }

        // Handle circular references
        if (seen.has(value as object)) {
            return '[Circular]';
        }
        seen.add(value as object);

        // Handle arrays
        if (Array.isArray(value)) {
            return value.map(item => serializeValue(item));
        }

        // Handle Axios error objects
        if ('isAxiosError' in value) {
            const error = value as any;
            return {
                name: error.name,
                message: error.message,
                response: serializeValue(error.response),
                config: serializeValue(error.config)
            };
        }

        // Handle Axios response-like objects
        if ('data' in value && 'status' in value && 'headers' in value) {
            return {
                data: serializeValue((value as any).data),
                status: (value as any).status,
                statusText: (value as any).statusText,
                headers: serializeValue((value as any).headers)
            };
        }

        // Handle Axios config objects
        if ('url' in value && 'method' in value && 'headers' in value) {
            const result: Record<string, unknown> = {};
            for (const [key, val] of Object.entries(value)) {
                if (!AXIOS_INTERNAL_PROPS.has(key) && isSerializable(val)) {
                    result[key] = serializeValue(val);
                }
            }
            return result;
        }

        // Handle regular objects
        const result: Record<string, unknown> = {};
        for (const [key, val] of Object.entries(value)) {
            if (isSerializable(val)) {
                result[key] = serializeValue(val);
            }
        }
        return result;
    }

    return serializeValue(obj);
}

/**
 * Safely converts an object to a JSON string, handling all edge cases
 */
export function safeStringify(obj: unknown): string {
    return JSON.stringify(safeSerialize(obj));
} 