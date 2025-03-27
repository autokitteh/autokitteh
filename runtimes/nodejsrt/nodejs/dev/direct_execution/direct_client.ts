import {TextDecoder} from "util";

/**
 * DirectHandlerClient mocks the HandlerService gRPC client.
 * This implementation acts as a bridge between the ActivityWaiter and
 * the Runner's service methods, simulating server behavior.
 */
export class DirectHandlerClient {
    private isActive: boolean = true;
    private runnerService: any = null;

    // Method to set active state - used for testing
    setActive(active: boolean) {
        this.isActive = active;
    }

    // Method to set the runner service after it's captured from createService
    setRunnerService(service: any): void {
        this.runnerService = service;
        console.debug("[DirectClient] Runner service connected");
    }

    // Called by the ActivityWaiter to request an activity
    async activity(request: any) {
        console.debug("[DirectClient] activity called");

        if (!request.data || !this.runnerService) {
            return {error: "Runner service not connected"};
        }

        try {
            // Parse the token from the request data
            const dataStr = new TextDecoder().decode(request.data);
            const parsedData = JSON.parse(dataStr);

            if (!parsedData.token) {
                return {error: "No token in request"};
            }

            // If this is the second call (with callInfo), schedule execution
            if (request.callInfo) {
                console.debug(`[DirectClient] Scheduling execution for token: ${parsedData.token}`);

                // Schedule execution to happen AFTER this method returns
                // This gives the waiter time to set up its listener
                queueMicrotask(() => this.executeActivity(request));
            } else {
                // First call just logging the token
                console.debug(`[DirectClient] Registered token: ${parsedData.token}`);
            }
        } catch (error) {
            console.error("[DirectClient] Error parsing activity data:", error);
            return {error: String(error)};
        }

        // Return success immediately
        return {error: ""};
    }

    // Helper method to execute an activity asynchronously
    private async executeActivity(request: any): Promise<void> {
        try {
            // Parse token from request data
            const dataStr = new TextDecoder().decode(request.data);
            const parsedData = JSON.parse(dataStr);
            const token = parsedData.token;

            console.debug(`[DirectClient] Executing activity for token: ${token}`);

            // Execute via runner service
            const executeResult = await this.runnerService.execute({data: request.data});

            // Send reply
            await this.runnerService.activityReply(executeResult);
            console.debug(`[DirectClient] Activity completed for token: ${token}`);

        } catch (error) {
            console.error("[DirectClient] Execution error:", error);
        }
    }

    // Other required client methods

    async done(request: any) {
        console.debug("[DirectClient] done called:", request);
        return {error: ""};
    }

    async print(request: any) {
        console.debug("[DirectClient] print requested:", request.message);
        return {error: ""};
    }

    async log(request: any) {
        console.debug("[DirectClient] log called:", request.level, request.message);
        return {error: ""};
    }

    async sleep(request: any) {
        console.debug("[DirectClient] sleep called:", request.durationMs, "ms");
        return {error: ""};
    }

    async health() {
        return {error: ""};
    }

    async isActiveRunner() {
        console.debug("[DirectClient] isActiveRunner called, returning:", this.isActive);
        return {isActive: this.isActive, error: ""};
    }

    // Other client methods with minimal implementations

    async subscribe(request: any) {
        console.debug("[DirectClient] subscribe called:", request);
        return {signalId: "mock-signal-" + Date.now(), error: ""};
    }

    async nextEvent(request: any) {
        console.debug("[DirectClient] nextEvent called:", request);
        return {error: ""};
    }

    async unsubscribe(request: any) {
        console.debug("[DirectClient] unsubscribe called:", request);
        return {error: ""};
    }

    async startSession(request: any) {
        console.debug("[DirectClient] startSession called:", request);
        return {sessionId: "mock-session-" + Date.now(), error: ""};
    }

    async encodeJWT(request: any) {
        console.debug("[DirectClient] encodeJWT called:", request);
        return {jwt: "mock-jwt-" + Date.now(), error: ""};
    }

    async refreshOAuthToken(request: any) {
        console.debug("[DirectClient] refreshOAuthToken called:", request);
        return {token: "mock-token-" + Date.now(), error: ""};
    }

    // Direct methods to communicate with the runner service
    async start(request: any) {
        console.debug("[DirectClient] start called with request:", request);

        if (!this.runnerService) {
            throw new Error("Runner service not connected");
        }

        console.debug("[DirectClient] Forwarding start request to runner service");
        return await this.runnerService.start(request);
    }

    async stop() {
        console.debug("[DirectClient] stop called");

        if (!this.runnerService) {
            throw new Error("Runner service not connected");
        }

        console.debug("[DirectClient] Forwarding stop request to runner service");
        return await this.runnerService.stop();
    }
}
