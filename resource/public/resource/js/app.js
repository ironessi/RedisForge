const state = {
  token: localStorage.getItem("redisforge_token") || "",
  profile: null,
  lastTeamId: localStorage.getItem("redisforge_team_id") || "",
};

const $ = (selector) => document.querySelector(selector);
const logOutput = $("#logOutput");

function setLog(title, payload, redisHint = "等待操作") {
  const time = new Date().toLocaleTimeString();
  const text = typeof payload === "string" ? payload : JSON.stringify(payload, null, 2);
  logOutput.textContent = `[${time}] ${title}\n${text}`;
  $("#flowRedis").textContent = redisHint;
  pushTrail(title, redisHint);
}

function pushTrail(title, detail) {
  const item = document.createElement("div");
  item.className = "event-item";
  item.textContent = `${new Date().toLocaleTimeString()} · ${title} · ${detail}`;
  const trail = $("#eventTrail");
  if (trail.querySelector("p")) {
    trail.innerHTML = "";
  }
  trail.prepend(item);
  [...trail.children].slice(5).forEach((node) => node.remove());
}

function getFormData(form) {
  return Object.fromEntries(new FormData(form).entries());
}

async function request(path, options = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...(options.headers || {}),
  };

  if (state.token) {
    headers.Authorization = `Bearer ${state.token}`;
  }

  const response = await fetch(path, {
    ...options,
    headers,
  });
  const text = await response.text();
  let body = text;

  try {
    body = text ? JSON.parse(text) : {};
  } catch (error) {
    body = { message: text || response.statusText };
  }

  if (!response.ok || body.code) {
    throw body;
  }

  return body;
}

function setToken(token) {
  state.token = token || "";
  if (state.token) {
    localStorage.setItem("redisforge_token", state.token);
  } else {
    localStorage.removeItem("redisforge_token");
  }
  renderSession();
}

function setLastTeamId(teamId) {
  state.lastTeamId = String(teamId || "");
  if (state.lastTeamId) {
    localStorage.setItem("redisforge_team_id", state.lastTeamId);
    $("#memberForm").elements.teamId.value = state.lastTeamId;
    $("#heartbeatForm").elements.teamId.value = state.lastTeamId;
  }
  renderMetrics();
}

function renderMetrics() {
  $("#authMetric").textContent = state.token ? "JWT 已签发" : "未登录";
  $("#cacheMetric").textContent = state.profile ? `user:${state.profile.userId}` : "未加载";
  $("#teamMetric").textContent = state.lastTeamId ? `team:${state.lastTeamId}` : "无团队";
  document.body.classList.toggle("is-online", Boolean(state.token));
}

function renderSession() {
  $("#sessionState").textContent = state.token ? "Online" : "Offline";
  $("#sessionUser").textContent = state.profile
    ? `${state.profile.username} · ID ${state.profile.userId}`
    : state.token
      ? "已保存 token，可以刷新资料。"
      : "登录后会显示当前用户和缓存 key。";
  renderMetrics();
}

function renderProfile(profile) {
  state.profile = profile;
  $("#profileAvatar").textContent = profile?.username?.slice(0, 1).toUpperCase() || "?";
  $("#profileName").textContent = profile ? profile.nickname || profile.username : "未登录";
  $("#profileMeta").textContent = profile
    ? `${profile.username} · user:profile:${profile.userId}`
    : "登录后读取 user:profile:{id}。";
  renderSession();
}

function renderMembers(target, members, emptyText) {
  target.innerHTML = members.length
    ? members.map((member) => `
        <div class="member-item">
          <span>${member.nickname || member.username} · #${member.userId}</span>
          <span class="role-pill">${member.role || "online"}</span>
        </div>
      `).join("")
    : emptyText;
}

async function refreshProfile() {
  const body = await request("/user/profile");
  renderProfile(body.data);
  setLog("GET /user/profile", body, "user:profile:{id}");
}

document.querySelectorAll(".tab").forEach((button) => {
  button.addEventListener("click", () => {
    document.querySelectorAll(".tab").forEach((item) => item.classList.remove("active"));
    document.querySelectorAll("[data-pane]").forEach((item) => item.classList.remove("active"));
    button.classList.add("active");
    document.querySelector(`[data-pane="${button.dataset.tab}"]`).classList.add("active");
  });
});

$("#captchaBtn").addEventListener("click", async () => {
  const username = $("#loginForm").elements.username.value.trim();
  if (!username) {
    $("#captchaHint").textContent = "先填写用户名。";
    return;
  }

  try {
    const body = await request("/auth/captcha", {
      method: "POST",
      body: JSON.stringify({ username }),
    });
    $("#loginForm").elements.captcha.value = body.data.code;
    $("#captchaHint").textContent = `验证码 ${body.data.code} 已写入 Redis，5 分钟内有效。`;
    setLog("POST /auth/captcha", body, "auth:captcha:{username}");
  } catch (error) {
    setLog("验证码失败", error, "String + TTL");
  }
});

$("#loginForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);

  try {
    const body = await request("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
    setToken(body.data.token);
    setLog("POST /auth/login", { ...body, data: { token: `${body.data.token.slice(0, 18)}...` } }, "JWT");
    await refreshProfile();
  } catch (error) {
    setLog("登录失败", error, "captcha 校验");
  }
});

$("#registerForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);

  try {
    const body = await request("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
    setLog("POST /auth/register", body, "MySQL user");
  } catch (error) {
    setLog("注册失败", error, "唯一用户名");
  }
});

$("#logoutBtn").addEventListener("click", async () => {
  if (!state.token) {
    setLog("退出登录", "当前没有 token。", "无 token");
    return;
  }

  try {
    const body = await request("/auth/logout", { method: "POST" });
    setLog("POST /auth/logout", body, "jwt:blacklist:{token}");
  } catch (error) {
    setLog("退出登录结果", error, "jwt:blacklist");
  } finally {
    setToken("");
    renderProfile(null);
  }
});

$("#refreshProfileBtn").addEventListener("click", async () => {
  try {
    await refreshProfile();
  } catch (error) {
    setLog("刷新资料失败", error, "user:profile:{id}");
  }
});

$("#profileForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);

  try {
    const body = await request("/user/profile", {
      method: "PUT",
      body: JSON.stringify(data),
    });
    setLog("PUT /user/profile", body, "删除 user:profile:{id}");
    await refreshProfile();
  } catch (error) {
    setLog("更新资料失败", error, "Cache Aside");
  }
});

$("#teamForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);

  try {
    const body = await request("/teams", {
      method: "POST",
      body: JSON.stringify(data),
    });
    setLastTeamId(body.data.teamId);
    setLog("POST /teams", body, "team:members:{teamId}");
    await loadMembers(body.data.teamId);
  } catch (error) {
    setLog("创建团队失败", error, "team + team_member");
  }
});

$("#memberForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);

  try {
    const body = await request(`/teams/${data.teamId}/members`, {
      method: "POST",
      body: JSON.stringify({ userId: Number(data.userId) }),
    });
    setLastTeamId(data.teamId);
    setLog(`POST /teams/${data.teamId}/members`, body, "SADD team:members:{teamId}");
    await loadMembers(data.teamId);
  } catch (error) {
    setLog("添加成员失败", error, "Set 去重");
  }
});

async function loadMembers(teamId) {
  if (!teamId) {
    setLog("查询成员", "先填写团队 ID。", "team:members");
    return;
  }

  const body = await request(`/teams/${teamId}/members`);
  const members = body.data.members || [];
  renderMembers($("#memberList"), members, "这个团队还没有成员。");
  setLastTeamId(teamId);
  setLog(`GET /teams/${teamId}/members`, body, "SMEMBERS team:members:{teamId}");
}

$("#loadMembersBtn").addEventListener("click", async () => {
  const teamId = $("#memberForm").elements.teamId.value.trim();

  try {
    await loadMembers(teamId);
  } catch (error) {
    setLog("查询成员失败", error, "Set 缓存");
  }
});

$("#heartbeatForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);

  try {
    const body = await request("/presence/heartbeat", {
      method: "POST",
      body: JSON.stringify({ teamId: Number(data.teamId) }),
    });
    setLastTeamId(data.teamId);
    $("#presenceMetric").textContent = "60 秒内在线";
    setLog("POST /presence/heartbeat", body, "presence:user:{id} + presence:team:{id}");
    await loadOnlineMembers(data.teamId);
  } catch (error) {
    setLog("心跳失败", error, "在线状态");
  }
});

async function loadOnlineMembers(teamId) {
  if (!teamId) {
    setLog("在线成员", "先填写团队 ID。", "presence:team");
    return;
  }

  const body = await request(`/teams/${teamId}/online-members`);
  const members = body.data.members || [];
  renderMembers($("#onlineList"), members, "当前没有在线成员。");
  $("#presenceMetric").textContent = `${members.length} 人在线`;
  setLog(`GET /teams/${teamId}/online-members`, body, "EXISTS presence:user:{id}");
}

$("#loadOnlineBtn").addEventListener("click", async () => {
  const teamId = $("#heartbeatForm").elements.teamId.value.trim() || state.lastTeamId;

  try {
    await loadOnlineMembers(teamId);
  } catch (error) {
    setLog("查询在线成员失败", error, "惰性清理 SREM");
  }
});

$("#clearLogBtn").addEventListener("click", () => {
  logOutput.textContent = "等待操作...";
});

renderSession();
if (state.lastTeamId) {
  $("#memberForm").elements.teamId.value = state.lastTeamId;
  $("#heartbeatForm").elements.teamId.value = state.lastTeamId;
}
if (state.token) {
  refreshProfile().catch((error) => setLog("自动刷新资料失败", error, "token 可能过期"));
}
