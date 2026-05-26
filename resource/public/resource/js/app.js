const state = {
  token: localStorage.getItem("redisforge_token") || "",
  profile: null,
  lastTeamId: localStorage.getItem("redisforge_team_id") || "",
  activityCount: null,
  taskCount: null,
  hotTaskCount: null,
  tasks: [],
  selectedTask: null,
};

const $ = (selector) => document.querySelector(selector);
const logOutput = $("#logOutput");
const taskStatusLabels = {
  todo: "待处理",
  doing: "进行中",
  done: "已完成",
};

function taskStatusLabel(status) {
  return taskStatusLabels[status] || status;
}

function taskPriorityLabel(priority) {
  return { 1: "低", 2: "中", 3: "高" }[priority] || priority;
}

function taskPayload(data) {
  return {
    title: data.title.trim(),
    description: data.description.trim(),
    assigneeId: Number(data.assigneeId || 0),
    priority: Number(data.priority),
  };
}

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
    $("#activityTeamId").value = state.lastTeamId;
    $("#taskCreateForm").elements.teamId.value = state.lastTeamId;
    $("#taskTeamId").value = state.lastTeamId;
    $("#hotTaskTeamId").value = state.lastTeamId;
  }
  renderMetrics();
}

function renderMetrics() {
  $("#authMetric").textContent = state.token ? "JWT 已签发" : "未登录";
  $("#cacheMetric").textContent = state.profile ? `user:${state.profile.userId}` : "未加载";
  $("#teamMetric").textContent = state.lastTeamId ? `team:${state.lastTeamId}` : "无团队";
  $("#activityMetric").textContent = state.activityCount === null ? "未加载" : `${state.activityCount} 条`;
  $("#taskMetric").textContent = state.taskCount === null ? "未加载" : `${state.taskCount} 条`;
  $("#hotMetric").textContent = state.hotTaskCount === null ? "未加载" : `${state.hotTaskCount} 条`;
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
  target.replaceChildren();
  if (!members.length) {
    target.textContent = emptyText;
    return;
  }

  members.forEach((member) => {
    const item = document.createElement("div");
    item.className = "member-item";

    const name = document.createElement("span");
    name.textContent = `${member.nickname || member.username} · #${member.userId}`;

    const role = document.createElement("span");
    role.className = "role-pill";
    role.textContent = member.role || "online";

    item.append(name, role);
    target.append(item);
  });
}

function renderActivities(activities) {
  const target = $("#activityList");
  target.replaceChildren();
  state.activityCount = activities.length;
  renderMetrics();

  if (!activities.length) {
    target.textContent = "暂无动态。";
    return;
  }

  activities.forEach((activity) => {
    const item = document.createElement("div");
    item.className = "activity-item";

    const mark = document.createElement("div");
    mark.className = "activity-mark";
    mark.textContent = {
      "team.created": "+",
      "member.added": "@",
      task_created: "T",
      task_updated: "~",
      task_status_updated: ">",
    }[activity.action] || "@";

    const body = document.createElement("div");
    body.className = "activity-body";
    const title = document.createElement("b");
    title.textContent = activity.content;
    const detail = document.createElement("small");
    if (activity.action === "team.created") {
      detail.textContent = `发起人 #${activity.actorId}`;
    } else if (activity.action === "member.added") {
      detail.textContent = `操作人 #${activity.actorId} · 成员 #${activity.targetUserId}`;
    } else if (activity.action.startsWith("task_")) {
      detail.textContent = `操作人 #${activity.actorId} · 任务动态`;
    } else {
      detail.textContent = `操作人 #${activity.actorId}`;
    }
    body.append(title, detail);

    const time = document.createElement("time");
    time.className = "activity-time";
    time.textContent = new Date(activity.createdAt * 1000).toLocaleString();

    item.append(mark, body, time);
    target.append(item);
  });
}

function selectTask(task) {
  state.selectedTask = task || null;
  const controls = $("#taskDetailControls");
  const status = $("#selectedTaskStatus");

  if (!task) {
    $("#selectedTaskTitle").textContent = "从列表中查看一条任务。";
    status.textContent = "未选择";
    status.className = "task-status neutral";
    controls.hidden = true;
    return;
  }

  $("#selectedTaskTitle").textContent = `#${task.taskId} ${task.title}`;
  status.textContent = taskStatusLabel(task.status);
  status.className = `task-status ${task.status}`;
  $("#selectedTaskMeta").textContent =
    `创建人 #${task.creatorId} · 负责人 ${task.assigneeId ? `#${task.assigneeId}` : "未分配"} · 优先级 ${taskPriorityLabel(task.priority)}`;

  const form = $("#taskEditForm");
  form.elements.title.value = task.title;
  form.elements.description.value = task.description;
  form.elements.assigneeId.value = task.assigneeId || 0;
  form.elements.priority.value = String(task.priority);
  controls.hidden = false;

  document.querySelectorAll(".task-status-action").forEach((button) => {
    button.classList.toggle("current", button.dataset.status === task.status);
  });
}

function renderTasks(tasks) {
  const target = $("#taskList");
  target.replaceChildren();
  state.tasks = tasks;
  state.taskCount = tasks.length;
  renderMetrics();

  if (state.selectedTask) {
    selectTask(tasks.find((task) => task.taskId === state.selectedTask.taskId) || null);
  }

  if (!tasks.length) {
    target.textContent = "这个团队暂时没有任务，可以从左侧创建第一条。";
    return;
  }

  tasks.forEach((task) => {
    const item = document.createElement("div");
    item.className = "task-item";

    const body = document.createElement("div");
    const title = document.createElement("b");
    title.textContent = task.title;
    const detail = document.createElement("small");
    detail.textContent =
      `#${task.taskId} · 负责人 ${task.assigneeId ? `#${task.assigneeId}` : "未分配"} · 优先级 ${taskPriorityLabel(task.priority)}`;
    body.append(title, detail);

    const actions = document.createElement("div");
    actions.className = "task-item-actions";
    const status = document.createElement("span");
    status.className = `task-status ${task.status}`;
    status.textContent = taskStatusLabel(task.status);
    const openButton = document.createElement("button");
    openButton.type = "button";
    openButton.className = "secondary task-open";
    openButton.textContent = "查看/编辑";
    openButton.addEventListener("click", async () => {
      try {
        await openTaskDetail(task.taskId);
      } catch (error) {
        setLog("查看任务详情失败", error, "ZSet 未计分");
      }
    });
    actions.append(status, openButton);

    item.append(body, actions);
    target.append(item);
  });
}

function renderHotTasks(tasks) {
  const target = $("#hotTaskList");
  target.replaceChildren();
  state.hotTaskCount = tasks.length;
  renderMetrics();

  if (!tasks.length) {
    target.textContent = "暂无热门任务。打开一次任务详情即可产生热度数据。";
    return;
  }

  tasks.forEach((task, index) => {
    const item = document.createElement("div");
    item.className = "hot-task-item";

    const rank = document.createElement("div");
    rank.className = "hot-task-rank";
    rank.textContent = `#${index + 1}`;

    const body = document.createElement("div");
    body.className = "hot-task-body";
    const title = document.createElement("b");
    title.textContent = task.title;
    const detail = document.createElement("small");
    detail.textContent = `任务 #${task.taskId} · ${taskStatusLabel(task.status)} · 优先级 ${taskPriorityLabel(task.priority)}`;
    body.append(title, detail);

    const score = document.createElement("span");
    score.className = "hot-task-score";
    score.textContent = `热度 ${task.viewCount}`;

    const openButton = document.createElement("button");
    openButton.type = "button";
    openButton.className = "secondary hot-task-open";
    openButton.textContent = "打开任务";
    openButton.addEventListener("click", async () => {
      try {
        await openTaskDetail(task.taskId);
      } catch (error) {
        setLog("查看任务详情失败", error, "ZSet 未计分");
      }
    });

    item.append(rank, body, score, openButton);
    target.append(item);
  });
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
    await loadActivities(body.data.teamId);
    await loadTasks(body.data.teamId);
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
    await loadActivities(data.teamId);
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

async function loadTasks(teamId, shouldLog = true) {
  if (!teamId) {
    setLog("查询任务", "先填写团队 ID。", "task");
    return [];
  }

  const body = await request(`/teams/${teamId}/tasks`);
  const tasks = body.data.tasks || [];
  renderTasks(tasks);
  setLastTeamId(teamId);
  if (shouldLog) {
    setLog(`GET /teams/${teamId}/tasks`, body, "MySQL task");
  }
  return tasks;
}

async function openTaskDetail(taskId) {
  const body = await request(`/tasks/${taskId}`);
  selectTask(body.data.task);
  if (state.lastTeamId) {
    await loadHotTasks(state.lastTeamId, false);
  }
  setLog(`GET /tasks/${taskId}`, body, "ZINCRBY team:task:hot:{teamId}");
}

$("#loadTasksBtn").addEventListener("click", async () => {
  const teamId = $("#taskTeamId").value.trim() || state.lastTeamId;

  try {
    await loadTasks(teamId);
  } catch (error) {
    setLog("查询任务失败", error, "任务权限校验");
  }
});

$("#taskCreateForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = getFormData(event.currentTarget);
  const teamId = data.teamId.trim();

  try {
    const body = await request(`/teams/${teamId}/tasks`, {
      method: "POST",
      body: JSON.stringify(taskPayload(data)),
    });
    setLastTeamId(teamId);
    const tasks = await loadTasks(teamId, false);
    selectTask(tasks.find((task) => task.taskId === body.data.taskId) || null);
    await Promise.all([loadActivities(teamId, false), loadHotTasks(teamId, false)]);
    event.currentTarget.reset();
    event.currentTarget.elements.teamId.value = teamId;
    event.currentTarget.elements.assigneeId.value = "0";
    event.currentTarget.elements.priority.value = "2";
    setLog(`POST /teams/${teamId}/tasks`, body, "LPUSH team:activities:{teamId}");
  } catch (error) {
    setLog("创建任务失败", error, "任务成员权限");
  }
});

$("#taskEditForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!state.selectedTask) {
    return;
  }

  const data = getFormData(event.currentTarget);
  const taskId = state.selectedTask.taskId;
  const teamId = state.lastTeamId;

  try {
    const body = await request(`/tasks/${taskId}`, {
      method: "PUT",
      body: JSON.stringify(taskPayload(data)),
    });
    await loadTasks(teamId, false);
    await Promise.all([loadActivities(teamId, false), loadHotTasks(teamId, false)]);
    setLog(`PUT /tasks/${taskId}`, body, "LPUSH team:activities:{teamId} + ZINCRBY");
  } catch (error) {
    setLog("更新任务失败", error, "任务权限校验");
  }
});

document.querySelectorAll(".task-status-action").forEach((button) => {
  button.addEventListener("click", async () => {
    if (!state.selectedTask || state.selectedTask.status === button.dataset.status) {
      return;
    }

    const taskId = state.selectedTask.taskId;
    const teamId = state.lastTeamId;

    try {
      const body = await request(`/tasks/${taskId}/status`, {
        method: "PATCH",
        body: JSON.stringify({ status: button.dataset.status }),
      });
      await loadTasks(teamId, false);
      await Promise.all([loadActivities(teamId, false), loadHotTasks(teamId, false)]);
      setLog(`PATCH /tasks/${taskId}/status`, body, "LPUSH team:activities:{teamId} + ZINCRBY");
    } catch (error) {
      setLog("更新任务状态失败", error, "任务权限校验");
    }
  });
});

// Activity 使用 Redis List，创建团队或添加成员后可立刻刷新查看最近事件。
async function loadActivities(teamId, shouldLog = true) {
  if (!teamId) {
    setLog("团队动态", "先填写团队 ID。", "team:activities");
    return;
  }

  const body = await request(`/teams/${teamId}/activities`);
  renderActivities(body.data.activities || []);
  setLastTeamId(teamId);
  if (shouldLog) {
    setLog(`GET /teams/${teamId}/activities`, body, "LRANGE team:activities:{teamId}");
  }
}

$("#loadActivitiesBtn").addEventListener("click", async () => {
  const teamId = $("#activityTeamId").value.trim() || state.lastTeamId;

  try {
    await loadActivities(teamId);
  } catch (error) {
    setLog("查询动态失败", error, "Redis List");
  }
});

// Hot Tasks 使用 Redis Sorted Set，按热度从高到低返回团队前 10 项任务。
async function loadHotTasks(teamId, shouldLog = true) {
  if (!teamId) {
    setLog("热门任务", "先填写团队 ID。", "team:task:hot");
    return;
  }

  const body = await request(`/teams/${teamId}/tasks/hot`);
  renderHotTasks(body.data.tasks || []);
  setLastTeamId(teamId);
  if (shouldLog) {
    setLog(`GET /teams/${teamId}/tasks/hot`, body, "ZREVRANGE team:task:hot:{teamId}");
  }
}

$("#loadHotTasksBtn").addEventListener("click", async () => {
  const teamId = $("#hotTaskTeamId").value.trim() || state.lastTeamId;

  try {
    await loadHotTasks(teamId);
  } catch (error) {
    setLog("查询热门任务失败", error, "Redis Sorted Set");
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
  $("#activityTeamId").value = state.lastTeamId;
  $("#taskCreateForm").elements.teamId.value = state.lastTeamId;
  $("#taskTeamId").value = state.lastTeamId;
  $("#hotTaskTeamId").value = state.lastTeamId;
}
if (state.token) {
  refreshProfile().catch((error) => setLog("自动刷新资料失败", error, "token 可能过期"));
}
