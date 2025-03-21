import { getUserInfo } from "./index";

async function test(): Promise<void> {
  console.log("Testing getUserInfo function:");
  const result = await getUserInfo(1);
  console.log("User info:", JSON.stringify(result, null, 2));
}
export { test}
// test();
