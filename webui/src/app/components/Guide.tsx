import { useState } from "react";
import {
  Activity,
  Grid3x3,
  Plus,
  Users,
  KeyRound,
  Hash,
  Play,
  Pause,
  Trash2,
  Link2,
  Copy,
  Eye,
  RefreshCw,
  ChevronDown,
  ChevronRight,
  BookOpen,
  Monitor,
  Terminal,
  Shield,
  Zap,
  Info,
} from "lucide-react";
import { Badge } from "./ui/badge";
import { Button } from "./ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

interface Section {
  id: string;
  title: string;
  icon: React.ReactNode;
}

const sections: Section[] = [
  { id: "overview", title: "产品概述", icon: <BookOpen className="h-4 w-4" /> },
  { id: "login", title: "登录与认证", icon: <Shield className="h-4 w-4" /> },
  { id: "dashboard", title: "控制台总览", icon: <Monitor className="h-4 w-4" /> },
  { id: "instances", title: "实例管理", icon: <Grid3x3 className="h-4 w-4" /> },
  { id: "allocate", title: "分配实例", icon: <Plus className="h-4 w-4" /> },
  { id: "users", title: "用户管理", icon: <Users className="h-4 w-4" /> },
  { id: "audit", title: "审计日志", icon: <Shield className="h-4 w-4" /> },
  { id: "api", title: "API 访问", icon: <Terminal className="h-4 w-4" /> },
  { id: "workflow", title: "典型使用流程", icon: <Zap className="h-4 w-4" /> },
];

function SectionCard({
  id,
  title,
  children,
}: {
  id: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className="scroll-mt-20">
      <Card className="bg-slate-900 border-slate-800">
        <CardHeader className="pb-3">
          <CardTitle className="text-white text-lg">{title}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 text-slate-300 text-sm leading-relaxed">
          {children}
        </CardContent>
      </Card>
    </section>
  );
}

function Step({
  num,
  title,
  desc,
}: {
  num: number;
  title: string;
  desc: React.ReactNode;
}) {
  return (
    <div className="flex gap-3">
      <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-cyan-600 text-white text-xs font-bold mt-0.5">
        {num}
      </div>
      <div>
        <p className="font-medium text-white">{title}</p>
        <div className="text-slate-400 mt-0.5">{desc}</div>
      </div>
    </div>
  );
}

function Tip({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex gap-2 rounded-md bg-cyan-900/20 border border-cyan-800/30 p-3">
      <Info className="h-4 w-4 text-cyan-400 shrink-0 mt-0.5" />
      <p className="text-cyan-300 text-sm">{children}</p>
    </div>
  );
}

function Warn({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex gap-2 rounded-md bg-amber-900/20 border border-amber-800/30 p-3">
      <Info className="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
      <p className="text-amber-300 text-sm">{children}</p>
    </div>
  );
}

function Code({ children }: { children: string }) {
  return (
    <code className="bg-slate-800 px-1.5 py-0.5 rounded text-cyan-400 text-xs font-mono">
      {children}
    </code>
  );
}

function CodeBlock({ children }: { children: string }) {
  return (
    <pre className="bg-slate-800 rounded-md p-3 text-cyan-400 text-xs font-mono overflow-x-auto whitespace-pre-wrap break-all">
      {children}
    </pre>
  );
}

function IconLabel({
  icon,
  label,
}: {
  icon: React.ReactNode;
  label: string;
}) {
  return (
    <span className="inline-flex items-center gap-1 bg-slate-800 px-2 py-0.5 rounded text-slate-300 text-xs">
      {icon}
      {label}
    </span>
  );
}

export function Guide() {
  const [activeSection, setActiveSection] = useState("overview");
  const [navOpen, setNavOpen] = useState(false);

  const scrollTo = (id: string) => {
    setActiveSection(id);
    setNavOpen(false);
    document.getElementById(id)?.scrollIntoView({ behavior: "smooth" });
  };

  return (
    <div className="min-h-screen bg-slate-950">
      <div className="container mx-auto px-4 py-8 max-w-6xl">
        {/* 标题 */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600">
              <Activity className="h-6 w-6 text-white" />
            </div>
            <div>
              <h1 className="text-2xl font-bold text-white">Claw Swarm 使用说明</h1>
              <p className="text-slate-400 text-sm">Operation Console — 产品使用指南</p>
            </div>
          </div>
        </div>

        <div className="flex gap-6">
          {/* 侧边导航 — 桌面端 */}
          <aside className="hidden lg:block w-52 shrink-0">
            <div className="sticky top-20 space-y-1">
              {sections.map((s) => (
                <button
                  key={s.id}
                  onClick={() => scrollTo(s.id)}
                  className={`w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors text-left ${
                    activeSection === s.id
                      ? "bg-cyan-600/20 text-cyan-400"
                      : "text-slate-400 hover:text-white hover:bg-slate-800/50"
                  }`}
                >
                  {s.icon}
                  {s.title}
                </button>
              ))}
            </div>
          </aside>

          {/* 移动端导航折叠 */}
          <div className="lg:hidden w-full mb-4">
            <button
              className="flex items-center gap-2 text-slate-300 text-sm border border-slate-700 rounded-md px-3 py-2"
              onClick={() => setNavOpen((v) => !v)}
            >
              {navOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
              目录导航
            </button>
            {navOpen && (
              <div className="mt-2 space-y-1 border border-slate-800 rounded-md p-2 bg-slate-900">
                {sections.map((s) => (
                  <button
                    key={s.id}
                    onClick={() => scrollTo(s.id)}
                    className="w-full flex items-center gap-2 px-3 py-2 rounded-md text-sm text-slate-400 hover:text-white hover:bg-slate-800 text-left"
                  >
                    {s.icon}
                    {s.title}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* 正文内容 */}
          <main className="flex-1 space-y-6">
            {/* ─── 产品概述 ─── */}
            <SectionCard id="overview" title="产品概述">
              <p>
                <strong className="text-white">Claw Swarm Operation Console</strong>{" "}
                是一套多集群 Kubernetes 容器实例管理平台，用于统一调度和管理
                <strong className="text-white"> OpenClaw</strong> 会话容器。
              </p>
              <p>核心能力：</p>
              <ul className="space-y-1.5 list-none ml-0">
                {[
                  ["实例池管理", "维护一批预先就绪的 OpenClaw 容器，按需分配给用户会话"],
                  ["一键分配/释放", "通过用户 ID 将空闲实例分配给指定用户，用完即释放回池"],
                  ["实例生命周期", "支持暂停 / 恢复 / 删除实例，实时监控 CPU & 内存"],
                  ["多租户用户管理", "管理员可创建账号并分配角色（user / admin）"],
                  ["REST API", "提供完整 HTTP API，支持程序化接入"],
                ].map(([title, desc]) => (
                  <li key={title} className="flex gap-2">
                    <span className="text-cyan-400 mt-0.5">▸</span>
                    <span>
                      <strong className="text-white">{title}</strong> — {desc}
                    </span>
                  </li>
                ))}
              </ul>
            </SectionCard>

            {/* ─── 登录与认证 ─── */}
            <SectionCard id="login" title="登录与认证">
              <div className="space-y-4">
                <Step
                  num={1}
                  title="打开登录页"
                  desc={
                    <>
                      访问控制台地址，未登录时自动跳转到{" "}
                      <Code>/login</Code> 页面。
                    </>
                  }
                />
                <Step
                  num={2}
                  title="填写凭据"
                  desc="输入管理员分配给你的用户名和密码，点击 Login。"
                />
                <Step
                  num={3}
                  title="首次登录改密"
                  desc={
                    <>
                      若账号被标记了 <strong className="text-white">Force Change Password</strong>，
                      登录后需先在右上角用户菜单 → <strong className="text-white">Change Password</strong> 修改密码。
                    </>
                  }
                />

                <div className="border-t border-slate-800 pt-4 space-y-3">
                  <p className="text-white font-medium">其他账号操作（右上角用户菜单）</p>
                  <div className="grid gap-2">
                    <div className="flex items-center gap-2">
                      <IconLabel icon={<KeyRound className="h-3 w-3" />} label="Change Password" />
                      <span className="text-slate-400">修改当前账号密码（最少 6 位）</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <IconLabel icon={<Copy className="h-3 w-3" />} label="API Secret" />
                      <span className="text-slate-400">
                        查看 / 复制 / 重新生成 API 密钥（用于程序化调用）
                      </span>
                    </div>
                  </div>
                </div>

                <Tip>Token 保存在浏览器本地存储中，关闭标签页后仍保持登录状态，点击 Logout 可主动退出。</Tip>
              </div>
            </SectionCard>

            {/* ─── 控制台总览 ─── */}
            <SectionCard id="dashboard" title="控制台总览（Dashboard）">
              <p>
                登录后默认进入 Dashboard，提供集群全局概况。
              </p>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-3 py-2">
                {[
                  ["总实例数", "blue", "全部 OpenClaw 实例数量"],
                  ["平均 CPU", "cyan", "所有运行中实例的 CPU 均值"],
                  ["平均内存", "purple", "所有运行中实例的内存均值"],
                  ["已分配数", "green", "当前已绑定用户的实例数量"],
                ].map(([label, color, desc]) => (
                  <div
                    key={label}
                    className="rounded-lg bg-slate-800 p-3 space-y-1"
                  >
                    <p className={`text-xs text-${color}-400 font-medium`}>{label}</p>
                    <p className="text-xs text-slate-500">{desc}</p>
                  </div>
                ))}
              </div>
              <div className="space-y-2">
                <p className="text-white font-medium">状态分布</p>
                <div className="flex flex-wrap gap-2">
                  <Badge className="bg-green-900/40 text-green-400 border-green-800">Active — 运行中</Badge>
                  <Badge className="bg-yellow-900/40 text-yellow-400 border-yellow-800">Idle — 空闲</Badge>
                  <Badge className="bg-slate-700 text-slate-400 border-slate-600">Stopped — 已停止</Badge>
                  <Badge className="bg-red-900/40 text-red-400 border-red-800">Error — 异常</Badge>
                </div>
              </div>
              <p>
                页面下方的实例列表展示最近 5 条记录，支持按{" "}
                <Badge className="text-xs bg-slate-800 text-slate-300 border-slate-700">All</Badge>{" "}
                <Badge className="text-xs bg-slate-800 text-slate-300 border-slate-700">Running</Badge>{" "}
                <Badge className="text-xs bg-slate-800 text-slate-300 border-slate-700">Paused</Badge>{" "}
                过滤，点击实例名称可进入详情页。
              </p>
              <Tip>Dashboard 每 10 秒自动刷新一次，也可点击页面中的刷新按钮手动触发。</Tip>
            </SectionCard>

            {/* ─── 实例管理 ─── */}
            <SectionCard id="instances" title="实例管理（Instances）">
              <p>
                点击顶部导航{" "}
                <IconLabel icon={<Grid3x3 className="h-3 w-3" />} label="Instances" />{" "}
                进入实例列表页，查看所有实例。
              </p>

              <div className="space-y-3">
                <p className="text-white font-medium">过滤与搜索</p>
                <ul className="space-y-1.5 ml-4 list-disc list-outside">
                  <li>顶部 Tab：<strong className="text-white">All / Allocated / Free</strong>，快速切换已分配 / 空闲视图</li>
                  <li>搜索框：按实例名称或用户 ID 模糊匹配</li>
                </ul>
              </div>

              <div className="space-y-3">
                <p className="text-white font-medium">实例卡片操作（右上角 ⋮ 菜单）</p>
                <div className="grid gap-2">
                  {[
                    [<Pause className="h-3 w-3" />, "Pause Instance", "暂停实例（停止资源使用，保留绑定关系）"],
                    [<Play className="h-3 w-3" />, "Resume Instance", "恢复已暂停的实例"],
                    [<Hash className="h-3 w-3" />, "Allocate to User", "将空闲实例绑定到指定用户 ID"],
                    [<RefreshCw className="h-3 w-3" />, "Free Instance", "解除绑定，将实例归还实例池"],
                    [<Trash2 className="h-3 w-3 text-red-400" />, "Delete Instance", "永久删除实例（不可恢复）"],
                  ].map(([icon, label, desc]) => (
                    <div key={label as string} className="flex items-start gap-2">
                      <IconLabel icon={icon} label={label as string} />
                      <span className="text-slate-400 mt-0.5">{desc}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="space-y-3">
                <p className="text-white font-medium">实例详情页</p>
                <p>点击实例名称进入详情，可查看：</p>
                <ul className="space-y-1.5 ml-4 list-disc list-outside">
                  <li>基本信息（名称、状态、命名空间）</li>
                  <li>资源配额（CPU / 内存 Request & Limit）</li>
                  <li>
                    <strong className="text-white">Claw Web UI</strong> — 实例自带的前端页面链接，可直接打开或复制
                  </li>
                  <li>
                    <strong className="text-white">Gateway Token</strong> — 网关访问凭证，点击{" "}
                    <IconLabel icon={<Eye className="h-3 w-3" />} label="Reveal Token" /> 后可复制
                  </li>
                </ul>
              </div>

              <Warn>Delete 操作不可撤销，请确认实例不再使用后再删除。</Warn>
            </SectionCard>

            {/* ─── 分配实例 ─── */}
            <SectionCard id="allocate" title="分配实例（Allocate Instance）">
              <p>
                点击顶部右侧{" "}
                <span className="inline-flex items-center gap-1 bg-cyan-600 text-white text-xs px-2 py-0.5 rounded">
                  <Plus className="h-3 w-3" />
                  Allocate Instance
                </span>{" "}
                按钮，进入分配流程。
              </p>

              <div className="space-y-4">
                <Step
                  num={1}
                  title="输入用户 ID"
                  desc={
                    <>
                      在输入框中填写目标用户的唯一标识（如会话 ID、用户名等），该 ID 将与实例绑定，
                      后续可通过此 ID 查询或释放实例。
                    </>
                  }
                />
                <Step
                  num={2}
                  title="点击 Allocate"
                  desc="系统从空闲池中选取一个实例，自动完成绑定操作。"
                />
                <Step
                  num={3}
                  title="查看分配结果"
                  desc={
                    <>
                      成功后页面展示：
                      <ul className="mt-1 space-y-1 ml-4 list-disc list-outside text-slate-400">
                        <li>实例名称与当前状态</li>
                        <li>绑定的用户 ID（可复制）</li>
                        <li>Gateway Token（可复制）</li>
                        <li>Claw Web UI 链接（可直接打开）</li>
                      </ul>
                    </>
                  }
                />
              </div>

              <Tip>若当前没有空闲实例，分配会失败并提示错误。请先确保实例池中有空闲实例，或联系管理员扩充容量。</Tip>
            </SectionCard>

            {/* ─── 用户管理 ─── */}
            <SectionCard id="users" title="用户管理（仅管理员）">
              <p>
                管理员账号在导航栏可见{" "}
                <IconLabel icon={<Users className="h-3 w-3" />} label="Users" />{" "}
                入口，用于管理系统用户。
              </p>

              <div className="space-y-3">
                <p className="text-white font-medium">创建用户</p>
                <div className="space-y-3">
                  <Step num={1} title="点击 New User" desc="右上角新建用户按钮，弹出创建对话框。" />
                  <Step
                    num={2}
                    title="填写信息"
                    desc={
                      <ul className="mt-1 space-y-1 ml-4 list-disc list-outside text-slate-400">
                        <li>Username — 登录用户名</li>
                        <li>Password — 初始密码（最少 6 位）</li>
                        <li>Role — <Code>user</Code> 普通用户 / <Code>admin</Code> 管理员</li>
                      </ul>
                    }
                  />
                  <Step num={3} title="确认创建" desc="点击 Create，用户即时生效。" />
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-white font-medium">删除用户</p>
                <p>在用户列表行末点击删除按钮（当前登录账号不可自删）。</p>
              </div>

              <div className="space-y-2">
                <p className="text-white font-medium">字段说明</p>
                <div className="rounded-md bg-slate-800 overflow-hidden">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="border-b border-slate-700">
                        <th className="px-3 py-2 text-left text-slate-400">字段</th>
                        <th className="px-3 py-2 text-left text-slate-400">说明</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700/50">
                      {[
                        ["ID", "系统自动分配的用户编号"],
                        ["Username", "登录用户名"],
                        ["Role", "admin 可访问所有功能；user 只能操作实例"],
                        ["Force Change Pwd", "首次登录强制修改密码的标记"],
                        ["Created At", "账号创建时间"],
                      ].map(([field, desc]) => (
                        <tr key={field}>
                          <td className="px-3 py-2 text-white font-mono">{field}</td>
                          <td className="px-3 py-2 text-slate-400">{desc}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </SectionCard>

            {/* ─── 审计日志 ─── */}
            <SectionCard id="audit" title="审计日志（仅管理员）">
              <p>
                系统会自动记录所有实例生命周期操作的审计日志，管理员可在导航栏{" "}
                <IconLabel icon={<Shield className="h-3 w-3" />} label="审计日志" />{" "}
                入口查看。
              </p>

              <div className="space-y-3">
                <p className="text-white font-medium">记录哪些操作？</p>
                <div className="rounded-md bg-slate-800 overflow-hidden">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="border-b border-slate-700">
                        <th className="px-3 py-2 text-left text-slate-400">操作</th>
                        <th className="px-3 py-2 text-left text-slate-400">触发时机</th>
                        <th className="px-3 py-2 text-left text-slate-400">附加详情</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700/50">
                      {[
                        ["分配（alloc）", "POST /claw/alloc 成功后", "model_type=lite / pro"],
                        ["释放（free）", "POST /claw/free 成功后", "—"],
                        ["暂停（pause）", "POST /claw/pause 成功后", "delay_minutes=N"],
                        ["恢复（resume）", "POST /claw/resume 成功后", "—"],
                      ].map(([action, trigger, detail]) => (
                        <tr key={action}>
                          <td className="px-3 py-2 text-white font-medium">{action}</td>
                          <td className="px-3 py-2 text-slate-400">{trigger}</td>
                          <td className="px-3 py-2 text-slate-500 font-mono">{detail}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              <div className="space-y-3">
                <p className="text-white font-medium">日志字段说明</p>
                <div className="rounded-md bg-slate-800 overflow-hidden">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="border-b border-slate-700">
                        <th className="px-3 py-2 text-left text-slate-400">字段</th>
                        <th className="px-3 py-2 text-left text-slate-400">说明</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700/50">
                      {[
                        ["时间", "操作发生的本地时间"],
                        ["操作者", "执行操作的管理员账号"],
                        ["操作", "alloc / free / pause / resume"],
                        ["目标用户", "被操作实例绑定的用户 ID"],
                        ["实例", "被操作的实例名称（StatefulSet name）"],
                        ["状态", "success（成功）"],
                        ["详情", "附加参数，如模型类型或延迟分钟数"],
                      ].map(([field, desc]) => (
                        <tr key={field}>
                          <td className="px-3 py-2 text-white font-mono">{field}</td>
                          <td className="px-3 py-2 text-slate-400">{desc}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-white font-medium">页面功能</p>
                <ul className="space-y-1.5 ml-4 list-disc list-outside">
                  <li>顶部下拉框按操作类型筛选（全部 / 分配 / 释放 / 暂停 / 恢复）</li>
                  <li>分页浏览，每页 20 条，按时间倒序排列</li>
                  <li>右上角刷新按钮手动重新加载</li>
                </ul>
              </div>

              <div className="space-y-2">
                <p className="text-white font-medium">API 查询</p>
                <p>
                  可通过 <Code>GET /audit/logs</Code> 程序化查询，支持 <Code>page</Code>、
                  <Code>page_size</Code>、<Code>action</Code> 参数过滤。
                </p>
                <CodeBlock>{`curl "http://<host>/audit/logs?page=1&page_size=20&action=alloc" \\
  -H "X-API-Key: <your-api-secret>"`}</CodeBlock>
              </div>

              <Tip>审计日志页面及 API 仅对 admin 角色开放，普通用户无法访问。</Tip>
            </SectionCard>

            {/* ─── API 访问 ─── */}
            <SectionCard id="api" title="API 访问">
              <p>
                所有 UI 功能均有对应的 REST API，可供外部系统（如 CI/CD、Agent 等）程序化调用。
              </p>

              <div className="space-y-3">
                <p className="text-white font-medium">获取 API 密钥</p>
                <p>
                  右上角用户菜单 → <strong className="text-white">API Secret</strong>，查看或重新生成密钥。
                  所有 API 请求需在 Header 中携带：
                </p>
                <CodeBlock>X-API-Key: &lt;your-api-secret&gt;</CodeBlock>
              </div>

              <div className="space-y-3">
                <p className="text-white font-medium">常用接口</p>
                <div className="rounded-md bg-slate-800 overflow-hidden">
                  <table className="w-full text-xs">
                    <thead>
                      <tr className="border-b border-slate-700">
                        <th className="px-3 py-2 text-left text-slate-400">方法</th>
                        <th className="px-3 py-2 text-left text-slate-400">路径</th>
                        <th className="px-3 py-2 text-left text-slate-400">说明</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-700/50">
                      {[
                        ["GET", "/claw/instances", "获取实例列表"],
                        ["GET", "/claw/instances/{name}", "获取单个实例详情"],
                        ["POST", "/claw/alloc", "分配实例（body: {user_id}）"],
                        ["POST", "/claw/free", "释放实例（body: {name}）"],
                        ["POST", "/claw/pause", "暂停实例（body: {name}）"],
                        ["POST", "/claw/resume", "恢复实例（body: {name}）"],
                        ["GET", "/claw/token?name={name}", "获取网关 Token"],
                        ["GET", "/auth/me", "获取当前用户信息"],
                        ["POST", "/auth/login", "登录（body: {username, password}）"],
                        ["GET", "/users", "获取用户列表（仅 admin）"],
                        ["POST", "/users", "创建用户（仅 admin）"],
                        ["DELETE", "/users/{id}", "删除用户（仅 admin）"],
                        ["GET", "/audit/logs", "查询审计日志（仅 admin）"],
                      ].map(([method, path, desc]) => (
                        <tr key={path + method}>
                          <td className="px-3 py-2">
                            <span
                              className={`font-mono font-bold ${
                                method === "GET"
                                  ? "text-green-400"
                                  : method === "POST"
                                  ? "text-cyan-400"
                                  : "text-red-400"
                              }`}
                            >
                              {method}
                            </span>
                          </td>
                          <td className="px-3 py-2 text-white font-mono">{path}</td>
                          <td className="px-3 py-2 text-slate-400">{desc}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              <div className="space-y-2">
                <p className="text-white font-medium">示例：分配实例</p>
                <CodeBlock>{`curl -X POST http://<host>/claw/alloc \\
  -H "X-API-Key: <your-api-secret>" \\
  -H "Content-Type: application/json" \\
  -d '{"user_id": "user-12345"}'`}</CodeBlock>
              </div>

              <div className="space-y-2">
                <p className="text-white font-medium">示例：释放实例</p>
                <CodeBlock>{`curl -X POST http://<host>/claw/free \\
  -H "X-API-Key: <your-api-secret>" \\
  -H "Content-Type: application/json" \\
  -d '{"name": "claw-abc123"}'`}</CodeBlock>
              </div>
            </SectionCard>

            {/* ─── 典型使用流程 ─── */}
            <SectionCard id="workflow" title="典型使用流程">
              <div className="space-y-6">
                <div className="space-y-4">
                  <p className="text-white font-medium flex items-center gap-2">
                    <span className="bg-cyan-600/20 text-cyan-400 border border-cyan-600/30 px-2 py-0.5 rounded text-xs">场景一</span>
                    运维人员手动为用户分配会话容器
                  </p>
                  <div className="space-y-3 pl-4 border-l border-slate-700">
                    <Step num={1} title="登录控制台" desc="使用管理员或运维账号登录。" />
                    <Step num={2} title="点击 Allocate Instance" desc="填写用户 ID（如会话 ID）。" />
                    <Step num={3} title="获取连接信息" desc="复制 Gateway Token 和 Claw Web UI 链接，发给用户。" />
                    <Step num={4} title="用完后释放" desc="在 Instances 列表找到该实例，选择 Free Instance 归还资源池。" />
                  </div>
                </div>

                <div className="space-y-4">
                  <p className="text-white font-medium flex items-center gap-2">
                    <span className="bg-purple-600/20 text-purple-400 border border-purple-600/30 px-2 py-0.5 rounded text-xs">场景二</span>
                    外部 Agent 系统自动化分配
                  </p>
                  <div className="space-y-3 pl-4 border-l border-slate-700">
                    <Step
                      num={1}
                      title="获取 API Secret"
                      desc="登录控制台 → 用户菜单 → API Secret，复制密钥。"
                    />
                    <Step
                      num={2}
                      title="调用 /claw/alloc"
                      desc={
                        <>
                          携带 <Code>X-API-Key</Code> Header 和用户 ID 调用分配接口，
                          响应中包含实例名、Token 和 Web UI 链接。
                        </>
                      }
                    />
                    <Step
                      num={3}
                      title="会话结束后调用 /claw/free"
                      desc="传入实例名释放资源，使其回到空闲池。"
                    />
                  </div>
                </div>

                <div className="space-y-4">
                  <p className="text-white font-medium flex items-center gap-2">
                    <span className="bg-amber-600/20 text-amber-400 border border-amber-600/30 px-2 py-0.5 rounded text-xs">场景三</span>
                    临时暂停资源占用
                  </p>
                  <div className="space-y-3 pl-4 border-l border-slate-700">
                    <Step
                      num={1}
                      title="找到目标实例"
                      desc="在 Instances 列表中搜索实例名或用户 ID。"
                    />
                    <Step
                      num={2}
                      title="执行 Pause"
                      desc="实例进入暂停态，不再消耗 CPU，但保留绑定关系。"
                    />
                    <Step
                      num={3}
                      title="恢复时执行 Resume"
                      desc="实例重新进入运行态，原有绑定关系不变。"
                    />
                  </div>
                </div>
              </div>

              <Tip>建议在生产环境中，通过 API 自动化完成分配和释放，减少人工操作，提升资源利用率。</Tip>
            </SectionCard>

            {/* 底部 */}
            <div className="text-center py-6 text-slate-600 text-xs">
              Claw Swarm Operation Console — 如有问题请联系管理员
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
