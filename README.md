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
- 💬 **Text & Media Posts Publishing:** Publish text posts or posts with image/carousel media attachments.
- 📊 **Post Analytics:** Fetch impression, reaction, and comment metrics for your LinkedIn posts.
- 🌐 **Dual Transport Mode:** Supports both standard `stdio` mode and `sse` (Server-Sent Events) mode for remote hosting.
- 🔑 **Automated Token Refresher:** CLI utility to automatically exchange and refresh OAuth 2.0 access tokens.

---

## 🧰 Available MCP Tools

### `publish_linkedin_post`
Publishes a text or media post to your LinkedIn profile.

**Parameters:**
- `commentary` (string, required): The text content of your post.
- `media_urn` (string, optional): LinkedIn digital media asset URN (e.g. `urn:li:digitalmediaAsset:...`).
- `media_title` (string, optional): Title for the attached media.

### `upload_linkedin_media`
Uploads a local image or PDF/document file to LinkedIn and returns the registered media asset URN.

**Parameters:**
- `file_path` (string, required): Local file path of the image or document to upload.
- `recipe` (string, optional): Media recipe type (`urn:li:digitalmediaRecipe:feedshare-image` or `urn:li:digitalmediaRecipe:feedshare-document`).

### `get_linkedin_profile`
Fetches the current authenticated user's LinkedIn profile URN and basic information.

### `get_linkedin_post_analytics`
Retrieves engagement metrics (likes, comments, etc.) for a specific LinkedIn post.

**Parameters:**
- `post_urn` (string, required): The URN or ID of the post.

---

## 🚀 Modes of Operation

### Stdio Mode (Default)
Used by Claude Desktop & Cursor via stdin/stdout:
```bash
./bin/linkmcp
```

### SSE (Server-Sent Events) Mode
Suitable for web services and cloud deployments:
```bash
# Via flags
./bin/linkmcp -mode sse -port 8080

# Or via environment variables
MCP_MODE=sse PORT=8080 ./bin/linkmcp
```

---

## 🔑 Token Refresher CLI

Build and run the `token-refresher` tool to exchange OAuth authorization codes or refresh existing tokens:

```bash
./bin/token-refresher --client-id=YOUR_CLIENT_ID --client-secret=YOUR_CLIENT_SECRET --code=AUTH_CODE --redirect-uri=http://localhost:8080/callback
```

---

## 🚀 Roadmap

- [x] Basic Text Post Publishing (`/rest/posts`)
- [x] Docker Container Support
- [x] Image and PDF/Carousel Upload Support
- [x] Post Analytics Tool (impressions, likes, comments)
- [x] SSE (Server-Sent Events) Mode for Cloud Hosting
- [x] Automated Token Refresher CLI

---

## 📄 License

Distributed under the MIT License. See [LICENSE](LICENSE) for more information.