"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g = Object.create((typeof Iterator === "function" ? Iterator : Object).prototype);
    return g.next = verb(0), g["throw"] = verb(1), g["return"] = verb(2), typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (g && (g = 0, op[0] && (_ = 0)), _) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
Object.defineProperty(exports, "__esModule", { value: true });
var globals_1 = require("@jest/globals");
var runner_1 = require("../runner");
var events_1 = require("events");
// Mock the HandlerService client
globals_1.jest.mock('@connectrpc/connect', function () { return ({
    createClient: globals_1.jest.fn()
}); });
(0, globals_1.describe)('Runner', function () {
    var runner;
    var mockClient;
    (0, globals_1.beforeEach)(function () {
        // Reset all mocks
        globals_1.jest.clearAllMocks();
        // Create mock client with proper response types
        mockClient = {
            health: globals_1.jest.fn().mockResolvedValue({
                error: ''
            }),
            isActiveRunner: globals_1.jest.fn().mockResolvedValue({
                error: '',
                isActive: true
            }),
            print: globals_1.jest.fn().mockResolvedValue({
                error: ''
            }),
        };
        // Create runner instance
        runner = new runner_1.default('test-runner-id', '/test/code/dir', mockClient);
        // Mock console methods
        console.log = globals_1.jest.fn();
        console.error = globals_1.jest.fn();
    });
    (0, globals_1.afterEach)(function () {
        // Clean up any timers
        globals_1.jest.useRealTimers();
    });
    (0, globals_1.describe)('initialization', function () {
        (0, globals_1.test)('should create runner with correct properties', function () {
            (0, globals_1.expect)(runner).toBeInstanceOf(runner_1.default);
            (0, globals_1.expect)(runner.id).toBe('test-runner-id');
            (0, globals_1.expect)(runner.codeDir).toBe('/test/code/dir');
            (0, globals_1.expect)(runner.client).toBe(mockClient);
            (0, globals_1.expect)(runner.events).toBeInstanceOf(events_1.EventEmitter);
            (0, globals_1.expect)(runner.isStarted).toBe(false);
        });
    });
    (0, globals_1.describe)('start', function () {
        (0, globals_1.beforeEach)(function () {
            globals_1.jest.useFakeTimers();
        });
        (0, globals_1.test)('should start health checks and set up event listener', function () { return __awaiter(void 0, void 0, void 0, function () {
            var startPromise;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        startPromise = runner.start();
                        // Verify health check is started
                        (0, globals_1.expect)(mockClient.isActiveRunner).toHaveBeenCalledWith({
                            runnerId: 'test-runner-id'
                        });
                        // Emit the started event
                        runner.events.emit('started');
                        // Fast-forward timers
                        globals_1.jest.advanceTimersByTime(1000);
                        return [4 /*yield*/, startPromise];
                    case 1:
                        _a.sent();
                        (0, globals_1.expect)(runner.isStarted).toBe(true);
                        return [2 /*return*/];
                }
            });
        }); });
        (0, globals_1.test)('should not allow multiple starts', function () { return __awaiter(void 0, void 0, void 0, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0: return [4 /*yield*/, runner.start()];
                    case 1:
                        _a.sent();
                        runner.events.emit('started');
                        return [4 /*yield*/, (0, globals_1.expect)(runner.start()).rejects.toThrow('Runner already started')];
                    case 2:
                        _a.sent();
                        return [2 /*return*/];
                }
            });
        }); });
        (0, globals_1.test)('should handle health check failures with retries', function () { return __awaiter(void 0, void 0, void 0, function () {
            var startPromise;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        mockClient.isActiveRunner
                            .mockRejectedValueOnce(new Error('Connection failed'))
                            .mockRejectedValueOnce(new Error('Connection failed'))
                            .mockResolvedValueOnce({
                            error: '',
                            isActive: true
                        });
                        startPromise = runner.start();
                        // Fast-forward past two failures and one success
                        globals_1.jest.advanceTimersByTime(3000);
                        return [4 /*yield*/, startPromise];
                    case 1:
                        _a.sent();
                        (0, globals_1.expect)(mockClient.isActiveRunner).toHaveBeenCalledTimes(3);
                        (0, globals_1.expect)(console.error).toHaveBeenCalledWith('Health check error:', globals_1.expect.any(Error));
                        return [2 /*return*/];
                }
            });
        }); });
    });
    (0, globals_1.describe)('stop', function () {
        (0, globals_1.test)('should clear health check timer and emit stop event', function () { return __awaiter(void 0, void 0, void 0, function () {
            var stopListener;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        stopListener = globals_1.jest.fn();
                        runner.events.on('stop', stopListener);
                        return [4 /*yield*/, runner.start()];
                    case 1:
                        _a.sent();
                        runner.events.emit('started');
                        runner.stop();
                        (0, globals_1.expect)(stopListener).toHaveBeenCalled();
                        (0, globals_1.expect)(runner.healthcheckTimer).toBeUndefined();
                        return [2 /*return*/];
                }
            });
        }); });
    });
    (0, globals_1.describe)('akPrint', function () {
        (0, globals_1.test)('should format and send print messages', function () { return __awaiter(void 0, void 0, void 0, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0: return [4 /*yield*/, runner.akPrint('test', 123, { foo: 'bar' })];
                    case 1:
                        _a.sent();
                        (0, globals_1.expect)(mockClient.print).toHaveBeenCalledWith({
                            runnerId: 'test-runner-id',
                            message: 'test 123 [object Object]'
                        });
                        return [2 /*return*/];
                }
            });
        }); });
        (0, globals_1.test)('should handle print failures gracefully', function () { return __awaiter(void 0, void 0, void 0, function () {
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        mockClient.print.mockRejectedValue(new Error('Print failed'));
                        return [4 /*yield*/, runner.akPrint('test')];
                    case 1:
                        _a.sent();
                        (0, globals_1.expect)(console.log).toHaveBeenCalledWith('Failed to send print message:', globals_1.expect.any(Error));
                        return [2 /*return*/];
                }
            });
        }); });
    });
    (0, globals_1.describe)('graceful shutdown', function () {
        (0, globals_1.test)('should handle shutdown process correctly', function () { return __awaiter(void 0, void 0, void 0, function () {
            var processExitSpy;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        processExitSpy = globals_1.jest.spyOn(process, 'exit').mockImplementation(function () { return undefined; });
                        return [4 /*yield*/, runner.start()];
                    case 1:
                        _a.sent();
                        runner.events.emit('started');
                        return [4 /*yield*/, runner.gracefulShutdown()];
                    case 2:
                        _a.sent();
                        (0, globals_1.expect)(processExitSpy).toHaveBeenCalledWith(0);
                        processExitSpy.mockRestore();
                        return [2 /*return*/];
                }
            });
        }); });
        (0, globals_1.test)('should handle shutdown errors', function () { return __awaiter(void 0, void 0, void 0, function () {
            var processExitSpy, error;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        processExitSpy = globals_1.jest.spyOn(process, 'exit').mockImplementation(function () { return undefined; });
                        error = new Error('Shutdown error');
                        return [4 /*yield*/, runner.start()];
                    case 1:
                        _a.sent();
                        runner.events.emit('started');
                        // Mock stop to throw an error
                        globals_1.jest.spyOn(runner, 'stop').mockImplementation(function () {
                            throw error;
                        });
                        return [4 /*yield*/, runner.gracefulShutdown()];
                    case 2:
                        _a.sent();
                        (0, globals_1.expect)(console.error).toHaveBeenCalledWith('Error during shutdown:', error);
                        (0, globals_1.expect)(processExitSpy).toHaveBeenCalledWith(1);
                        processExitSpy.mockRestore();
                        return [2 /*return*/];
                }
            });
        }); });
    });
});
