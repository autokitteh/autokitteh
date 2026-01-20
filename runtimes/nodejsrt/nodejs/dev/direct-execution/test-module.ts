// Test module for direct execution
// Note: This simulates code that has already been through the transformation phase
// where awaited calls have been wrapped with ak_call

// The main function that gets called by the runner
async function callExternalService(): Promise<string> {
  console.log("callExternalService called");

  // Using a hardcoded name for testing
  const name = "TestUser";
  console.log("Calling external service for:", name);

  // This is what the code looks like AFTER patching
  // Original: return await externalServiceMock(name);
  // Patched:  return await ak_call(externalServiceMock, name);
  return await (global as any).ak_call(externalServiceMock, name);
}

// This simulates an external service function
async function externalServiceMock(name: string): Promise<string> {
  console.log("External service called with:", name);
  return "Response from external service for: " + name;
}

// Export the function
export {
  callExternalService,
};
