import { getUserInfo, greet } from "./index";

async function test(): Promise<void> {
  try {
    console.log("Testing greet function:");
    const greeting = await greet("Alice");
    console.log(greeting);

    console.log("\nTesting getUserInfo function:");
    const result = await getUserInfo(1);
    console.log("User info:", JSON.stringify(result, null, 2));
  } catch (error) {
    if (error instanceof Error) {
      console.error("Error:", error.message);
    } else {
      console.error("Unknown error:", error);
    }
  }
}

test(); 