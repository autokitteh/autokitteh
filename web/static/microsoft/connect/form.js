// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Hide/show the custom OAuth 2.0 app fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isDefaultApp = this.value === "oauthDefault";
  const customApp = document.getElementById("customApp");
  if (isDefaultApp) {
    customApp.classList.add("hidden");
  } else {
    customApp.classList.remove("hidden");
  }
  document.getElementById("clientId").disabled = isDefaultApp;
  document.getElementById("clientSecret").disabled = isDefaultApp;
});
