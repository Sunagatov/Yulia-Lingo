<div align="center">
  <br>
  <h1>🇬🇧 Yulia Lingo</h1>
  <p><strong>A Telegram English learning bot — study irregular verbs, build your word list, and translate on the go.</strong></p>
  <p>
    <a href="https://t.me/zufarexplained">💬 Community</a> ·
    <a href="https://github.com/Sunagatov/Yulia-Lingo/issues?q=is%3Aopen+label%3A%22good+first+issue%22">🟢 Good First Issues</a> ·
    <a href="https://github.com/Sunagatov/Yulia-Lingo/issues">🐛 Issues</a>
  </p>

  [![License: CC BY-NC 4.0](https://img.shields.io/badge/license-CC%20BY--NC%204.0-lightgrey.svg)](LICENSE)
  [![Docker Pulls](https://img.shields.io/docker/pulls/zufarexplainedit/yulia-lingo-backend.svg)](https://hub.docker.com/r/zufarexplainedit/yulia-lingo-backend/)
  [![GitHub Stars](https://img.shields.io/github/stars/Sunagatov/Yulia-Lingo)](https://github.com/Sunagatov/Yulia-Lingo/stargazers)
  [![GitHub Issues](https://img.shields.io/github/issues/Sunagatov/Yulia-Lingo)](https://github.com/Sunagatov/Yulia-Lingo/issues)
</div>

---

## 🚀 Quick Start

**📋 Prerequisites:** Go 1.21+, PostgreSQL 17+, Docker Desktop, Telegram Bot Token (from [@BotFather](https://t.me/BotFather))

```bash
# 1. 📥 Clone
git clone https://github.com/Sunagatov/Yulia-Lingo.git && cd Yulia-Lingo

# 2. 🔧 Fill in your credentials
# edit TELEGRAM_BOT_TOKEN and POSTGRESQL_PASSWORD in .env
```

---

### Option A — Local Go + infra in Docker *(recommended for development)*

```bash
# Start only PostgreSQL
docker compose up -d postgres
```

Then run the app from the terminal:

```bash
go run cmd/app/main.go
```

---

### Option B — Everything in Docker

```bash
docker compose up -d --build
```

**Production:**
```bash
# Fill in .env.prod, then:
docker compose -f docker-compose.prod.yml up -d --build
```

---

**🧪 Run the tests:**
```bash
go test ./...
```

---

## 🤔 What is this?

Yulia Lingo is a Telegram bot that helps you learn English interactively. Practice irregular verbs with quizzes, build a personal vocabulary list, and get instant word translations — all without leaving Telegram.

---

## 🛠️ Tech Stack

| 📂 Category | 🔧 Technology |
|---|---|
| 💻 Language | Go 1.21 |
| 🗄️ Database | PostgreSQL 17 |
| 🤖 Telegram | go-telegram-bot-api v5 |
| 📝 Logging | Logrus (structured JSON) |
| 🚢 Deployment | Docker (multi-stage build) |

---

## ✨ Features

- 📚 **Irregular verbs** — interactive quiz to practice all three verb forms
- 📝 **Personal word list** — save, browse, and remove your own vocabulary
- 🔍 **Translation** — instant word translation with multiple meanings
- 📄 **Pagination** — smooth browsing through large word sets
- 🔒 **Secure** — parameterized queries, input sanitization, SSRF protection
- 🛑 **Graceful shutdown** — clean resource release on SIGTERM

---

## 🤖 Commands

| 🎯 Command | 📝 Description |
|---|---|
| `/start` | Welcome message and main menu |
| `/irregular_verbs` | Start irregular verbs practice |
| `/my_word_list` | View and manage your word list |
| `/translate` | Translate a word |
| `/cancel` | Cancel the current operation |

---

## 📁 Project Structure

```
cmd/app/                 # Application entry point
internal/
├── config/             # Environment-based configuration
├── database/           # PostgreSQL connection and pooling
├── logger/             # Structured logging
├── telegram/           # Bot setup and update routing
│   └── handlers/       # Command and callback handlers
├── irregular_verbs/    # Irregular verbs domain
├── my_word_list/       # Personal word list domain
├── translate/          # Translation service
└── util/               # Shared utilities
```

---

## ⚙️ Environment Variables

| Variable | Required | Description |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | ✅ | Token from @BotFather |
| `POSTGRESQL_PASSWORD` | ✅ | Database password |
| `POSTGRESQL_HOST` | ❌ | Defaults to `localhost` |
| `POSTGRESQL_PORT` | ❌ | Defaults to `5432` |
| `POSTGRESQL_USER` | ❌ | Defaults to `postgres` |
| `POSTGRESQL_DATABASE_NAME` | ❌ | Defaults to `yulia_lingo` |
| `LOG_LEVEL` | ❌ | `debug`, `info`, `warn`, `error` |
| `IRREGULAR_VERBS_FILE_PATH` | ❌ | Path to irregular verbs Excel file |

See `.env` for local defaults and `.env.prod` for the production template.

---

## 🤝 Contributing

🎉 Contributions are welcome.

| 🎯 Situation | 🚀 Action |
|---|---|
| 🐛 Found a bug | [Open an issue](https://github.com/Sunagatov/Yulia-Lingo/issues/new) with the `bug` label |
| 💡 Want a feature | Start a [Discussion](https://github.com/Sunagatov/Yulia-Lingo/discussions) first |
| 👨‍💻 Ready to code | Pick a [`good first issue`](https://github.com/Sunagatov/Yulia-Lingo/issues?q=is%3Aopen+label%3A%22good+first+issue%22), comment "I'm on it" |
| 🔧 Big change | Comment on the issue before writing code — tickets may have hidden constraints |

---

## 📄 License

📜 [CC BY-NC 4.0](LICENSE) — free for educational and personal use with author attribution. Commercial use requires explicit written permission from the author ([zufar.sunagatov@gmail.com](mailto:zufar.sunagatov@gmail.com)).

---

## 📞 Contact

- 💬 **Telegram community:** [Zufar Explained IT](https://t.me/zufarexplained)
- 👤 **Personal Telegram:** [@lucky_1uck](https://web.telegram.org/k/#@lucky_1uck)
- 📧 **Email:** [zufar.sunagatov@gmail.com](mailto:zufar.sunagatov@gmail.com)
- 🐛 **Issues:** [GitHub Issues](https://github.com/Sunagatov/Yulia-Lingo/issues)

❤️
