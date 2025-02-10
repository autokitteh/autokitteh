import {getData} from "./lib";

import {promises} from "fs"

async function on_event(args: any) {
    const r = await getData(args)
    console.log("event args: ", args)
    console.log("getData results:", r)
    console.log("readdir:", (await promises.readdir(".")))
    return r
}

