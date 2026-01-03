(() => {
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
