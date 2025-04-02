# Simple Test Project

This is a simple Node.js project used for testing the autokitteh runner. It contains basic functions that demonstrate different aspects of JavaScript/Node.js functionality:

## Functions

### greet(name)
A simple async function that returns a greeting message.
```javascript
await greet("World") // Returns: "Hello, World!"
```

### divide(a, b)
A function that demonstrates error handling.
```javascript
await divide(10, 2) // Returns: 5
await divide(10, 0) // Throws: Error('Cannot divide by zero')
```

### delayedResponse(message, delayMs)
A function that demonstrates async operations with delays.
```javascript
await delayedResponse("Test", 1000) // Returns after 1 second: "Delayed message: Test"
```

## Usage with Runner

This project is designed to be used with the autokitteh runner for testing purposes. The functions can be called through the runner's interface to verify different aspects of the runtime behavior. 