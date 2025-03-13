import { fastify } from "fastify";
import { fastifyConnectPlugin } from "@connectrpc/connect-fastify";
import { Command } from "commander";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import { HandlerService } from "./pb/autokitteh/user_code/v1/handler_svc_pb";
import Runner from "./runner";

const program = new Command();

program
    .requiredOption('--worker-address <TYPE>', 'worker address')
    .requiredOption('--port <TYPE>', 'port', parseInt)
    .requiredOption('--runner-id <TYPE>', 'runner ID')
    .requiredOption('--code-dir <TYPE>', 'user code directory');

program.parse(process.argv);
const options = program.opts();

async function main() {
    console.log("Starting main with options:", options);
    
    const transport = createGrpcTransport({
        baseUrl: options.workerAddress,
        nodeOptions: { rejectUnauthorized: false }
    });

    console.log("Creating gRPC client...");
    const client = createClient(HandlerService, transport);
    const runner = new Runner(options.runnerId, options.codeDir, client);
    
    console.log("Creating Fastify server...");
    const server = fastify({ http2: true });
    await server.register(fastifyConnectPlugin, {
        routes: runner.createService()
    });

    server.get("/", (_, reply) => {
        reply.type("text/plain");
        reply.send("Hello World!");
    });

    // Handle graceful shutdown
    const signals = ["SIGTERM", "SIGINT"];
    signals.forEach(signal => {
        process.on(signal, () => {
            console.log(`Received ${signal}, shutting down...`);
            runner.stop();
            server.close(() => {
                console.log("Server closed");
                process.exit(0);
            });
        });
    });

    try {
        console.log("Starting runner...");
        await runner.start();
        
        console.log("Starting server...");
        await server.listen({ host: "localhost", port: options.port });
        console.log("Server is listening at", server.addresses(), "code dir:", options.codeDir);
        
        // Wait a bit to ensure everything is ready
        await new Promise(resolve => setTimeout(resolve, 1000));
        console.log("Emitting started event...");
        runner.emit("started");
    } catch (err) {
        console.error("Failed to start server:", err);
        process.exit(1);
    }
}

void main();
