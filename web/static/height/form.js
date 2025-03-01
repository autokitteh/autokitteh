// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Show/hide fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isApiKey = this.value === "apiKey";
  const isOauthPrivate = this.value === "oauthPrivate";

  const privateOauthSection = document.getElementById("privateOauthSection");
  if (isOauthPrivate) {
    privateOauthSection.classList.remove("hidden");
  } else {
    privateOauthSection.classList.add("hidden");
  }
  document.getElementById("clientId").disabled = !isOauthPrivate;
  document.getElementById("clientSecret").disabled = !isOauthPrivate;

  const apiKeySection = document.getElementById("apiKeySection");
  const submitButton = document.getElementById("submit");
  if (isApiKey) {
    apiKeySection.classList.remove("hidden");
    submitButton.textContent = "Save Connection";
  } else {
    apiKeySection.classList.add("hidden");
    submitButton.textContent = "Start OAuth Flow";
  }
  document.getElementById("apiKey").disabled = !isApiKey;
});
