# Seli - 命令行启动器

[English](../README.md) | 简体中文

Seli 是一个基于 TUI 的命令行工具启动器，让你方便地管理和执行预设的命令。

## ✨ 功能特性

- 🎨 **美观的终端用户界面** - 现代化的 TUI 设计
- 📁 **文件夹和文件浏览** - 支持层级目录导航
- 📄 **多格式配置文件** - 支持 JSON、YAML、TOML 格式
- 🚀 **环境变量支持** - 支持 `.env` 文件和命令级环境变量
- 🔄 **智能变量替换** - 支持环境变量在配置中的动态替换
- 🎯 **命令显示控制** - 通过 `show` 字段控制命令可见性
- 📂 **工作目录配置** - 每个命令可设置独立工作目录
- ⌨️ **键盘快捷键** - 直观的键盘操作
- 🏠 **自动配置目录** - 自动创建 `~/.seli/` 配置目录
- 🔄 **循环导航** - 列表首尾循环导航

## 🎬 演示

![Terminal Demo Animation](../demo.gif)

## 🚀 快速开始

### 1. 安装

```bash
go install github.com/shapled/seli@latest
```

### 2. 创建配置

```bash
# create config directory
mkdir ~/.seli

# create env file
echo 'TEST_ENV_A=Apple
TEST_ENV_B=Banana' > ~/.seli/.env

# create config file
echo 'name: Fruits Commands
description: Demonstrates setting and using specific environment variables for command execution.

commands:
  - name: "Show Fruit A"
    description: "Sets TEST_ENV_A and prints it."
    command: "echo"
    args: ["Fruit A is: ${TEST_ENV_A}"]

  - name: "Show Fruit B"
    description: "Sets TEST_ENV_B and runs in tmp directory."
    command: "sh"
    args: ["-c", "echo \\${PWD}; echo Fruit B is: ${TEST_ENV_B}"]
    workDir: "/tmp"
    show: true

  - name: "Show Fruit C"
    description: "Sets TEST_ENV_C and shows usage."
    command: "echo"
    args: ["Cherry", "details:", "${TEST_ENV_C}"]
    env:
      TEST_ENV_C: "Cherry - Often used in juice"' > ~/.seli/fruits.yml

# run
seli
```

## 使用方法

### 1. 运行程序

```bash
seli
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

## 📖 配置文件字段说明

### 命令字段

| 字段          | 类型              | 必填 | 说明                 |
| ------------- | ----------------- | ---- | -------------------- |
| `name`        | string            | 是   | 配置文件或命令的名称 |
| `description` | string            | 否   | 描述信息             |
| `command`     | string            | 是   | 要执行的命令         |
| `args`        | []string          | 否   | 命令参数             |
| `env`         | map[string]string | 否   | 命令级环境变量       |
| `workDir`     | string            | 否   | 工作目录             |
| `show`        | bool              | 否   | 是否显示在命令列表中 |

### 环境变量优先级

环境变量的替换遵循以下优先级（从高到低）：

1. **命令级环境变量** (`env` 字段中定义的变量)
2. **`.env` 文件中的变量** (`~/.seli/.env` 和配置目录下的 `.env`)
3. **系统环境变量** (系统级的环境变量)

### 变量替换规则

- 支持 `${VAR_NAME}` 格式的变量替换
- 命令级环境变量可以引用 `.env` 文件中的变量
- 支持转义字符 `\${VAR_NAME}` 来避免变量替换
- 变量替换在配置加载时进行

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License