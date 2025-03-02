// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Show/hide fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isDefaultApp = this.value === "oauthDefault";
  const isOauthPrivate = this.value === "oauthPrivate";

  const privateOauthSection = document.getElementById("privateOauthSection");
  if (isOauthPrivate) {
    privateOauthSection.classList.remove("hidden");
  } else {
    privateOauthSection.classList.add("hidden");
  }
  document.getElementById("clientId").disabled = !isOauthPrivate;
  document.getElementById("clientSecret").disabled = !isOauthPrivate;
  document.getElementById("signingSecret").disabled = !isOauthPrivate;

  const privateSocketModeSection = document.getElementById(
    "privateSocketModeSection"
  );
  if (isDefaultApp || isOauthPrivate) {
    privateSocketModeSection.classList.add("hidden");
  } else {
    privateSocketModeSection.classList.remove("hidden");
  }
  document.getElementById("botToken").disabled = isDefaultApp || isOauthPrivate;
  document.getElementById("appToken").disabled = isDefaultApp || isOauthPrivate;

  const submitButton = document.getElementById("submit");
  if (isDefaultApp || isOauthPrivate) {
    submitButton.textContent = "Start OAuth Flow";
  } else {
    submitButton.textContent = "Save Connection";
  }
});
