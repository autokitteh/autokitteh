// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Add form validation and user feedback
document.addEventListener("DOMContentLoaded", function() {
  const form = document.querySelector("form");
  const botTokenInput = document.getElementById("botToken");
  const submitButton = document.getElementById("submit");

  // Bot token validation
  function validateBotToken(token) {
    // Telegram bot token format: bot<number>:<alphanumeric_string>
    const tokenRegex = /^\d+:[A-Za-z0-9_-]{35}$/;
    return tokenRegex.test(token);
  }

  // Real-time validation feedback
  botTokenInput.addEventListener("input", function() {
    const token = this.value.trim();
    if (token && !validateBotToken(token)) {
      this.style.borderColor = "#e74c3c";
      this.style.backgroundColor = "#fdf2f2";
    } else {
      this.style.borderColor = "#ccc";
      this.style.backgroundColor = "white";
    }
  });

  // Form submission handling
  form.addEventListener("submit", function(e) {
    const botToken = botTokenInput.value.trim();
    
    if (!botToken) {
      e.preventDefault();
      alert("Bot token is required.");
      botTokenInput.focus();
      return;
    }

    if (!validateBotToken(botToken)) {
      e.preventDefault();
      alert("Invalid bot token format. Please check your token and try again.");
      botTokenInput.focus();
      return;
    }

    // Show loading state
    submitButton.disabled = true;
    submitButton.textContent = "Saving Connection...";
  });
});
