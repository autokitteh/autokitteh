import {mainDirect} from "./main_direct";
import {Command} from "commander";

const program = new Command();

program.requiredOption('--input-dir <TYPE>', 'inputDir')

program.parse(process.argv);
const options = program.opts();

async function main() {
    await mainDirect(options.inputDir);
}
void main();
