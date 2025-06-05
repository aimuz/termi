# Termi

自然语言转 Bash 命令行助手

Termi 让你可以用自然语言描述想要做的事情，它会调用多种 LLM（OpenAI GPT、Azure OpenAI、Google Gemini、Anthropic Claude、Llama.cpp 等）将其翻译成可以直接粘贴执行的 Bash 命令，并在终端内提供交互式候选选择与一键执行。

---

## 功能亮点

1. **自然语言 ➜ 命令行**：输入任何中文需求，Termi 会返回可直接运行的 Bash 命令。
2. **智能追问**：当关键信息缺失时，LLM 会自动以中文向你提问，补全上下文。
3. **TUI 候选列表**：基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 的终端 UI，显示多条候选命令，↑/↓ 选择，Enter 执行。
4. **完整交互执行**：命令通过 `bash -c` 启动，标准输入/输出与当前终端直连，体验与手动输入无异。
5. **零本地规则依赖**：所有解析逻辑均在 LLM 中完成，代码简洁易于扩展。

---

## 快速开始

### 1. 安装 Go

确保本机拥有 **Go ≥ 1.23**（推荐使用最新版）。

```bash
$ go version
```

### 2. 获取源码

```bash
$ git clone https://github.com/aimuz/termi.git
$ cd termi
```

### 3. 配置 LLM 提供商

Termi 支持多种 LLM 提供商，请设置其中一个环境变量：

#### OpenAI
```bash
$ export OPENAI_API_KEY="sk-..."
$ export OPENAI_BASE_URL="https://api.openai.com"  # 可选，自定义API地址
```

#### Azure OpenAI
```bash
$ export AZURE_OPENAI_API_KEY="your-key"
$ export AZURE_OPENAI_BASE_URL="https://your-resource.openai.azure.com"
$ export AZURE_OPENAI_DEPLOYMENT_ID="your-deployment-id"
$ export AZURE_OPENAI_API_VERSION="2023-12-01-preview"  # 可选
```

#### Google Gemini
```bash
$ export GEMINI_API_KEY="your-gemini-api-key"
$ export GEMINI_MODEL="gemini-pro"  # 可选，默认为 gemini-pro
```

#### Anthropic Claude
```bash
$ export ANTHROPIC_API_KEY="your-claude-api-key"
$ export CLAUDE_MODEL="claude-3-haiku-20240307"  # 可选
```

#### Llama.cpp（本地部署）
```bash
$ export LLAMA_CPP_BASE_URL="http://localhost:8080"
```

或者，你也可以创建配置文件 `~/.config/termi/config.json`：

```json
{
  "llm": {
    "provider": "openai",
    "openai": {
      "api_key": "your-api-key",
      "model": "gpt-3.5-turbo",
      "timeout": 30
    }
  }
}
```

参考 `config.example.json` 获取完整配置示例。

### 4. 编译 / 安装

```bash
# 在当前目录编译二进制
$ go build -o termi .

# 或直接安装到 GOPATH/bin，并加入 PATH
$ go install termi.sh/termi@latest
```

### 5. 使用示例

```bash
$ termi 我想对 baidu.com 发起 ping

候选命令 (↑/↓ 选择, Enter 确定, q 退出):
➜ 1. ping -c 4 baidu.com [llm]
  2. ping baidu.com [llm]
```

按 **Enter** 即开始执行，输出与你在终端直接输入 `ping -c 4 baidu.com` 完全一致。

> 如果缺少参数，Termi 会先向你提问：
>
> ```bash
> $ termi 删除文件
> 你要删除哪个文件(支持通配符)? _
> ```
>
> 填写后继续生成命令并进入候选界面。

---

## 工作原理

```mermaid
graph TD
    A[自然语言语句] --> B(AskSmart 调用 GPT-3.5/GPT-4o)
    B -->|返回 command| C[候选列表]
    B -->|返回 ask| D[追问用户]
    D -->|补全| B
    C --> E[Bubble Tea UI 选择]
    E --> F[Runner bash -c 执行]
```

核心包概览：

- `internal/llm` OpenAI API 封装，提供 `Ask` / `AskCommand` / `AskSmart` 三种模式。
- `internal/ui` Bubble Tea TUI，实现候选展示与加载动画。
- `internal/runner` 包装 `exec.Command`，负责命令执行与 I/O 直通。
- `internal/suggest` 候选命名空间（预留本地规则扩展）。
- `cmd/termi/main.go` CLI 入口，整合各组件。

---

## 常见问题 FAQ

1. **为什么提示 "未找到任何 LLM 提供商配置"？**  
   请设置对应的环境变量或创建配置文件，参考上面的配置说明。
2. **支持 Windows 吗？**  
   理论上可以，但尚未充分测试，欢迎 PR。
3. **如何切换不同的 LLM 提供商？**  
   通过设置不同的环境变量或修改配置文件中的 `provider` 字段。
4. **可以同时配置多个提供商吗？**  
   可以，但同时只会使用一个提供商，优先级：配置文件 > 环境变量检测（OpenAI > Azure > Gemini > Claude > Llama.cpp）。

---

## 贡献指南

欢迎提交 Issue、PR，或在 Discussions 中交流新点子。

1. Fork 仓库并创建分支；
2. 保持 `go vet`, `go test` 通过；
3. 提交 PR 时附上说明截图 / 文字。

---

## Roadmap

- [ ] 支持本地命令规则建议
- [x] 接入更多 LLM (Gemini, Llama-cpp, Azure OpenAI, Claude)
- [ ] 增加批量模式，直接输出命令而不执行
- [ ] 增加 `--dry-run` / `--yes` 等安全选项
- [ ] 支持插件系统和自定义提供商

---

## License

MIT © 2025 aimuz & Contributors 
