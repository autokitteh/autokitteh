// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Hide/show the private OAuth 2.0 and API key fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isPrivateOauth = this.value === "oauthPrivate";
  const isApiKey = this.value === "apiKey";

  const privateOauthSection = document.getElementById("privateOauthSection");
  if (isPrivateOauth) {
    privateOauthSection.classList.remove("hidden");
  } else {
    privateOauthSection.classList.add("hidden");
  }
  document.getElementById("clientId").disabled = !isPrivateOauth;
  document.getElementById("clientSecret").disabled = !isPrivateOauth;
  document.getElementById("webhookSecret").disabled = !isPrivateOauth;

  const apiKeySection = document.getElementById("apiKeySection");
  if (isApiKey) {
    apiKeySection.classList.remove("hidden");
  } else {
    apiKeySection.classList.add("hidden");
  }
  document.getElementById("apiKey").disabled = !isApiKey;

  const submitButton = document.getElementById("submit");
  if (this.value === "apiKey") {
    submitButton.textContent = "Save Connection";
  } else {
    submitButton.textContent = "Start OAuth Flow";
  }
});
