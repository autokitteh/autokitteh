// Display the error message in the page, not just as a URL parameter.
const err = new URLSearchParams(window.location.search).get("error");
const elem = document.getElementById("error");

if (err) {
  elem.textContent = err;
}

// Allow the user to retry initializing the connection.
document.getElementById("retry").onclick = function () {
  const id = new URLSearchParams(window.location.search).get("cid");
  const origin = new URLSearchParams(window.location.search).get("origin");
  window.location.href = "./?cid=".concat(id, "&origin=", origin);
};
