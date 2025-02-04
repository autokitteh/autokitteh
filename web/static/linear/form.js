// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Show/hide fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isApiKey = this.value === "apiKey";
  const isOauthPrivate = this.value === "oauthPrivate";

  const oauthSection = document.getElementById("oauthSection");
  if (isApiKey) {
    oauthSection.classList.add("hidden");
  } else {
    oauthSection.classList.remove("hidden");
  }
  document.getElementById("actor").disabled = isApiKey;

  const privateOauthSection = document.getElementById("privateOauthSection");
  if (isOauthPrivate) {
    privateOauthSection.classList.remove("hidden");
  } else {
    privateOauthSection.classList.add("hidden");
  }
  document.getElementById("clientId").disabled = !isOauthPrivate;
  document.getElementById("clientSecret").disabled = !isOauthPrivate;
  document.getElementById("webhookSecret").disabled = !isOauthPrivate;

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
