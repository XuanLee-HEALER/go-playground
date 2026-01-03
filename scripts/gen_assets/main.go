package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root := "."

	// Keep generated assets defined in one place to avoid manual edits.
	files := map[string]string{
		"webapp/index.html": indexHTML,
		"webapp/styles.css": stylesCSS,
		"webapp/app.js":     appJS,
		"webapp/nginx.conf": nginxConf,
		"docker/web.Dockerfile": webDockerfile,
		"docker/docker-compose.yml": dockerCompose,
		"scripts/web-up.ps1":   webUpPS1,
		"scripts/web-down.ps1": webDownPS1,
		"scripts/web-restart.ps1": webRestartPS1,
		"scripts/web-up.fish":     webUpFish,
		"scripts/web-down.fish":   webDownFish,
		"scripts/web-restart.fish": webRestartFish,
		"scripts/e2e-up.fish":     e2eUpFish,
		"scripts/e2e-down.fish":   e2eDownFish,
	}

	for path, content := range files {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		// Ensure parent directories exist before writing each file.
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			fail("create dir", path, err)
		}
		// Overwrite content so repeated runs stay deterministic.
		if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
			fail("write file", path, err)
		}
		fmt.Printf("wrote %s\n", path)
	}
}

func fail(action, path string, err error) {
	fmt.Fprintf(os.Stderr, "%s %s: %v\n", action, path, err)
	os.Exit(1)
}

const indexHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Go Playground WebSocket Lab</title>
    <link rel="stylesheet" href="styles.css">
  </head>
  <body>
    <div class="glow"></div>
    <div class="page">
      <header class="hero">
        <div class="title-block">
          <p class="eyebrow">Go Playground</p>
          <h1>WebSocket Test Rig</h1>
          <p class="subhead">Full-window session desk with group-scoped echo.</p>
        </div>
        <div class="status-card" aria-live="polite">
          <div class="status-text">
            <p class="label">Sessions</p>
            <p class="value" id="sessionCount">1 active rig</p>
          </div>
          <div class="status-metrics">
            <div>
              <p class="label">Tip</p>
              <p class="value">Same id = shared echo</p>
            </div>
          </div>
        </div>
      </header>

      <section class="panel">
        <div class="panel-header">
          <div>
            <p class="label">Session Deck</p>
            <p class="value">Broadcast stays inside a group, ids stay per user.</p>
          </div>
          <button id="addSessionBtn" class="btn primary">Add Session</button>
        </div>
        <div class="row">
          <label class="field grow">
            <span>Notify User Message</span>
            <input id="notifyMessage" type="text" value="server notice">
          </label>
          <label class="field">
            <span>User Id</span>
            <input id="notifyUserId" type="text" value="alpha">
          </label>
          <div class="button-column">
            <button id="notifyUserBtn" class="btn ghost">Notify User</button>
          </div>
        </div>
        <div class="row">
          <label class="field">
            <span>Redis User Ids (comma)</span>
            <input id="redisUserIds" type="text" value="alpha,beta">
          </label>
          <label class="field">
            <span>Redis Message</span>
            <input id="redisMessage" type="text" value="redis ping">
          </label>
          <label class="field">
            <span>Interval (ms)</span>
            <input id="redisInterval" type="number" min="100" step="100" value="1000">
          </label>
          <div class="button-column">
            <button id="redisStartBtn" class="btn accent">Start Redis</button>
            <button id="redisStopBtn" class="btn ghost" disabled>Stop Redis</button>
          </div>
        </div>
      </section>

      <section class="sessions" id="sessions" aria-live="polite"></section>
    </div>

    <script src="app.js"></script>
  </body>
</html>
`

const stylesCSS = `:root {
  --bg-deep: #0d0f16;
  --bg-mid: #1c2033;
  --bg-light: #2b3150;
  --accent: #ff7a18;
  --accent-2: #19d1ff;
  --text: #f6f7fb;
  --muted: #aab1c4;
  --panel: rgba(20, 24, 38, 0.92);
  --border: rgba(255, 255, 255, 0.08);
  --shadow: 0 30px 60px rgba(0, 0, 0, 0.35);
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  min-height: 100vh;
  font-family: "Space Grotesk", "Trebuchet MS", Verdana, sans-serif;
  color: var(--text);
  background: radial-gradient(circle at top, #353b63 0%, var(--bg-deep) 55%), var(--bg-deep);
  overflow-x: hidden;
}

.glow {
  position: fixed;
  inset: 0;
  background: radial-gradient(circle at 15% 20%, rgba(255, 122, 24, 0.18), transparent 45%),
    radial-gradient(circle at 80% 10%, rgba(25, 209, 255, 0.18), transparent 40%);
  pointer-events: none;
  z-index: 0;
}

.page {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  margin: 0;
  padding: 32px 32px 40px;
  display: flex;
  flex-direction: column;
  gap: 24px;
  animation: fadeIn 0.8s ease-out;
}

.hero {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  align-items: center;
  justify-content: space-between;
}

.eyebrow {
  letter-spacing: 0.3em;
  text-transform: uppercase;
  font-size: 12px;
  margin: 0 0 8px;
  color: var(--muted);
}

h1 {
  margin: 0 0 12px;
  font-size: clamp(32px, 4vw, 48px);
}

.subhead {
  margin: 0;
  color: var(--muted);
  max-width: 420px;
}

.status-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 18px 22px;
  border-radius: 20px;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.06), rgba(255, 255, 255, 0.02));
  border: 1px solid var(--border);
  box-shadow: var(--shadow);
  min-width: 280px;
}

.status-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: #ff3d3d;
  box-shadow: 0 0 12px rgba(255, 61, 61, 0.7);
}

.status-text .label,
.status-metrics .label {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.2em;
}

.status-text .value,
.status-metrics .value {
  margin: 4px 0 0;
  font-size: 16px;
  font-weight: 600;
}

.status-metrics {
  display: flex;
  gap: 18px;
}

.panel {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 24px;
  padding: 24px;
  box-shadow: var(--shadow);
  backdrop-filter: blur(12px);
  animation: liftIn 0.9s ease-out;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: stretch;
  margin-bottom: 18px;
}

.row:last-child {
  margin-bottom: 0;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1 1 300px;
}

.field span {
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: var(--muted);
}

input,
textarea,
select {
  background: rgba(8, 10, 18, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 14px;
  color: var(--text);
  padding: 12px 14px;
  font-size: 14px;
  font-family: "Space Grotesk", "Trebuchet MS", Verdana, sans-serif;
  outline: none;
  transition: border 0.2s ease, box-shadow 0.2s ease;
}

input:focus,
textarea:focus {
  border-color: rgba(255, 122, 24, 0.7);
  box-shadow: 0 0 0 2px rgba(255, 122, 24, 0.2);
}

.button-row,
.button-column {
  display: flex;
  gap: 12px;
}

.button-column {
  flex-direction: column;
}

.btn {
  border: none;
  border-radius: 999px;
  padding: 12px 18px;
  font-weight: 600;
  cursor: pointer;
  color: var(--text);
  background: rgba(255, 255, 255, 0.08);
  transition: transform 0.2s ease, box-shadow 0.2s ease, opacity 0.2s ease;
}

.btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.25);
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
  transform: none;
  box-shadow: none;
}

.btn.primary {
  background: linear-gradient(120deg, var(--accent), #ffb347);
  color: #1a0f02;
}

.btn.accent {
  background: linear-gradient(120deg, #19d1ff, #6af3ff);
  color: #05222b;
}

.btn.ghost {
  border: 1px solid rgba(255, 255, 255, 0.12);
}

.preset-bar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.chip {
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: rgba(255, 255, 255, 0.04);
  color: var(--text);
  padding: 6px 14px;
  font-size: 12px;
  cursor: pointer;
  transition: background 0.2s ease;
}

.chip:hover {
  background: rgba(255, 122, 24, 0.2);
}

.log-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.log-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.log-header h2 {
  margin: 0;
}

.legend {
  display: flex;
  gap: 12px;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.2em;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.legend-item::before {
  content: "";
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.legend-item.sent::before {
  background: var(--accent);
}

.legend-item.received::before {
  background: var(--accent-2);
}

.legend-item.system::before {
  background: #8d9db6;
}

.sessions {
  display: flex;
  gap: 20px;
  align-items: stretch;
  overflow-x: auto;
  padding-bottom: 8px;
  flex: 1 1 auto;
}

.session-card {
  min-width: 360px;
  max-width: 420px;
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 24px;
  padding: 20px;
  box-shadow: var(--shadow);
  animation: liftIn 0.9s ease-out;
}

.session-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.session-title {
  margin: 0;
  font-size: 18px;
}

.session-status {
  display: flex;
  align-items: center;
  gap: 10px;
}

.session-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #ff3d3d;
  box-shadow: 0 0 10px rgba(255, 61, 61, 0.7);
}

.session-metrics {
  display: flex;
  gap: 16px;
}

.log {
  min-height: 200px;
  max-height: 320px;
  overflow-y: auto;
  padding: 14px;
  border-radius: 16px;
  background: rgba(6, 8, 14, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.log-entry {
  padding: 10px 14px;
  border-radius: 14px;
  font-size: 13px;
  line-height: 1.4;
  background: rgba(255, 255, 255, 0.04);
  border-left: 4px solid transparent;
  white-space: pre-wrap;
  word-break: break-word;
}

.log-entry.sent {
  border-left-color: var(--accent);
}

.log-entry.received {
  border-left-color: var(--accent-2);
}

.log-entry.system {
  border-left-color: #8d9db6;
}

.log-entry .meta {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  color: var(--muted);
  margin-bottom: 6px;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes liftIn {
  from {
    opacity: 0;
    transform: translateY(16px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 860px) {
  .status-card {
    width: 100%;
    justify-content: space-between;
  }

  .button-row {
    width: 100%;
  }

  .button-row .btn {
    flex: 1;
  }

  .sessions {
    flex-direction: column;
    overflow-x: visible;
  }

  .session-card {
    min-width: auto;
    max-width: none;
  }
}
`

const appJS = `(() => {
  const sessionsEl = document.getElementById('sessions');
  const addSessionBtn = document.getElementById('addSessionBtn');
  const sessionCountEl = document.getElementById('sessionCount');
  const notifyMessageInput = document.getElementById('notifyMessage');
  const notifyUserInput = document.getElementById('notifyUserId');
  const notifyUserBtn = document.getElementById('notifyUserBtn');
  const redisUserIdsInput = document.getElementById('redisUserIds');
  const redisMessageInput = document.getElementById('redisMessage');
  const redisIntervalInput = document.getElementById('redisInterval');
  const redisStartBtn = document.getElementById('redisStartBtn');
  const redisStopBtn = document.getElementById('redisStopBtn');

  const presets = [
    { label: 'Ping', value: '{ "type": "ping", "value": "ping" }' },
    { label: 'Broadcast', value: '{ "type": "broadcast", "value": "all hands" }' },
    { label: 'Plain Text', value: 'plain text payload' },
  ];

  // Fixed ports to simulate multiple backend nodes locally.
  const nodePorts = [8080, 8081];
  // Persist session configuration across reloads using localStorage.
  const storageKey = 'wsSessions';
  const controlKey = 'wsControls';

  let sessionIndex = 1;
  let redisTimer = null;

  function updateSessionCount() {
    const count = sessionsEl.children.length;
    sessionCountEl.textContent = count + ' active rig' + (count === 1 ? '' : 's');
  }

  function loadSessions() {
    try {
      const raw = localStorage.getItem(storageKey);
      if (!raw) {
        return [];
      }
      const parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch (err) {
      return [];
    }
  }

  function saveSessions() {
    const sessions = Array.from(sessionsEl.children).map((card) => ({
      host: card.querySelector('.ws-host').value,
      port: card.querySelector('.node-port').value,
      userId: card.querySelector('.session-id').value,
      group: card.querySelector('.group-id').value,
      message: card.querySelector('.message-input').value,
    }));
    localStorage.setItem(storageKey, JSON.stringify(sessions));
  }

  function loadControls() {
    try {
      const raw = localStorage.getItem(controlKey);
      if (!raw) {
        return null;
      }
      return JSON.parse(raw);
    } catch (err) {
      return null;
    }
  }

  function saveControls() {
    const controls = {
      notifyMessage: notifyMessageInput.value,
      notifyUserId: notifyUserInput.value,
      redisUserIds: redisUserIdsInput.value,
      redisMessage: redisMessageInput.value,
      redisInterval: redisIntervalInput.value,
    };
    localStorage.setItem(controlKey, JSON.stringify(controls));
  }

  // Log a message entry with styling for sent, received, or system events.
  function logEntry(logEl, kind, title, body) {
    const entry = document.createElement('div');
    entry.className = 'log-entry ' + kind;

    const meta = document.createElement('div');
    meta.className = 'meta';
    meta.textContent = title;

    const content = document.createElement('div');
    content.textContent = body;

    entry.appendChild(meta);
    entry.appendChild(content);
    logEl.appendChild(entry);
    logEl.scrollTop = logEl.scrollHeight;
  }

  // Build a new session card with independent websocket state.
  function createSessionCard(sessionData) {
    const card = document.createElement('article');
    card.className = 'session-card';

    card.innerHTML =
      '<div class="session-header">' +
      '<h2 class="session-title">Session ' + sessionIndex + '</h2>' +
      '<div class="session-status">' +
      '<span class="session-dot"></span>' +
      '<span class="label status-text">Disconnected</span>' +
      '</div>' +
      '</div>' +
      '<div class="session-metrics">' +
      '<div>' +
      '<p class="label">Sent</p>' +
      '<p class="value sent-count">0</p>' +
      '</div>' +
      '<div>' +
      '<p class="label">Received</p>' +
      '<p class="value recv-count">0</p>' +
      '</div>' +
      '</div>' +
      '<div class="row">' +
      '<label class="field">' +
      '<span>Host</span>' +
      '<input class="ws-host" type="text" value="localhost">' +
      '</label>' +
      '<label class="field">' +
      '<span>Node</span>' +
      '<select class="node-port">' +
      nodePorts
        .map((port, idx) => '<option value="' + port + '">Node ' + (idx + 1) + ' : ' + port + '</option>')
        .join('') +
      '</select>' +
      '</label>' +
      '</div>' +
      '<div class="row">' +
      '<label class="field">' +
      '<span>User Id</span>' +
      '<input class="session-id" type="text" value="alpha">' +
      '</label>' +
      '<label class="field">' +
      '<span>Group</span>' +
      '<input class="group-id" type="text" value="alpha-team">' +
      '</label>' +
      '</div>' +
      '<div class="row">' +
      '<div class="button-row">' +
      '<button class="btn primary connect-btn">Connect</button>' +
      '<button class="btn ghost disconnect-btn" disabled>Disconnect</button>' +
      '</div>' +
      '</div>' +
      '<div class="row">' +
      '<label class="field grow">' +
      '<span>Message Payload</span>' +
      '<textarea class="message-input" rows="4" spellcheck="false">{ "type": "echo", "value": "hello" }</textarea>' +
      '</label>' +
      '<div class="button-column">' +
      '<button class="btn accent send-btn" disabled>Send</button>' +
      '<button class="btn ghost format-btn">Format JSON</button>' +
      '<button class="btn ghost clear-btn">Clear Log</button>' +
      '</div>' +
      '</div>' +
      '<div class="preset-bar"></div>' +
      '<div class="log-header">' +
      '<h3 class="session-title">Traffic Log</h3>' +
      '<div class="legend">' +
      '<span class="legend-item sent">Sent</span>' +
      '<span class="legend-item received">Received</span>' +
      '<span class="legend-item system">System</span>' +
      '</div>' +
      '</div>' +
      '<div class="log" role="log" aria-live="polite"></div>';

    sessionIndex += 1;
    wireSession(card, sessionData || {});
    return card;
  }

  function withQuery(url, id, group) {
    const params = [];
    if (id) {
      params.push('id=' + encodeURIComponent(id));
    }
    if (group) {
      params.push('group=' + encodeURIComponent(group));
    }
    if (!params.length) {
      return url;
    }
    return url + (url.includes('?') ? '&' : '?') + params.join('&');
  }

  function buildWsUrl(host, port) {
    return 'ws://' + host + ':' + port + '/ws';
  }

  // Attach websocket behavior to a session card.
  function wireSession(card, sessionData) {
    const hostInput = card.querySelector('.ws-host');
    const nodeSelect = card.querySelector('.node-port');
    const idInput = card.querySelector('.session-id');
    const groupInput = card.querySelector('.group-id');
    const connectBtn = card.querySelector('.connect-btn');
    const disconnectBtn = card.querySelector('.disconnect-btn');
    const sendBtn = card.querySelector('.send-btn');
    const formatBtn = card.querySelector('.format-btn');
    const clearBtn = card.querySelector('.clear-btn');
    const messageInput = card.querySelector('.message-input');
    const logEl = card.querySelector('.log');
    const statusText = card.querySelector('.status-text');
    const statusDot = card.querySelector('.session-dot');
    const sentCountEl = card.querySelector('.sent-count');
    const recvCountEl = card.querySelector('.recv-count');
    const presetBar = card.querySelector('.preset-bar');

    let socket = null;
    let sentCount = 0;
    let recvCount = 0;

    presets.forEach((preset) => {
      const btn = document.createElement('button');
      btn.className = 'chip';
      btn.textContent = preset.label;
      btn.addEventListener('click', () => {
        messageInput.value = preset.value;
      });
      presetBar.appendChild(btn);
    });

    const portOptions = nodePorts.map((port) => String(port));
    if (sessionData.host) {
      hostInput.value = sessionData.host;
    }
    if (sessionData.port && portOptions.includes(String(sessionData.port))) {
      nodeSelect.value = String(sessionData.port);
    }
    if (sessionData.userId) {
      idInput.value = sessionData.userId;
    }
    if (sessionData.group) {
      groupInput.value = sessionData.group;
    }
    if (sessionData.message) {
      messageInput.value = sessionData.message;
    }
    [hostInput, nodeSelect, idInput, groupInput, messageInput].forEach((el) => {
      el.addEventListener('input', saveSessions);
      el.addEventListener('change', saveSessions);
    });

    function setStatus(connected) {
      statusText.textContent = connected ? 'Connected' : 'Disconnected';
      statusDot.style.background = connected ? '#29ff8f' : '#ff3d3d';
      statusDot.style.boxShadow = connected
        ? '0 0 12px rgba(41, 255, 143, 0.7)'
        : '0 0 12px rgba(255, 61, 61, 0.7)';
      connectBtn.disabled = connected;
      disconnectBtn.disabled = !connected;
      sendBtn.disabled = !connected;
    }

    function connect() {
      const host = hostInput.value.trim();
      if (!host) {
        logEntry(logEl, 'system', 'System', 'Host is empty.');
        return;
      }
      const port = nodeSelect.value;
      const baseUrl = buildWsUrl(host, port);
      const sessionId = idInput.value.trim();
      const groupId = groupInput.value.trim();
      // Keep id for identity, scope echo by group.
      const url = withQuery(baseUrl, sessionId, groupId);

      socket = new WebSocket(url);
      logEntry(logEl, 'system', 'System', 'Connecting to ' + url + ' ...');

      socket.addEventListener('open', () => {
        setStatus(true);
        logEntry(logEl, 'system', 'System', 'Connection established.');
      });

      socket.addEventListener('message', (event) => {
        recvCount += 1;
        recvCountEl.textContent = String(recvCount);
        logEntry(logEl, 'received', 'Received', String(event.data));
      });

      socket.addEventListener('close', () => {
        setStatus(false);
        logEntry(logEl, 'system', 'System', 'Connection closed.');
        socket = null;
      });

      socket.addEventListener('error', () => {
        logEntry(logEl, 'system', 'System', 'Socket error detected.');
      });
    }

    function disconnect() {
      if (socket) {
        socket.close();
      }
    }

    function sendMessage() {
      if (!socket || socket.readyState !== WebSocket.OPEN) {
        logEntry(logEl, 'system', 'System', 'Socket is not connected.');
        return;
      }

      const payload = messageInput.value;
      socket.send(payload);
      sentCount += 1;
      sentCountEl.textContent = String(sentCount);
      logEntry(logEl, 'sent', 'Sent', payload);
    }

    function formatJson() {
      const raw = messageInput.value.trim();
      if (!raw) {
        logEntry(logEl, 'system', 'System', 'Message payload is empty.');
        return;
      }
      try {
        const parsed = JSON.parse(raw);
        messageInput.value = JSON.stringify(parsed, null, 2);
        saveSessions();
      } catch (err) {
        logEntry(logEl, 'system', 'System', 'Invalid JSON payload.');
      }
    }

    function clearLog() {
      logEl.innerHTML = '';
    }

    connectBtn.addEventListener('click', connect);
    disconnectBtn.addEventListener('click', disconnect);
    sendBtn.addEventListener('click', sendMessage);
    formatBtn.addEventListener('click', formatJson);
    clearBtn.addEventListener('click', clearLog);

    setStatus(false);
  }

  function activeNode() {
    const firstSession = sessionsEl.querySelector('.session-card');
    if (!firstSession) {
      return null;
    }
    return {
      host: firstSession.querySelector('.ws-host').value.trim(),
      port: firstSession.querySelector('.node-port').value,
    };
  }

  function notifyUser() {
    const node = activeNode();
    if (!node || !node.host) {
      return;
    }
    const userID = notifyUserInput.value.trim();
    if (!userID) {
      return;
    }
    const message = notifyMessageInput.value.trim() || 'notification';
    const url =
      'http://' +
      node.host +
      ':' +
      node.port +
      '/notify/user?id=' +
      encodeURIComponent(userID) +
      '&message=' +
      encodeURIComponent(message);

    fetch(url, { method: 'POST' }).catch(() => {});
  }

  function publishRedis() {
    const node = activeNode();
    if (!node || !node.host) {
      return;
    }
    const ids = redisUserIdsInput.value.trim();
    if (!ids) {
      return;
    }
    const message = redisMessageInput.value.trim() || 'redis notification';
    const url =
      'http://' +
      node.host +
      ':' +
      node.port +
      '/notify/redis?ids=' +
      encodeURIComponent(ids) +
      '&message=' +
      encodeURIComponent(message);
    fetch(url, { method: 'POST' }).catch(() => {});
  }

  function startRedisInterval() {
    if (redisTimer) {
      return;
    }
    const interval = Number(redisIntervalInput.value) || 1000;
    redisTimer = setInterval(publishRedis, interval);
    redisStartBtn.disabled = true;
    redisStopBtn.disabled = false;
  }

  function stopRedisInterval() {
    if (redisTimer) {
      clearInterval(redisTimer);
      redisTimer = null;
    }
    redisStartBtn.disabled = false;
    redisStopBtn.disabled = true;
  }

  [notifyMessageInput, notifyUserInput, redisUserIdsInput, redisMessageInput, redisIntervalInput].forEach((el) => {
    el.addEventListener('input', saveControls);
    el.addEventListener('change', saveControls);
  });

  notifyUserBtn.addEventListener('click', notifyUser);
  redisStartBtn.addEventListener('click', () => {
    publishRedis();
    startRedisInterval();
  });
  redisStopBtn.addEventListener('click', stopRedisInterval);

  addSessionBtn.addEventListener('click', () => {
    const card = createSessionCard();
    sessionsEl.appendChild(card);
    updateSessionCount();
    saveSessions();
  });

  const storedControls = loadControls();
  if (storedControls) {
    if (storedControls.notifyMessage) {
      notifyMessageInput.value = storedControls.notifyMessage;
    }
    if (storedControls.notifyUserId) {
      notifyUserInput.value = storedControls.notifyUserId;
    }
    if (storedControls.redisUserIds) {
      redisUserIdsInput.value = storedControls.redisUserIds;
    }
    if (storedControls.redisMessage) {
      redisMessageInput.value = storedControls.redisMessage;
    }
    if (storedControls.redisInterval) {
      redisIntervalInput.value = storedControls.redisInterval;
    }
  }

  const storedSessions = loadSessions();
  if (storedSessions.length) {
    storedSessions.forEach((session) => {
      sessionsEl.appendChild(createSessionCard(session));
    });
  } else {
    sessionsEl.appendChild(createSessionCard());
  }
  updateSessionCount();
})();
`

const nginxConf = `server {
  listen 80;
  server_name localhost;
  root /usr/share/nginx/html;
  index index.html;

  location / {
    try_files $uri $uri/ /index.html;
  }
}
`

const webDockerfile = `FROM nginx:1.27-alpine
COPY webapp/nginx.conf /etc/nginx/conf.d/default.conf
COPY webapp/ /usr/share/nginx/html
`

const dockerCompose = `services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
  app:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - REDIS_ADDR=redis:6379
      - REDIS_CHANNEL=ws:broadcast
    depends_on:
      - redis
  web:
    build:
      context: ..
      dockerfile: docker/web.Dockerfile
    ports:
      - "3000:80"
`

const webUpPS1 = `# Build and run the web frontend with redis.
docker compose -f docker/docker-compose.yml up --build -d web redis
`

const webDownPS1 = `# Stop and remove only the web frontend container.
docker compose -f docker/docker-compose.yml stop web
docker compose -f docker/docker-compose.yml rm -f web
`

const webRestartPS1 = `# Recreate the web container and ensure redis is up.
docker compose -f docker/docker-compose.yml up --build -d --force-recreate web redis
docker image prune -f
`

const webUpFish = `#!/usr/bin/env fish
# Build and run the web frontend with redis.
docker compose -f docker/docker-compose.yml up --build -d web redis
`

const webDownFish = `#!/usr/bin/env fish
# Stop and remove only the web frontend container.
docker compose -f docker/docker-compose.yml stop web
docker compose -f docker/docker-compose.yml rm -f web
`

const webRestartFish = `#!/usr/bin/env fish
# Recreate the web container and ensure redis is up.
docker compose -f docker/docker-compose.yml up --build -d --force-recreate web redis
docker image prune -f
`

const e2eUpFish = `#!/usr/bin/env fish
# Build and run the app container for local e2e tests.
docker compose -f docker/docker-compose.yml up --build -d
`

const e2eDownFish = `#!/usr/bin/env fish
# Tear down the app container after tests.
docker compose -f docker/docker-compose.yml down
`
