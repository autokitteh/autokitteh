import {listSnippets} from "./gmail";
import {isInvoice} from "./openai";

(async() => {
    let snippets = await listSnippets()
    await Promise.all(
        snippets.map(async (snippet) => {
            console.log(await isInvoice(snippet), snippet)
        })
    )
})()

async function on_event(args: any) {
    let snippets = await listSnippets()
    await Promise.all(
        snippets.map(async (snippet) => {
            console.log(await isInvoice(snippet), snippet)
        })
    )
}
