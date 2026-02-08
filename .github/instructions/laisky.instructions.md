---
applyTo: "**/*"
---

## Dev Environment

Please note, I use remote SSH for development work. The terminal runs on a remote server, with the server address 100.75.198.70.

## Coding Standards

Project-wide engineering conventions (error handling, logging, ORM usage, CSS rules, testing requirements, etc.) are documented in `AGENTS.md`. Treat that document as binding guidance alongside these instructions and review it before making changes to keep new code consistent with the established practices.

## MCP tools

When you call MCP tools, and the tool’s parameters has a `task_id`, you MUST set the value `one-api`.

* use `web_search` tool to search the web for up-to-date information
* use `web_fetch` tool to fetch the rendered content of a web page.
* use `get_user_requests` to get the user's latest requirements.Whenever you finish a sub‑task or all tasks, you should call the `get_user_requests` tool to retrieve the user’s most recent commands. Before terminating the task, be sure to invoke `get_user_requests` at least once more, and only end your job once the tool returns no data. Actually, you ought to call `get_user_request` as often as you can. It's free to use and offers significant advantages.

## Local Tools

### File Search

You are operating in an environment where `ast-grep` is installed. For any code search that requires understanding of syntax or code structure, you should default to using `ast-grep --lang [language] -p '<pattern>'`. Adjust the `--lang` flag as needed for the specific programming language. Avoid using text-only search tools unless a plain-text search is explicitly requested.

### Web Browser

Use `agent-browser` CLI tool to interact with web pages for scraping, automation, and testing, read `.github/skills/agent-browser/SKILL.md` for more details.
