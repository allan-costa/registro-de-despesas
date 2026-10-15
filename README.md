**OpenCode**
**qwen2.5-coder:14b**

## Stack

**GNU LINUX**: Debian 13 ou distro derivada
**Backend:** Go 1.22+, Gin v1.9, PostgreSQL (SQL puro com `database/sql` + `lib/pq` — proibido ORM)
**Auth:** JWT v5 (`golang-jwt/jwt/v5`), bcrypt (`x/crypto/bcrypt`)
**Frontend:** HTML5, Tailwind CSS CDN, Chart.js CDN, Vanilla JS ES6
**Outros:** `godotenv`, `google/uuid`