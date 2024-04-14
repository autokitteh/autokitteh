// Display the token in the page too, not just as a URL parameter.
// Also allow the user to easily copy it into the clipboard.

const param = new URLSearchParams(window.location.search).get("token");
const token = document.getElementById("token");
const notif = document.getElementById("notif");

if (param) {
  token.textContent = param;
  token.addEventListener("click", () => {
    navigator.clipboard.writeText(param);
    notif.style.display = "block";
    setTimeout(() => {
      notif.style.display = "none";
    }, 2000);
  });
} else {
  token.textContent = "Connection token not found in URL";
}
