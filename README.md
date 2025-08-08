# SHCalendar — Simple Habits Calendar

A tiny single-page app to mark habit completion on a yearly calendar. Backend: Go + SQLite (pure Go driver). Frontend is embedded HTML.

## Run locally

```bash
# Go 1.23+
go run ./
```

Environment variables:
- PORT (default 8085)
- DB_PATH (default calendar.db)
- HABITS_FILE (default habits.txt)

Open http://localhost:8085

## Docker

Build and run:

```bash
docker build -t shcalendar:latest .
# persistent data under named volume
Docker volume "shcalendar_data" will be created by compose

docker compose up --build
```

Compose binds port 8085 and mounts a volume at /data; the app uses DB_PATH=/data/calendar.db.

## Habits list

`habits.txt` is a simple CSV with integer id and display name:

```
1,Пить воду
2,Читать 20 минут
...
```

IDs are integers now. The DB stores habit ids as INTEGER, date as INTEGER (Unix epoch seconds at midnight UTC).

## Graceful shutdown & Health

- The server uses timeouts and graceful shutdown on SIGINT/SIGTERM.
- Health endpoint: `GET /healthz` returns 200 when DB is reachable.

## Notes

- Security headers and gzip compression are enabled.
- For a single user, no special concurrency handling is added.
