import { TextDecoder, TextEncoder } from "util";

/**
 * DirectHandlerClient mocks the HandlerService gRPC client.
 * This implementation acts as a bridge between the ActivityWaiter and 
 * the Runner's service methods, simulating server behavior.
 */
export class DirectHandlerClient {
    private originalConsoleLog: typeof console.log;
    private isActive: boolean = true;
    private runnerService: any = null;
    
    constructor() {
        // Store the original console.log
        this.originalConsoleLog = console.log;
    }

    // Method to set active state - used for testing
    setActive(active: boolean) {
        this.isActive = active;
        this.originalConsoleLog("[DirectClient] Setting isActive to:", active);
    }

    // Method to set the runner service after it's captured from createService
    setRunnerService(service: any): void {
        this.runnerService = service;
        this.originalConsoleLog("[DirectClient] Runner service connected");
    }

    // Called by the ActivityWaiter to request an activity
    async activity(request: any) {
        this.originalConsoleLog("[DirectClient] activity called");
        
        if (!request.data || !this.runnerService) {
            return { error: "Runner service not connected" };
        }
        
        try {
            // Parse the token from the request data
            const dataStr = new TextDecoder().decode(request.data);
            const parsedData = JSON.parse(dataStr);
            
            if (!parsedData.token) {
                return { error: "No token in request" };
            }
            
            // If this is the second call (with callInfo), schedule execution
            if (request.callInfo) {
                this.originalConsoleLog(`[DirectClient] Scheduling execution for token: ${parsedData.token}`);
                
                // Schedule execution to happen AFTER this method returns
                // This gives the waiter time to set up its listener
                queueMicrotask(() => this.executeActivity(request));
            } else {
                // First call just logging the token
                this.originalConsoleLog(`[DirectClient] Registered token: ${parsedData.token}`);
            }
        } catch (error) {
            this.originalConsoleLog("[DirectClient] Error parsing activity data:", error);
            return { error: String(error) };
        }
        
        // Return success immediately
        return { error: "" };
    }
    
    // Helper method to execute an activity asynchronously
    private async executeActivity(request: any): Promise<void> {
        try {
            // Parse token from request data
            const dataStr = new TextDecoder().decode(request.data);
            const parsedData = JSON.parse(dataStr);
            const token = parsedData.token;
            
            this.originalConsoleLog(`[DirectClient] Executing activity for token: ${token}`);
            
            // Execute via runner service
            const executeResult = await this.runnerService.execute({
                data: request.data
            });
            
            // Send reply
            const encoder = new TextEncoder();
            await this.runnerService.activityReply({
                error: "",
                result: {
                    custom: {
                        data: encoder.encode(JSON.stringify({
                            token: token,
                            results: executeResult.value
                        })),
                        executorId: ""
                    }
                }
            });
            
            this.originalConsoleLog(`[DirectClient] Activity completed for token: ${token}`);
        } catch (error) {
            this.originalConsoleLog("[DirectClient] Execution error:", error);
        }
    }

    // Other required client methods

    async done(request: any) {
        this.originalConsoleLog("[DirectClient] done called:", request);
        return { error: "" };
    }

    async print(request: any) {
        // Use the original console.log to avoid infinite recursion
        this.originalConsoleLog("[DirectClient] print requested:", request.message);
        return { error: "" };
    }
    
    async log(request: any) {
        this.originalConsoleLog("[DirectClient] log called:", request.level, request.message);
        return { error: "" };
    }
    
    async sleep(request: any) {
        this.originalConsoleLog("[DirectClient] sleep called:", request.durationMs, "ms");
        return { error: "" };
    }
    
    async health() {
        return { error: "" };
    }
    
    async isActiveRunner() {
        this.originalConsoleLog("[DirectClient] isActiveRunner called, returning:", this.isActive);
        return { isActive: this.isActive, error: "" };
    }
    
    // Other client methods with minimal implementations
    
    async subscribe(request: any) {
        this.originalConsoleLog("[DirectClient] subscribe called:", request);
        return { signalId: "mock-signal-" + Date.now(), error: "" };
    }
    
    async nextEvent(request: any) {
        this.originalConsoleLog("[DirectClient] nextEvent called:", request);
        return { error: "" };
    }
    
    async unsubscribe(request: any) {
        this.originalConsoleLog("[DirectClient] unsubscribe called:", request);
        return { error: "" };
    }
    
    async startSession(request: any) {
        this.originalConsoleLog("[DirectClient] startSession called:", request);
        return { sessionId: "mock-session-" + Date.now(), error: "" };
    }
    
    async encodeJWT(request: any) {
        this.originalConsoleLog("[DirectClient] encodeJWT called:", request);
        return { jwt: "mock-jwt-" + Date.now(), error: "" };
    }
    
    async refreshOAuthToken(request: any) {
        this.originalConsoleLog("[DirectClient] refreshOAuthToken called:", request);
        return { token: "mock-token-" + Date.now(), error: "" };
    }
} 