// Switch between 2 available modes: user impersonation (using
// OAuth 2.0), and a GCP service account (using a JSON key).

function toggleTab(id) {
  // Update the toggle buttons.
  const buttons = document.getElementsByClassName("toggle");
  for (let i = 0; i < buttons.length; i++) {
    buttons[i].classList.remove("active");
  }
  document.getElementById("toggle" + id).classList.add("active");

  // Synchronize the form ID text fields.
  if (id === "tab1") {
    jsonValue = document.getElementById("formIdJson").value;
    document.getElementById("formIdOauth").value = jsonValue;
    document.getElementById("formIdJson").value = "";
  } else {
    oauthValue = document.getElementById("formIdOauth").value;
    document.getElementById("formIdJson").value = oauthValue;
    document.getElementById("formIdOauth").value = "";
  }

  // Update the tab contents.
  const tabs = document.getElementsByClassName("tab");
  for (let i = 0; i < tabs.length; i++) {
    tabs[i].classList.remove("active");
  }
  document.getElementById(id).classList.add("active");
}

window.onload = function () {
  toggleTab("tab1");
};
