(function() {
  "use strict";

  let colorHashFiles = function() {
    let elements = window.location.hash.substring(1).split("/");

    elements.forEach(function(element) {
      if (element.length === 0) return;
      if (element.substring(0, 2) !== "f:") return;

      let file = element.substring(2);
      let item = document.querySelectorAll(`[data-name="${file}"]`);
      item.forEach(function(i) { i.classList.add("file-selected") })
    });
  }

  window.addEventListener("hashchange", colorHashFiles);
  window.addEventListener("load", colorHashFiles);

  document.querySelectorAll("[data-name]").forEach(function(element) {
    element.addEventListener("click", function(e) {
      if (!((e.ctrlKey || e.metaKey) && e.shiftKey)) return;

      e.preventDefault();
      e.currentTarget.classList.toggle("file-selected");

      let selected = document.querySelectorAll(".file-selected");
      let selectedNames = [];
      selected.forEach(function(element) {
        selectedNames.push(`f:${element.dataset.name}`);
      });

      if (selectedNames.length > 0) {
        window.location.hash = selectedNames.join("/");
      } else {
        history.replaceState({}, document.title, ".");
      }
    });
  });
})();