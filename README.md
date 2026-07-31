# 🚀 LinkMCP

> **High-Performance, Zero-Dependency LinkedIn MCP Server in Golang.**

`linkmcp` is an ultra-lightweight **Model Context Protocol (MCP)** server built in Go. It enables AI Assistants (like Claude Desktop, Cursor, and custom LLM Agents) to interact natively with the LinkedIn Official API to publish posts and fetch profile analytics.

Unlike Node.js or Python implementations, `linkmcp` compiles down to a single static binary of **~10MB**, starts in **milliseconds**, and consumes **under 8MB of RAM**.

---

## ✨ Features

- ⚡ **Ultra Lightweight:** Written in pure Go with minimal memory footprint.
- 📦 **Single Static Binary:** No `npm`, `node`, or `python` environments required.
- 🐳 **Docker Support:** Ready-to-use Dockerfile with multi-stage scratch image (~12MB total container size).
- 🔐 **Official LinkedIn API Integration:** Safe and compliant with LinkedIn Developer Policies.
- 🛠️ **MCP Native:** Exposes Tools and Resources for seamless integration with Claude Desktop & Cursor.
- 💬 **Text Posts Publishing:** Allow your LLM agent to generate and post text directly to your feed.

---

## 🛠️ Installation

### Option 1: Docker Image (Recommended)

Build the Docker image locally:

```bash
docker build -t wst-link-mcp:latest .
```

### Option 2: Quick Install (Binary)

Download the latest binary for your operating system from the [Releases](https://github.com/wblackstorm/wst-link-mcp/releases) page.

### Option 3: Build from Source

Requirements: **Go 1.22+**

```bash
# Clone the repository
git clone https://github.com/wblackstorm/wst-link-mcp.git
cd wst-link-mcp

# Build the binary
go build -o linkmcp cmd/linkmcp/main.go
```

### Makefile Commands

For convenience, a `Makefile` is provided with common workflow commands:

```bash
make help          # Show all available make targets
make build         # Build binary to bin/linkmcp
make run           # Run server locally (uses .env if present)
make stop          # Stop running processes or containers
make clean         # Remove build artifacts
make test          # Run unit tests
make docker-build   # Build Docker image (wst-link-mcp:latest)
make docker-run    # Run Docker container with .env file
make docker-stop   # Stop running Docker container
```

---

## ⚙️ Configuration

### 1. Get a LinkedIn Access Token

To use `linkmcp`, you need a LinkedIn Access Token with `w_member_social` and `openid` scopes:

1. Go to the [LinkedIn Developer Portal](https://www.linkedin.com/developers/).
2. Create an App and enable **Share on LinkedIn** and **Sign In with LinkedIn using OpenID Connect**.
3. Generate a Member Access Token in the **OAuth 2.0 Tools** section.

### 2. Environment Variables & `.env.example`

Copy `.env.example` to `.env` or set the environment variables in your MCP client configuration:

```env
# LinkedIn OAuth 2.0 Access Token (Required - with w_member_social and openid scopes)
LINKEDIN_ACCESS_TOKEN=your_linkedin_access_token_here

# LinkedIn API Base Host URL (Optional - default: https://api.linkedin.com)
LINKEDIN_BASE_URL=https://api.linkedin.com
```

| Environment Variable | Required | Default Value | Description |
| --- | --- | --- | --- |
| `LINKEDIN_ACCESS_TOKEN` | **Yes** | *None* | LinkedIn OAuth 2.0 member access token. |
| `LINKEDIN_BASE_URL` | **No** | `https://api.linkedin.com` | Base URL for LinkedIn API calls. |

---

## 🔌 Integration with Claude Desktop

Add `linkmcp` to your `claude_desktop_config.json`:

- **MacOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

### Option A: Using Docker (Containerized)

```json
{
  "mcpServers": {
    "linkedin": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "LINKEDIN_ACCESS_TOKEN=seu_access_token_aqui",
        "wst-link-mcp:latest"
      ]
    }
  }
}
```

> **Note:** The `-i` (interactive) flag is required for Stdio communication between Claude Desktop and Docker. Do not use `-t` (tty).

### Option B: Using Binary

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
- [x] Docker Container Support
- [ ] Image and PDF/Carousel Upload Support
- [ ] Post Analytics Tool (impressions, likes, comments)
- [ ] SSE (Server-Sent Events) Mode for Cloud Hosting
- [ ] Automated Token Refresher CLI

---

## 📄 License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.