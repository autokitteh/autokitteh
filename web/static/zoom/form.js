// Copy the connection ID and origin query parameters from the URL to
// the form, i.e. pass them through to the connection saving endpoint.
const urlParams = new URLSearchParams(window.location.search);
document.getElementById("cid").value = urlParams.get("cid") ?? "";
document.getElementById("origin").value = urlParams.get("origin") ?? "";

// Show/hide fields based on the selected auth type.
document.getElementById("authType").addEventListener("change", function () {
  const isOauthPrivate = this.value === "oauthPrivate";

  const privateOauthSection = document.getElementById("privateOauthSection");
  if (isOauthPrivate) {
    privateOauthSection.classList.remove("hidden");
  } else {
    privateOauthSection.classList.add("hidden");
  }
  document.getElementById("clientId").disabled = !isOauthPrivate;
  document.getElementById("clientSecret").disabled = !isOauthPrivate;
});
