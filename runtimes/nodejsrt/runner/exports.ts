import {listSymbolsInDirectory} from "./ast_utils";

(async () => {
    const codeDir = process.argv[2];
    const symbols = await listSymbolsInDirectory(codeDir)
    console.log(JSON.stringify(symbols))
})()

