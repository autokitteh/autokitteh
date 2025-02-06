import { fastify } from "fastify";
import { fastifyConnectPlugin } from "@connectrpc/connect-fastify";
import {createService} from "./server"
import { Command} from "commander";
const program = new Command();

program
    .requiredOption('--worker-address <TYPE>', 'worker address')
    .requiredOption('--port <TYPE>', 'port', parseInt)
    .requiredOption('--runner-id <TYPE>', 'runner ID')
    .requiredOption('--code-dir <TYPE>', 'user code directory')

const options =  program.opts();

console.log(process.argv);

program.parse(process.argv);

let routes = createService(options.codeDir, options.runnerId, options.workerAddress);

async function main() {
    const server = fastify({http2: true});
    await server.register(fastifyConnectPlugin, {routes});
    server.get("/", (_, reply) => {
        reply.type("text/plain");
        reply.send("Hello World!");
    });
    await server.listen({ host: "localhost", port: options.port });
    console.log("server is listening at", server.addresses(), "code dir:", options.codeDir);
}

void main();
