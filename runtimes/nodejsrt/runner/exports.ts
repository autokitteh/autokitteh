import {listExportsInDirectory} from "./ast_utils";

(async () => {
    const codeDir = process.argv[2];
    const symbols = await listExportsInDirectory(codeDir)
    console.log(JSON.stringify(symbols))
})()

