
import { main } from './testdata/simple-test/runtime/.ak/runtime/runner/main';

const options = {
    workerAddress: 'localhost:51975',
    port: 51982,
    runnerId: "runner_01jqkvrkb3e7kbn83xrs3e3tyr",
    codeDir: "testdata/simple-test"
}
main(options).catch(err => {
    console.error("Error in main:", err);
    process.exit(1);
});

