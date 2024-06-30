// Display the error message in the page too, not just as a URL parameter.

const param = new URLSearchParams(window.location.search).get("error");
const elem = document.getElementById("error");

if (param) {
  elem.textContent = param;
}
