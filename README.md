# Seli - 命令行启动器

Seli 是一个基于 TUI 的命令行工具启动器，让你方便地管理和执行预设的命令。

## 功能特性

- 🎨 美观的终端用户界面 (TUI)
- 📁 支持文件夹和文件浏览
- 📄 支持 JSON、YAML、TOML 配置文件格式
- 🚀 支持环境变量和工作目录配置
- ⌨️ 键盘快捷键操作
- 🏠 自动创建 `~/.seli/` 配置目录

## 安装

直接安装到 $GOPATH/bin

```bash
go install github.com/shapled/seli@latest
```

## 使用方法

### 1. 运行程序

```bash
./seli
```

### 2. 配置文件结构

在 `~/.seli/` 目录下创建配置文件，支持以下格式：

#### JSON 格式 (`development.json`)

```json
{
  "name": "Development Tools",
  "description": "Common development commands",
  "commands": [
    {
      "name": "Start Dev Server",
      "description": "Start the development server",
      "command": "npm",
      "args": ["run", "dev"],
      "env": {
        "NODE_ENV": "development",
        "PORT": "3000"
      }
    },
    {
      "name": "Git Status",
      "description": "Check git status",
      "command": "git",
      "args": ["status"]
    }
  ]
}
```

#### YAML 格式 (`system.yaml`)

```yaml
name: System Commands
description: System administration commands
commands:
  - name: "Disk Usage"
    description: "Check disk usage"
    command: "df"
    args: ["-h"]

  - name: "Memory Usage"
    description: "Check memory usage"
    command: "free"
    args: ["-h"]
```

#### TOML 格式 (`docker.toml`)

```toml
name = "Docker Commands"
description = "Docker container management commands"

[[commands]]
name = "List Containers"
description = "List all running containers"
command = "docker"
args = ["ps"]

[[commands]]
name = "Stop All Containers"
description = "Stop all running containers"
command = "docker"
args = ["stop", "$(docker ps -q)"]
```

### 3. 键盘操作

- **↑/↓** 或 **j/k**: 上下移动选择
- **Enter**: 选择文件/文件夹或执行命令
- **Backspace**: 返回上级目录（在命令列表中）
- **q**: 返回目录浏览（在命令列表中）
- **Esc/Ctrl+C**: 退出程序

### 4. 文件夹结构

```
~/.seli/
├── development.json    # 开发相关命令
├── system.yaml        # 系统管理命令
├── docker.toml        # Docker 相关命令
└── work/              # 工作相关配置
    ├── projects.json
    └── scripts.yaml
```

## 配置文件字段说明

| 字段          | 类型              | 必填 | 说明                 |
| ------------- | ----------------- | ---- | -------------------- |
| `name`        | string            | 是   | 配置文件或命令的名称 |
| `description` | string            | 否   | 描述信息             |
| `command`     | string            | 是   | 要执行的命令         |
| `args`        | []string          | 否   | 命令参数             |
| `env`         | map[string]string | 否   | 环境变量             |
| `workDir`     | string            | 否   | 工作目录             |

## 示例配置

项目已提供了一些示例配置文件，你可以根据需要修改：

- `development.json`: 开发工具命令
- `system.yaml`: 系统管理命令
- `docker.toml`: Docker 管理命令
- `work/projects.json`: 工作项目命令

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
