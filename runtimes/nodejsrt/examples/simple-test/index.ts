import UserService from "./user-service";

// Simple function to demonstrate the runner
async function greet(name: string): Promise<string> {
  const response = await delayedResponse(`Hello, ${name}!`, 1000);
  return response as string;
}

// Function to demonstrate error handling
async function divide(a: number, b: number): Promise<number> {
  if (b === 0) {
    throw new Error("Cannot divide by zero");
  }
  return a / b;
}

// Function to demonstrate async operations
async function delayedResponse(message: string, delayMs: number): Promise<string> {
  await new Promise(resolve => setTimeout(resolve, delayMs));
  return `Delayed message: ${message}`;
}

interface UserInfo {
  user: {
    name: string;
    email: string;
    company: string;
  };
  posts: Array<{
    title: string;
    body: string;
  }>;
}

// Function using the UserService class
async function getUserInfo(userId: number): Promise<UserInfo> {
  const userService = new UserService();
  const user = await userService.getUserById(userId);
  console.log(user);
  const posts = await userService.getUserPosts(userId);
  console.log(posts);
  return {
    user,
    posts: posts.slice(0, 3) // Get first 3 posts only
  };
}

async function getUserInfo1() {
  const users = await getUserInfo(1);
  console.log(users);
}

// Export the functions
export {
  greet,
  divide,
  delayedResponse,
  getUserInfo,
  UserService,
  getUserInfo1
};
