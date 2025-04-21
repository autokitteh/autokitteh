import {buildDirect} from "./build-direct";

test("buildDirect", async () => {
    await buildDirect("../examples/invoices-app", "build/invoices-app");
})
