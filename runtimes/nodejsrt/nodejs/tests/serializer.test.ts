import { safeSerialize, safeStringify } from '../common/serializer';
import axios from 'axios';
import nock from 'nock';

describe('serializer utility', () => {
    describe('safeSerialize', () => {
        test('handles primitive values', () => {
            expect(safeSerialize(42)).toBe(42);
            expect(safeSerialize('hello')).toBe('hello');
            expect(safeSerialize(true)).toBe(true);
            expect(safeSerialize(null)).toBe(null);
            expect(safeSerialize(undefined)).toBe(null);
        });

        test('handles arrays', () => {
            const arr = [1, 'two', { three: 3 }];
            const result = safeSerialize(arr);
            expect(result).toEqual([1, 'two', { three: 3 }]);
        });

        test('handles nested objects', () => {
            const obj = {
                a: 1,
                b: {
                    c: 2,
                    d: {
                        e: 3
                    }
                }
            };
            const result = safeSerialize(obj);
            expect(result).toEqual(obj);
        });

        test('handles circular references', () => {
            const obj: any = { a: 1 };
            obj.self = obj;
            obj.nested = { ref: obj };

            const result = safeSerialize(obj) as any;
            expect(result.a).toBe(1);
            expect(result.self).toBe('[Circular]');
            expect(result.nested.ref).toBe('[Circular]');
        });

        test('handles non-serializable values', () => {
            const obj = {
                fn: () => {},
                sym: Symbol('test'),
                undef: undefined,
                normal: 'value'
            };

            const result = safeSerialize(obj) as any;
            expect(result.fn).toBeUndefined();
            expect(result.sym).toBeUndefined();
            expect(result.undef).toBeUndefined();
            expect(result.normal).toBe('value');
        });

        test('handles Axios response-like objects', () => {
            const axiosResponse = {
                data: { id: 1, name: 'test' },
                status: 200,
                statusText: 'OK',
                headers: { 'content-type': 'application/json' },
                config: { timeout: 1000 }, // should be excluded
                request: {}, // should be excluded
            };

            const result = safeSerialize(axiosResponse) as any;
            expect(result).toEqual({
                data: { id: 1, name: 'test' },
                status: 200,
                statusText: 'OK',
                headers: { 'content-type': 'application/json' }
            });
            expect(result.config).toBeUndefined();
            expect(result.request).toBeUndefined();
        });

        test('handles complex nested Axios responses', () => {
            const complexResponse = {
                data: {
                    user: {
                        id: 1,
                        responses: [
                            {
                                data: { value: 'nested' },
                                status: 200,
                                headers: { 'x-test': 'true' }
                            }
                        ]
                    }
                },
                status: 200,
                headers: {}
            };

            const result = safeSerialize(complexResponse) as any;
            expect(result.data.user.responses[0].data.value).toBe('nested');
            expect(result.data.user.responses[0].status).toBe(200);
        });

        test('handles arrays with circular references', () => {
            const arr: any[] = [1, 2, 3];
            arr.push(arr);
            const obj = { ref: arr };
            arr.push(obj);

            const result = safeSerialize(arr) as any[];
            expect(result[0]).toBe(1);
            expect(result[1]).toBe(2);
            expect(result[2]).toBe(3);
            expect(result[3]).toBe('[Circular]');
            expect(result[4].ref).toBe('[Circular]');
        });
    });

    describe('safeStringify', () => {
        test('converts object to JSON string safely', () => {
            const obj = {
                num: 42,
                str: 'hello',
                nested: {
                    arr: [1, 2, { value: 3 }]
                }
            };
            const result = safeStringify(obj);
            expect(JSON.parse(result)).toEqual(obj);
        });

        test('handles circular references in stringification', () => {
            const obj: any = { a: 1 };
            obj.self = obj;

            const result = safeStringify(obj);
            const parsed = JSON.parse(result);
            expect(parsed.a).toBe(1);
            expect(parsed.self).toBe('[Circular]');
        });

        test('produces valid JSON for Axios responses', () => {
            const axiosResponse = {
                data: { id: 1 },
                status: 200,
                statusText: 'OK',
                headers: { 'content-type': 'application/json' },
                config: { timeout: 1000 }
            };

            const result = safeStringify(axiosResponse);
            const parsed = JSON.parse(result);
            expect(parsed.data).toEqual({ id: 1 });
            expect(parsed.status).toBe(200);
            expect(parsed.config).toBeUndefined();
        });
    });

    describe('safeSerialize with real Axios', () => {
        beforeAll(() => {
            // Mock API responses
            nock('https://api.example.com')
                .get('/data')
                .reply(200, { message: 'success' }, {
                    'content-type': 'application/json'
                });

            nock('https://api.example.com')
                .get('/error')
                .reply(404, { error: 'not found' });
        });

        afterAll(() => {
            nock.cleanAll();
        });

        test('handles real axios.get success response', async () => {
            const response = await axios.get('https://api.example.com/data');
            const result = safeSerialize(response) as any;

            // Check that we have the essential Axios response properties
            expect(result.data).toEqual({ message: 'success' });
            expect(result.status).toBe(200);
            expect(result.statusText).toBeDefined();
            expect(result.headers).toBeDefined();

            // Check that internal Axios properties are excluded
            expect(result.config).toBeUndefined();
            expect(result.request).toBeUndefined();
            expect(result.isAxiosResponse).toBeUndefined();
        });

        test('handles real axios.get error response', async () => {
            try {
                await axios.get('https://api.example.com/error');
                fail('Expected axios.get to throw an error');
            } catch (error) {
                const result = safeSerialize(error) as any;

                // Check error response structure
                expect(result.response.data).toEqual({ error: 'not found' });
                expect(result.response.status).toBe(404);
                expect(result.response.headers).toBeDefined();

                // Verify that we have the essential error properties
                expect(result.message).toBeDefined();
                expect(result.name).toBeDefined();

                // Verify that sensitive/internal properties are excluded or properly handled
                expect(result.stack).toBeUndefined();
                expect(result.cause).toBeUndefined();
                
                // The config object should be serialized but simplified
                if (result.config) {
                    expect(result.config.url).toBeDefined();
                    expect(result.config.method).toBeDefined();
                    // Internal implementation details should be excluded
                    expect(result.config.transformRequest).toBeUndefined();
                    expect(result.config.transformResponse).toBeUndefined();
                    expect(result.config.adapter).toBeUndefined();
                }
            }
        });

        test('handles real axios.get with nested response data', async () => {
            // Mock a complex nested response
            nock('https://api.example.com')
                .get('/nested')
                .reply(200, {
                    users: [
                        { id: 1, details: { name: 'John', age: 30 } },
                        { id: 2, details: { name: 'Jane', age: 25 } }
                    ],
                    metadata: {
                        total: 2,
                        page: 1
                    }
                });

            const response = await axios.get('https://api.example.com/nested');
            const result = safeSerialize(response) as any;

            // Verify the nested data structure is preserved
            expect(result.data.users).toHaveLength(2);
            expect(result.data.users[0].details.name).toBe('John');
            expect(result.data.metadata.total).toBe(2);

            // Verify Axios response structure
            expect(result.status).toBe(200);
            expect(result.headers).toBeDefined();
            expect(result.config).toBeUndefined();
        });
    });
}); 