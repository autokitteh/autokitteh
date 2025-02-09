import {getData} from "./lib";

import {promises} from "fs"

async function on_event(args: any) {
    console.log("event args: ", args)
    console.log("getData results:", await getData(args))
    console.log("readdir:", (await promises.readdir(".")))
}

