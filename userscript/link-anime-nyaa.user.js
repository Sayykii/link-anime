// ==UserScript==
// @name         Link Anime - Add to qBittorrent
// @namespace    link-anime
// @version      1.1
// @description  Add torrents from nyaa.si directly to link-anime's qBittorrent
// @author       link-anime
// @match        https://nyaa.si/*
// @match        https://nyaa.land/*
// @match        https://www.nyaa.si/*
// @grant        GM_xmlhttpRequest
// @grant        GM_getValue
// @grant        GM_setValue
// @grant        GM_registerMenuCommand
// @connect      *
// ==/UserScript==

(function () {
  "use strict";

  const DEFAULT_URL = "http://localhost:8787";

  function getServerUrl() {
    return GM_getValue("serverUrl", DEFAULT_URL).replace(/\/+$/, "");
  }

  function getApiKey() {
    return GM_getValue("apiKey", "");
  }

  GM_registerMenuCommand("Set link-anime server URL", () => {
    const current = getServerUrl();
    const url = prompt("Enter your link-anime server URL:", current);
    if (url !== null) {
      GM_setValue("serverUrl", url.replace(/\/+$/, ""));
      alert("Server URL saved: " + url);
    }
  });

  GM_registerMenuCommand("Set link-anime API key", () => {
    const current = getApiKey();
    const key = prompt(
      "Enter your API key (generate one in link-anime Settings):",
      current
    );
    if (key !== null) {
      GM_setValue("apiKey", key.trim());
      alert(key.trim() ? "API key saved" : "API key cleared");
    }
  });

  function addTorrent(magnet, button) {
    const apiKey = getApiKey();
    if (!apiKey) {
      alert(
        "No API key configured.\n\n" +
          "1. Go to your link-anime Settings page\n" +
          "2. Generate an API key\n" +
          "3. Right-click Tampermonkey icon → Link Anime → Set link-anime API key"
      );
      return;
    }

    const orig = button.textContent;
    button.textContent = "Adding...";
    button.disabled = true;
    button.style.opacity = "0.6";

    GM_xmlhttpRequest({
      method: "POST",
      url: getServerUrl() + "/api/qbit/add",
      headers: {
        "Content-Type": "application/json",
        "X-API-Key": apiKey,
      },
      data: JSON.stringify({ magnet }),
      onload(res) {
        if (res.status === 200) {
          button.textContent = "Added!";
          button.style.background = "#2ecc40";
          button.style.borderColor = "#2ecc40";
        } else if (res.status === 401) {
          button.textContent = "Bad API key";
          button.style.background = "#ff851b";
          button.style.borderColor = "#ff851b";
          setTimeout(() => resetButton(button, orig), 3000);
        } else {
          let msg = "Error";
          try {
            msg = JSON.parse(res.responseText).error || msg;
          } catch {}
          button.textContent = msg;
          button.style.background = "#ff4136";
          button.style.borderColor = "#ff4136";
          setTimeout(() => resetButton(button, orig), 3000);
        }
      },
      onerror() {
        button.textContent = "Failed";
        button.style.background = "#ff4136";
        button.style.borderColor = "#ff4136";
        setTimeout(() => resetButton(button, orig), 3000);
      },
    });
  }

  function resetButton(btn, text) {
    btn.textContent = text;
    btn.disabled = false;
    btn.style.opacity = "";
    btn.style.background = "#6366f1";
    btn.style.borderColor = "#6366f1";
  }

  const BUTTON_STYLE = `
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    font-size: 12px;
    font-weight: 600;
    color: #fff;
    background: #6366f1;
    border: 1px solid #6366f1;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s ease;
    white-space: nowrap;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  `;

  function createButton(magnet) {
    const btn = document.createElement("button");
    btn.textContent = "LA";
    btn.title = "Add to link-anime";
    btn.style.cssText = BUTTON_STYLE;
    btn.addEventListener("mouseenter", () => {
      if (!btn.disabled) btn.style.background = "#4f46e5";
    });
    btn.addEventListener("mouseleave", () => {
      if (!btn.disabled) {
        btn.style.background = "#6366f1";
        btn.style.borderColor = "#6366f1";
      }
    });
    btn.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      addTorrent(magnet, btn);
    });
    return btn;
  }

  // --- Nyaa search results page (table rows) ---
  function injectTableButtons() {
    const rows = document.querySelectorAll(".torrent-list tbody tr");
    for (const row of rows) {
      const links = row.querySelectorAll("td:nth-child(3) a");
      let magnet = null;
      for (const a of links) {
        if (a.href && a.href.startsWith("magnet:")) {
          magnet = a.href;
          break;
        }
      }
      if (!magnet) continue;

      const cell = row.querySelector("td:nth-child(3)");
      if (!cell || cell.querySelector(".la-btn")) continue;

      const btn = createButton(magnet);
      btn.classList.add("la-btn");
      btn.style.marginLeft = "4px";
      cell.appendChild(btn);
    }
  }

  // --- Nyaa torrent detail page ---
  function injectDetailButton() {
    const magnetLink = document.querySelector('a[href^="magnet:"]');
    if (!magnetLink) return;

    const panel =
      magnetLink.closest(".panel-footer") || magnetLink.parentElement;
    if (!panel || panel.querySelector(".la-btn")) return;

    const btn = createButton(magnetLink.href);
    btn.classList.add("la-btn");
    btn.style.marginLeft = "8px";
    btn.style.padding = "4px 12px";
    btn.style.fontSize = "13px";
    btn.textContent = "Add to link-anime";
    magnetLink.parentElement.insertBefore(btn, magnetLink.nextSibling);
  }

  // Run on page load
  if (document.querySelector(".torrent-list")) {
    injectTableButtons();
  }
  injectDetailButton();
})();
