const state = {
  token: localStorage.getItem("redisforge_token") || "",
  profile: null,
};

const $ = (selector) => document.querySelector(selector);
const logOutput = $("#logOutput");

function setLog(title, payload) {
  const time = new Date().toLocaleTimeString();
  const text = typeof payload === "string" ? payload : JSON.stringify(payload, null, 2);
  logOutput.textContent = `[${time}] ${title}\n${text}`;
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

function renderSession() {
  $("#sessionState").textContent = state.token ? "Online" : "Offline";
  $("#sessionUser").textContent = state.profile
    ? `${state.profile.username} · ID ${state.profile.userId}`
    : state.token
      ? "已保存 token，可以刷新资料。"
      : "登录后这里会显示当前用户。";
}

function renderProfile(profile) {
  state.profile = profile;
  $("#profileAvatar").textContent = profile?.username?.slice(0, 1).toUpperCase() || "?";
  $("#profileName").textContent = profile ? profile.nickname || profile.username : "未登录";
  $("#profileMeta").textContent = profile
    ? `${profile.username} · user:profile:${profile.userId}`
    : "登录后读取用户资料缓存。";
  renderSession();
}

async function refreshProfile() {
  const body = await request("/user/profile");
  renderProfile(body.data);
  setLog("GET /user/profile", body);
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
    setLog("POST /auth/captcha", body);
  } catch (error) {
    setLog("验证码失败", error);
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
    setLog("POST /auth/login", { ...body, data: { token: `${body.data.token.slice(0, 18)}...` } });
    await refreshProfile();
  } catch (error) {
    setLog("登录失败", error);
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
    setLog("POST /auth/register", body);
  } catch (error) {
    setLog("注册失败", error);
  }
});

$("#logoutBtn").addEventListener("click", async () => {
  if (!state.token) {
    setLog("退出登录", "当前没有 token。");
    return;
  }

  try {
    const body = await request("/auth/logout", { method: "POST" });
    setLog("POST /auth/logout", body);
  } catch (error) {
    setLog("退出登录结果", error);
  } finally {
    setToken("");
    renderProfile(null);
  }
});

$("#refreshProfileBtn").addEventListener("click", async () => {
  try {
    await refreshProfile();
  } catch (error) {
    setLog("刷新资料失败", error);
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
    setLog("PUT /user/profile", body);
    await refreshProfile();
  } catch (error) {
    setLog("更新资料失败", error);
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
    $("#memberForm").elements.teamId.value = body.data.teamId;
    setLog("POST /teams", body);
  } catch (error) {
    setLog("创建团队失败", error);
  }
});

$("#memberForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);

  try {
    const body = await request(`/team/${data.teamId}/members`, {
      method: "POST",
      body: JSON.stringify({ userId: Number(data.userId) }),
    });
    setLog(`POST /team/${data.teamId}/members`, body);
  } catch (error) {
    setLog("添加成员失败", error);
  }
});

$("#clearLogBtn").addEventListener("click", () => {
  logOutput.textContent = "等待操作...";
});

renderSession();
if (state.token) {
  refreshProfile().catch((error) => setLog("自动刷新资料失败", error));
}
