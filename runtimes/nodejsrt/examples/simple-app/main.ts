import { getUserInfo } from "./index";

async function on_event(args: any): Promise<void>{
    console.log("on_event", args);
    const result = await getUserInfo(1);
    console.log("User info:", JSON.stringify(result, null, 2));


}
export { on_event }
