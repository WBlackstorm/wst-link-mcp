# 🚀 LinkMCP

> **High-Performance, Zero-Dependency LinkedIn MCP Server in Golang.**

`linkmcp` is an ultra-lightweight **Model Context Protocol (MCP)** server built in Go. It enables AI Assistants (like Claude Desktop, Cursor, and custom LLM Agents) to interact natively with the LinkedIn Official API to publish posts and fetch profile analytics.

Unlike Node.js or Python implementations, `linkmcp` compiles down to a single static binary of **~10MB**, starts in **milliseconds**, and consumes **under 8MB of RAM**.

---

## ✨ Features

- ⚡ **Ultra Lightweight:** Written in pure Go with minimal memory footprint.
- 📦 **Single Static Binary:** No `npm`, `node`, or `python` environments required.
- 🔐 **Official LinkedIn API Integration:** Safe and compliant with LinkedIn Developer Policies.
- 🛠️ **MCP Native:** Exposes Tools and Resources for seamless integration with Claude Desktop & Cursor.
- 💬 **Text Posts Publishing:** Allow your LLM agent to generate and post text directly to your feed.

---

## 🛠️ Installation

### Option 1: Quick Install (Binary)

Download the latest binary for your operating system from the [Releases](https://github.com/wblackstorm/wst-link-mcp/releases) page.

### Option 2: Build from Source

Requirements: **Go 1.22+**

```bash
# Clone the repository
git clone https://github.com/wblackstorm/wst-link-mcp.git
cd wst-link-mcp

# Build the binary
go build -o linkmcp cmd/linkmcp/main.go
```

---

## ⚙️ Configuration

### 1. Get a LinkedIn Access Token

To use the local version, you need a LinkedIn Access Token with the `w_member_social` and `openid` scope:

1. Go to the [LinkedIn Developer Portal](https://www.linkedin.com/developers/).
2. Create an App and enable **Share on LinkedIn** and **Sign In with LinkedIn using OpenID Connect**.
3. Generate a Member Access Token in the **OAuth 2.0 Tools** section.

---

## 🔌 Integration with Claude Desktop

Add `linkmcp` to your `claude_desktop_config.json`:

- **MacOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "linkedin": {
      "command": "/caminho/para/seu/linkmcp",
      "env": {
        "LINKEDIN_ACCESS_TOKEN": "seu_access_token_aqui"
      }
    }
  }
}
```

---

## 🧰 Available MCP Tools

### `publish_linkedin_post`
Publishes a text post to your LinkedIn profile.

**Parameters:**
- `commentary` (string, required): The text content of your post.

### `get_linkedin_profile`
Fetches the current authenticated user's LinkedIn profile URN and basic information.

---

## 🚀 Roadmap

- [x] Basic Text Post Publishing (`/rest/posts`)
- [ ] Image and PDF/Carousel Upload Support
- [ ] Post Analytics Tool (impressions, likes, comments)
- [ ] SSE (Server-Sent Events) Mode for Cloud Hosting
- [ ] Automated Token Refresher CLI

---

## 📄 License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.