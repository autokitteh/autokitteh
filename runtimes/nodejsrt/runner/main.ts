import { fastify } from "fastify";
import { fastifyConnectPlugin } from "@connectrpc/connect-fastify";
import {createService} from "./server"
import { Command} from "commander";
const program = new Command();
import {Sandbox} from "./sandbox";
import {ak_call, ActivityWaiter} from "./ak_call";
import {createClient} from "@connectrpc/connect";
import {createGrpcTransport} from "@connectrpc/connect-node";
import {HandlerService} from "./pb/autokitteh/user_code/v1/handler_svc_pb";

program
    .requiredOption('--worker-address <TYPE>', 'worker address')
    .requiredOption('--port <TYPE>', 'port', parseInt)
    .requiredOption('--runner-id <TYPE>', 'runner ID')
    .requiredOption('--code-dir <TYPE>', 'user code directory')

const options =  program.opts();

console.log(process.argv);

program.parse(process.argv);

const transport = createGrpcTransport({
    baseUrl: `http://${options.workerAddress}`,
});

const client = createClient(HandlerService, transport);
const waiter = new ActivityWaiter(client, options.runnerId)
const _ak_call = ak_call(waiter)
const sandbox = new Sandbox(options.codeDir, _ak_call);
const routes = createService(options.codeDir, options.runnerId, sandbox, waiter);


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
