import {getData} from "./lib";

import {promises} from "fs"

async function on_event(a: any, b: any) {
    console.log("event args: ", a, b)
    console.log("getData results:", await getData())
    console.log("readdir:", (await promises.readdir(".")))
}

// (async () => {
//     await on_event(1, 2);
// })();
