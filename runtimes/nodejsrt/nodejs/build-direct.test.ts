import {buildDirect} from "./dev/direct-execution/build-direct";

test("buildDirect", async () => {
    await buildDirect("../examples/invoices-app", "examples-build/invoices-app");
})
