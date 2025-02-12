// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Show/hide fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isDefaultApp = this.value === "oauthDefault";
  const isOauthPrivate = this.value === "oauthPrivate";

  const privateAppSection = document.getElementById("privateAppSection");
  if (isDefaultApp) {
    privateAppSection.classList.add("hidden");
  } else {
    privateAppSection.classList.remove("hidden");
  }
  document.getElementById("clientId").disabled = isDefaultApp;
  document.getElementById("clientSecret").disabled = isDefaultApp;

  const privateS2SSection = document.getElementById("privateS2SSection");
  if (isDefaultApp || isOauthPrivate) {
    privateS2SSection.classList.add("hidden");
  } else {
    privateS2SSection.classList.remove("hidden");
  }
  document.getElementById("accountId").disabled = isDefaultApp || isOauthPrivate;

  const privateOauthSection = document.getElementById("privateOauthSection");
  if (isOauthPrivate) {
    privateOauthSection.classList.remove("hidden");
  } else {
    privateOauthSection.classList.add("hidden");
  }
  document.getElementById("secretToken").disabled = !isOauthPrivate;

  const submitButton = document.getElementById("submit");
  if (isDefaultApp || isOauthPrivate) {
    submitButton.textContent = "Start OAuth Flow";
  } else {
    submitButton.textContent = "Save Connection";
  }
});
