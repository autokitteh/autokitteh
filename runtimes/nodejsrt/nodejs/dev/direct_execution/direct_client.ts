import { DirectActivityWaiter, IDirectHandlerClient } from "./direct_waiter";
import { TextDecoder } from "util";

/**
 * DirectHandlerClient mocks the HandlerService gRPC client.
 * This implementation works with DirectActivityWaiter to simulate the activity flow.
 */
export class DirectHandlerClient implements IDirectHandlerClient {
    private originalConsoleLog: typeof console.log;
    private isActive: boolean = true;
    private waiter: DirectActivityWaiter;

    constructor(waiter: DirectActivityWaiter) {
        // Store the original console.log
        this.originalConsoleLog = console.log;
        this.waiter = waiter;
        
        // Set the client reference in the waiter
        waiter.setClient(this);
    }

    // Method to set active state - used for testing
    setActive(active: boolean) {
        this.isActive = active;
        this.originalConsoleLog("[DirectClient] Setting isActive to:", active);
    }

    async activity(request: any) {
        this.originalConsoleLog("[DirectClient] activity called:", request);
        
        // If this is an activity request with a token, process it
        if (request.data) {
            try {
                // Parse the token from the request data
                const dataStr = new TextDecoder().decode(request.data);
                const parsedData = JSON.parse(dataStr);
                
                if (parsedData.token) {
                    // In a real implementation, this would trigger the handler service
                    // For our direct execution, we'll execute the function right away
                    this.originalConsoleLog("[DirectClient] Processing activity with token:", parsedData.token);
                    
                    setTimeout(async () => {
                        try {
                            // Execute the function
                            const result = await this.waiter.execute_signal(parsedData.token);
                            
                            // Reply with the result
                            await this.waiter.reply_signal(parsedData.token, result);
                        } catch (error) {
                            this.originalConsoleLog("[DirectClient] Error executing activity:", error);
                        }
                    }, 0); // Use setTimeout to make this async
                }
            } catch (error) {
                this.originalConsoleLog("[DirectClient] Error parsing activity data:", error);
            }
        }
        
        return { error: "" };
    }

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