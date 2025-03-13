// Simple function to demonstrate the runner
async function greet(name) {
    return `Hello, ${name}!`;
}

// Function to demonstrate error handling
async function divide(a, b) {
    if (b === 0) {
        throw new Error('Cannot divide by zero');
    }
    return a / b;
}

// Function to demonstrate async operations
async function delayedResponse(message, delayMs) {
    await new Promise(resolve => setTimeout(resolve, delayMs));
    return `Delayed message: ${message}`;
}

// Export the functions
module.exports = {
    greet,
    divide,
    delayedResponse
}; 