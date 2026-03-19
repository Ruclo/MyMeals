# MyMeals Backend
Backend for a restaurant dashboard. Customers can order food directly to their table and leave reviews, and employees can see active orders in real time.

**Environment**
- Required variables: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_NAME`, `DB_PASSWORD`, `JWT_SECRET`, `CLOUDINARY_URL`
- Create `MyMeals/.env` with the values from `MyMeals/.env.example`
- For Docker Compose, set `DB_HOST=db` and `DB_PORT=5432`

**Run (Docker Compose)**
1. Ensure `MyMeals/.env` exists with the database values that match `docker-compose.yml`
2. From `C:\Users\vavri\Desktop\chob\mymeals`, run:
```bash
docker compose up --build
```
3. The API is exposed on `http://localhost:8080`

**Run (Local)**
1. Ensure `MyMeals/.env` exists
2. Start Postgres locally
3. From `MyMeals`, run:
```bash
go run ./cmd
```

**Notes**
- Two default users are created at startup: `admin` / `password` and `regular` / `password`
