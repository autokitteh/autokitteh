// Switch between 2 available modes: user impersonation (using
// a Personal Access Token), and GitHub App (using OAuth v2).

function toggleTab(id) {
  // Update the toggle buttons.
  const buttons = document.getElementsByClassName("toggle");
  for (let i = 0; i < buttons.length; i++) {
    buttons[i].classList.remove("active");
  }
  document.getElementById("toggle" + id).classList.add("active");

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

// Copy the webhook URL into to the clipboard
// when the user clicks "Copy" the button.

const copyButton = document.getElementById("copyButton");
const webhookURL = document.getElementById("webhook");
const notif = document.getElementById("notif");

copyButton.addEventListener("click", () => {
  navigator.clipboard.writeText(webhookURL.value);
  notif.style.display = "inline-block";
  setTimeout(() => {
    notif.style.display = "none";
  }, 1000);
});
