function showToast(message) {
  // Get the snackbar DIV
  var x = document.getElementById("snackbar");

  // Add the "show" class to DIV
  x.className = "show";
  x.innerHTML = message

  // close toast by clicking
  x.onclick = function () {
    x.className = x.className.replace("show", "");
  };

  // After 3 seconds, remove the show class from DIV
  setTimeout(function () { x.className = x.className.replace("show", ""); }, 7000);
}