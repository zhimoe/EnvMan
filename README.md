# EnvMan

一个紧凑、原生的 Windows 环境变量管理工具。界面使用
[MyGo](https://github.com/egoist/mygo) native UI 绘制，不依赖 WebView、
HTML、CSS 或 JavaScript。

![EnvMan 界面截图](doc/screenshot.png)

## 特性

- **双范围浏览**：同时查看用户变量与系统变量，并标记同名覆盖关系。
- **快速定位**：搜索、排序、键盘导航、批量选择、复制与 `.env` 导出。
- **多值变量**：PATH 等 `;` 分隔变量可逐条展开，支持列表编辑、排序与整段编辑。
- **实时差异**：整段编辑时实时显示新增、移除、重复及顺序变化，附带行号。
- **直接写注册表**：读写 Windows 注册表，不使用会截断长值的 `setx`。
- **按需提权**：查看与导出系统变量无需权限；只有实际写入时才通过 UAC 提权。
- **自动备份**：每次写入前自动备份，保留最近 20 份，支持恢复与撤销。
- **并发安全**：保存时保留其他程序在 EnvMan 打开期间新增或修改的无关变量。
- **主题与高 DPI**：浅色、深色、跟随系统三种主题，原生高 DPI 渲染。

## 开发

需要 Go 1.27.3 或更高版本。MyGo CLI 已通过 `go.mod` 的 `tool` 指令固定为
v0.8.3；所有 Go 依赖都提交在 `vendor/` 中。

```sh
# 在当前平台运行；非 Windows 平台使用内置演示数据
go run ./cmd/envman

# 运行交互测试和静态检查
go test ./...
go vet ./...
go tool mygo vet .

# 构建 Windows x64 可执行文件及安装器
go tool mygo build -platform windows/amd64 .
```

构建产物位于 `build/windows-amd64/`：

- `EnvMan.exe`
- `envman-setup-<version>-amd64.exe`

Go 1.14 及之后的工具链在存在 `vendor/` 时会自动使用它。也可以显式进行离线校验：

```sh
GOPROXY=off GOSUMDB=off go test -mod=vendor ./...
```

### Windows 原生集成测试

普通测试不会修改真实环境变量。需要在 Windows 上验证注册表类型、差异写入、
非字符串值保护和环境变化广播时，可显式运行临时变量集成测试；测试结束会自动清理：

```powershell
$env:ENVMAN_WINDOWS_INTEGRATION = "1"
go test -count=1 -run '^TestWindowsUserRegistryIntegration$' ./internal/environment
```

系统变量测试必须从“以管理员身份运行”的 PowerShell 执行：

```powershell
$env:ENVMAN_WINDOWS_SYSTEM_INTEGRATION = "1"
go test -count=1 -run '^TestWindowsSystemRegistryIntegration$' ./internal/environment
```

安装器最终验收还应确认：标准权限启动不弹 UAC；只有保存系统变量时弹 UAC；
提权重启后修改生效且待应用文件被清理；新开的终端能读到修改后的值。

升级依赖后必须同步 vendor 内容：

```sh
go get github.com/egoist/mygo@<version>
go get -tool github.com/egoist/mygo/cmd/mygo@<version>
go mod tidy
go mod vendor
```

## 界面尺寸约定

界面尺寸集中在两个文件：`internal/app/metrics.go`（字号、间距、控件高度）和
`internal/app/theme.go`（圆角、配色）。改一处即可全局生效，不要在新代码里写字面量。

**圆角**：沿用 Windows Fluent 的两档语义——控件 4、浮层 8，胶囊为全圆。
图标方块按尺寸取 4（<32）或 8（≥32）。

**字号**：共 5 档，随系统"文本缩放"等比放大。

| token | 值 | 用途 |
| --- | --- | --- |
| `fontCaption` | 10 | 分区标签、页脚、行内徽章 |
| `fontSmall` | 11 | 提示、callout、次要说明 |
| `fontBody` | 12 | 默认：行值、按钮、导航项 |
| `fontSubtitle` | 14 | 抽屉、弹窗、空状态标题 |
| `fontTitle` | 18 | 页面标题 |

**间距**：4 栅格加一个 2 半步，`space2 / 4 / 6 / 8 / 12 / 16 / 24 / 32`。
所有 `Gap` 与容器 `Padding` 都取自这一档；组件内部的光学微调（如 1px、1.5px）
保留字面量并就近注释。

**高度**：不拍数字，由行框推导。Segoe UI 的行框约为字号的 1.33 倍，正文 12 →
行框 16，再加垂直内边距：

| token | 值 | 公式 |
| --- | --- | --- |
| `controlXs` | 24 | 行框 + 2×4（行内按钮、图标按钮） |
| `controlSm` | 28 | 行框 + 2×6（轻按钮、行内输入） |
| `controlMd` | 32 | 行框 + 2×8（主次按钮、输入框） |
| `rowH` / `barH` | 32 | 行框 + 2×8（表格行、表头表尾、横幅） |
| `titleBarH` / `paneHeaderH` | 40 | 控件 + 2×8 |
| `pageHeaderH` | 48 | 控件 + 2×8 |
| `sheetFooterH` | 56 | 控件 + 3×8 |

原则：**能由内边距算出来的高度就不写 `Height`**，只有必须固定或等宽的控件
（图标按钮、`List` 行高）才指定高度，这样文本缩放时布局不会错位。

## Windows 数据与权限

- 用户变量：`HKCU\Environment`
- 系统变量：`HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment`
- 备份和待应用更改：`%AppData%\EnvMan`
- 修改后广播 `WM_SETTINGCHANGE`；已经打开的终端不会自动刷新，请新开终端。

EnvMan 默认以标准权限启动。系统变量可以查看、复制和导出；只有实际写入时才请求
UAC。注册表中的非字符串值不会被 EnvMan 删除。

## 项目结构

- `cmd/envman/`：程序入口；`mygo.json` 的 `main` 字段指向此包。
- `internal/app/`：MyGo 原生界面（`metrics.go` 尺寸、`theme.go` 配色与圆角）、
  应用状态和提交/撤销等流程编排。
- `internal/domain/`：环境变量领域类型、差异计算和并发变更合并。
- `internal/environment/`：Windows 注册表后端与非 Windows 内存后端。
- `internal/storage/`：快照、待提权变更和设置的 JSON 持久化。
- 各包内的 `*_test.go`：领域、平台、存储与 native UI 交互测试。
- `vendor/`：应用与 MyGo 构建工具的完整依赖源码。
