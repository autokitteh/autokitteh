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
