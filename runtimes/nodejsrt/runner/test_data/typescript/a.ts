import {readdirSync} from "node:fs";

function f(): string[] {
    return readdirSync(".");
}

const results = f();
