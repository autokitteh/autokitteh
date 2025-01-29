// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Hide/show the OAuth 2.0 private app fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isDefaultApp = this.value === "oauthDefault";
  const privateApp = document.getElementById("privateApp");
  if (isDefaultApp) {
    privateApp.classList.add("hidden");
  } else {
    privateApp.classList.remove("hidden");
  }
  document.getElementById("clientId").disabled = isDefaultApp;
  document.getElementById("clientSecret").disabled = isDefaultApp;
  document.getElementById("tenantId").disabled = isDefaultApp;

  const submitButton = document.getElementById("submit");
  if (this.value === "daemonApp") {
    submitButton.textContent = "Save Connection";
  } else {
    submitButton.textContent = "Start OAuth Flow";
  }
});
