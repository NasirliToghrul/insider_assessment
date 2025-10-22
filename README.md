# Insider Case Study – Message Scheduler & Sender Service

A Go service that:
- Stores messages in **PostgreSQL**
- Periodically sends **pending** messages to an external **webhook**
- Provides REST endpoints to **start/stop** the scheduler and **manually send** messages
- Caches sent message metadata in **Redis**
- Exposes a **Swagger/OpenAPI** spec at `/swagger.yaml`

---
## 🧩 Architecture Overview

```
cmd/main.go                 → Application bootstrap
internal/api                → REST router, handlers, swagger
internal/config             → Environment-based configuration
internal/database           → GORM DB connection, repository
internal/models             → Message model & statuses
internal/sender             → Scheduler, webhook sender, Redis cache
internal/docs/swagger.yaml  → OpenAPI 3.0.3 specification
```

---

## ⚙️ Configuration (.env)

All configuration is done via environment variables. Example:

```env
HTTP_PORT=8080

# PostgreSQL
DB_DRIVER=postgres
DB_DSN=host=postgres user=postgres password=postgres dbname=insider port=5432 sslmode=disable

# Webhook endpoint
WEBHOOK_URL=https://webhook.site/ac992970-67e2-4d79-9ef3-5248b62a7fc9

# Scheduler
TICKER_SECONDS=120
BATCH_SIZE=2
MESSAGE_CHAR_LIMIT=160

# Redis
REDIS_ENABLED=true
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
```
---

## 🌐 API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/start` | Start the scheduler |
| `POST` | `/api/stop` | Stop the scheduler |
| `POST` | `/api/sent-messages` | Send message manually |
| `GET` | `/swagger.yaml` | Download OpenAPI specification |

### Example Request
```json
{
  "to": "+905553333333",
  "content": "Insider - Project D"
}
```

### Example Response
```json
{
  "message": "Accepted",
  "messageId": "0732d23b-c629-4aca-b94f-ca6e9abb6cfd"
}
```

---

## 🐳 Docker Usage

### Build Image
```bash
docker compose up --build
```

### Run Container
```bash
docker run --env-file .env -p 8080:8080 case-study:latest
```
---
## 🧪 Testing

### 🧭 1. Postman Test (Using Webhook.site)

Webhook URL:  
**`https://webhook.site/ac992970-67e2-4d79-9ef3-5248b62a7fc9`**

#### Step 1 — Configure webhook.site
1. Go to your bin page.
2. Click **Edit Response**.
3. Set:
    - **Status:** `202`
    - **Content-Type:** `application/json`
    - **Body:**
      ```json
      {
        "message": "Accepted",
        "messageId": "0732d23b-c629-4aca-b94f-ca6e9abb6cfd"
      }
      ```

#### Step 2 — Postman Request
- **Method:** `POST`
- **URL:** `http://localhost:8080/api/sent-messages`
- **Headers:** `Content-Type: application/json`
- **Body (raw/JSON):**
  ```json
  {
    "to": "+905553333333",
    "content": "Insider - Project D"
  }
  ```

✅ **Expected Result:**
```json
{
  "message": "Accepted",
  "messageId": "0732d23b-c629-4aca-b94f-ca6e9abb6cfd"
}
```

---

### 🧮 2. Database Import Test (Scheduler)

To test the background scheduler, insert a few pending messages:

**Open a new terminal and enter the Postgres container:**
```bash
docker exec -it $(docker ps -qf "name=postgres") psql -U postgres -d insider
```

**Inside the psql prompt:**
```sql
INSERT INTO messages ("to", content, status, created_at, updated_at)
VALUES
  ('+905550000001', 'Insider - Project #1', 'pending', NOW(), NOW()),
  ('+905550000002', 'Insider - Project #2', 'pending', NOW(), NOW());
```
**Exit Postgres:**
```bash
\q
exit
```

The scheduler (every `TICKER_SECONDS=120`) will:
- Claim messages → `processing`
- Send to your webhook
- On success → mark `sent`, set `sent_at` & `remote_message_id`
- On failure → mark `failed`, store `last_error`

**Verify in DB:**
```bash
docker compose exec -T postgres psql -U postgres -d insider   -c "SELECT id, to, status, sent_at, remote_message_id, last_error FROM messages ORDER BY id DESC LIMIT 10;"
```
---

## 🧯 Troubleshooting

| Issue | Possible Fix |
|-------|---------------|
| **405 Method Not Allowed** | Ensure `/api/sent-messages` is POST and implemented |
| **DB connection refused** | Check port mapping (`5434:5432`) and DSN |
| **Webhook error / decode error** | Make sure webhook.site returns `202` + JSON with `messageId` |
| **Redis timeout** | Disable Redis (`REDIS_ENABLED=false`) for debugging |

---

## 👤 Author

**Toghrul Nasirli**  
Software Developer 