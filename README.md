# Домашнє завдання: Тема 8 — Інтеграція БД (PostgreSQL + GORM з AI)
## Наскрізний проєкт: LibraTrack (Система обліку бібліотеки)

**Часовий ліміт:** до 2 годин (завдяки використанню GenAI інструментів для прискорення розробки).
**Мета роботи:** Перевести проєкт **LibraTrack** зі спрощеного in-memory сховища (з Теми 7) на реальну реляційну базу даних PostgreSQL за допомогою ORM GORM та міграцій. Виявити й виправити класичну проблему N+1, а також закріпити навички AI-асистованої розробки.

---

### ЧАСТИНА 1: Промпт-інжиніринг (Проєктування БД та Генерація міграцій) [~35 хв]
*Цей етап виконується за допомогою LLM (ChatGPT / Claude / Copilot Chat).*

#### Контекст та завдання:
У Темі 7 сутність книги (`Book`) містила автора та категорію у вигляді простих текстових полів. Тепер нам потрібно нормалізувати схему до **3NF (третьої нормальної форми)**:
1. Створити окрему таблицю `authors` (id, full_name, bio).
2. Створити окрему таблицю `categories` (id, name).
3. Зв'язати `books` та `categories` як **One-to-Many** (у однієї категорії багато книг).
4. Зв'язати `books` та `authors` як **Many-to-Many** через проміжну (join) таблицю `book_authors`.

#### Ваші дії:
1. **Написання промпту:** Складіть промпт до вашого AI-асистента для генерації файлів міграції SQL для інструменту `golang-migrate`.
   * **Вимога до промпту:** Передайте ваші Go-структури моделей або детальний опис зв'язків. Попросіть згенерувати два файли: `000002_normalize_books.up.sql` та `000002_normalize_books.down.sql`.
   * **Обов'язково вкажіть в промпті:** генерацію явних обмежень зовнішніх ключів (Foreign Keys), поведінку `ON DELETE RESTRICT` (для категорій і авторів) або `ON DELETE CASCADE` (для таблиці зв'язків `book_authors`), а також індекси на поля зовнішніх ключів для оптимізації швидкодії.
2. **Аналіз результату:** Перевірте згенерований AI `.down.sql` файл. Переконайтеся, що він правильно видаляє всі створені таблиці, індекси та обмеження у зворотному порядку.
3. **Запуск міграцій:** Створіть файли міграцій локально та виконайте команду `migrate -database $DATABASE_URL -path migrations up`.

---

### ЧАСТИНА 2: Використання AI-агента (Моделі GORM та Репозиторій) [~45 хв]
*Цей етап виконується за допомогою AI-агента/асистента (наприклад, Cursor, GitHub Copilot, GoLand AI або ChatGPT).*

#### Контекст та завдання:
Необхідно оновити Go-моделі та замінити in-memory репозиторій на роботу з GORM, **не змінюючи при цьому шар хендлерів та сервісів (інтерфейс `BookRepository` має залишитися незмінним)**.

#### Ваші дії:
1. **Оновлення моделей:** Попросіть AI-агента оновити Go-структури у файлі `internal/model/book.go` відповідно до нової схеми. Структура `Book` повинна містити зв'язки:
   * `Category   Category `gorm:"foreignKey:CategoryID"``
   * `Authors    []Author `gorm:"many2many:book_authors;"``
2. **Генерація репозиторію:** Використовуючи AI-агент, згенеруйте код для нового репозиторію `gormBookRepository` у файлі `internal/repository/book_gorm.go`.
   * **Вимога до промпту/агента:** "Реалізуй інтерфейс `BookRepository` з Теми 7 за допомогою GORM. Використовуй `db *gorm.DB`. Обов'язково прокидай контекст у кожен запит через `.WithContext(ctx)`. Напиши методи: `Create`, `FindByID`, `Update`, `Delete`, `List`".
3. **Конфігурація DB:** Налаштуйте підключення до PostgreSQL в `cmd/api/main.go` через Docker Compose. Переконайтеся, що параметри пулу з'єднань (`SetMaxOpenConns`, `SetMaxIdleConns`) конфігуруються відповідно до найкращих практик.

---

### ЧАСТИНА 3: Ручне виконання (Верифікація, Діагностика N+1 та Фікс) [~40 хв]
*Цей етап виконується виключно вами вручну для глибокого розуміння роботи ORM.*

#### Контекст та завдання:
Популярна проблема ORM — N+1 запит, коли завантаження списку сутностей супроводжується окремим SQL-запитом для кожної асоціації в циклі.

#### Ваші дії:
1. **Підготовка середовища:** Увімкніть логування всіх SQL-запитів у GORM за допомогою `logger.Info`.
2. **Запуск та тестування:** Запустіть застосунок та базу даних через `docker-compose up -d`. Наповніть базу тестовими даними (мінімум 10 книг, 5 авторів, 3 категорії).
3. **Аналіз логів (Виявлення N+1):**
   * Виконайте HTTP-запит `GET /api/v1/books`.
   * Подивіться в консоль логів вашого Go-додатка. Ви побачите, що для кожної книги робиться додатковий запит для отримання авторів та категорій.
   * Зафіксуйте кількість запитів (наприклад: 1 загальний запит + N запитів до `authors` + N запитів до `categories`).
4. **Мануальний фікс:**
   * Самостійно (без допомоги AI) змініть метод `List` у вашому Gorm-репозиторії, додавши завантаження зв'язків через `.Preload("Authors").Preload("Category")`.
5. **Повторний замір:** Знову виконайте запит `GET /api/v1/books` і перевірте лог терміналу. Переконайтеся, що кількість запитів скоротилася до константи (2-3 запити сумарно, незалежно від кількості книг).

---

### Вимоги до здачі домашнього завдання (Deliverables):
Студент має надати посилання на свій репозиторій, де обов'язково є:
1. Створені `.sql` файли міграцій (`golang-migrate`).
2. Оновлений код моделей та GORM-репозиторію.
3. Файл **`README.md` з розділом "Lesson 8 Report"**, який містить:
   * Використаний промпт для генерації міграцій з Частини 1.
   * **Докази виправлення N+1:** фрагмент логів терміналу "ДО" (із повторюваними SQL-запитами) та логів "ПІСЛЯ" (де видно роботу `Preload` та батчевий `SELECT ... WHERE id IN (...)`).
   * Різницю в локальній затримці (latency) у мілісекундах для обох варіантів.

---
---

# Homework Assignment: Lesson 8 — DB Integration (PostgreSQL + GORM with AI)
## Capstone Project: LibraTrack (Library Management System)

**Time Budget:** Up to 2 hours (engineered for speed and efficiency using GenAI tools).
**Objective:** Migrate the **LibraTrack** project from the simplified in-memory storage (developed in Lesson 7) to a durable PostgreSQL database using the GORM ORM and schema migrations. Identify and resolve a classic N+1 query problem, and strengthen your AI-assisted development skills.

---

### PART 1: Prompt Engineering (DB Schema Design & Migration Generation) [~35 mins]
*This stage is completed using an LLM (ChatGPT / Claude / Copilot Chat).*

#### Context and Task:
In Lesson 7, the `Book` entity kept author and category as basic string fields. We now need to normalize this database schema to **3rd Normal Form (3NF)**:
1. Create a separate `authors` table (id, full_name, bio).
2. Create a separate `categories` table (id, name).
3. Associate `books` and `categories` as a **One-to-Many** relationship (one category has many books).
4. Associate `books` and `authors` as a **Many-to-Many** relationship via a join table named `book_authors`.

#### Steps to follow:
1. **Write the Prompt:** Construct a detailed prompt for your AI assistant to generate the standard Up and Down SQL migration files for `golang-migrate`.
   * **Prompt Requirements:** Provide your model structs or a precise description of the relationships. Ask for `000002_normalize_books.up.sql` and `000002_normalize_books.down.sql`.
   * **Mandatory Specifications:** Instruct the AI to include explicit foreign key constraints, define proper behavior on delete (`ON DELETE RESTRICT` for categories/authors, `ON DELETE CASCADE` for the `book_authors` join table), and indexes on foreign key columns.
2. **Analyze the Output:** Critically review the AI-generated `.down.sql` file. Verify that it correctly drops all tables, indexes, and constraints in the exact reverse order.
3. **Execute Migrations:** Create the files locally and run the migrations command: `migrate -database $DATABASE_URL -path migrations up`.

---

### PART 2: AI Agent Usage (GORM Models & Repository Implementation) [~45 mins]
*This stage is completed using an AI agent/coding assistant (Cursor, GitHub Copilot, GoLand AI, or ChatGPT).*

#### Context and Task:
You must update your Go models and replace the in-memory repository implementation with GORM **without modifying the handler or service layers (the `BookRepository` interface must remain completely unchanged)**.

#### Steps to follow:
1. **Update Models:** Use your AI agent to update the Go structs in `internal/model/book.go`. The `Book` struct must define associations:
   * `Category   Category `gorm:"foreignKey:CategoryID"``
   * `Authors    []Author `gorm:"many2many:book_authors;"``
2. **Generate the Repository:** Prompt the AI agent to write the new `gormBookRepository` in `internal/repository/book_gorm.go`.
   * **Agent/Prompt Guidance:** "Implement the `BookRepository` interface from Lesson 7 using GORM. Use `db *gorm.DB`. Ensure you propagate the Go context in every single database call using `.WithContext(ctx)`. Implement `Create`, `FindByID`, `Update`, `Delete`, and `List`."
3. **DB Configuration:** Establish the PostgreSQL connection in `cmd/api/main.go` using the Docker Compose setup. Set appropriate connection pool parameters (`SetMaxOpenConns`, `SetMaxIdleConns`) following industry best practices.

---

### PART 3: Manual Execution (Verification, N+1 Diagnostics & Fixing) [~40 mins]
*This stage must be performed entirely by you manually to ensure a solid, fundamental understanding of ORMs.*

#### Context and Task:
A notorious ORM pitfall is the N+1 query problem, where loading a list of entities triggers separate SQL queries for each individual row's association in a loop.

#### Steps to follow:
1. **Prepare Environment:** Enable verbose SQL logging in GORM using `logger.Info` in your DB configuration config.
2. **Run and Seed:** Start your environment via `docker-compose up -d`. Seed the database with sample data (at least 10 books, 5 authors, and 3 categories).
3. **Analyze Logs (Diagnose N+1):**
   * Trigger the `GET /api/v1/books` endpoint.
   * Inspect your Go API terminal logs. Identify how many SQL queries are fired to retrieve the list.
   * Document this pattern (e.g., 1 main query + N author queries + N category queries).
4. **Manual Fix:**
   * Personally (without AI assistance) edit the `List` method in your GORM repository to pre-load associations via GORM's built-in eager loading: `.Preload("Authors").Preload("Category")`.
5. **Re-Test & Benchmark:** Trigger `GET /api/v1/books` again. Verify the log output. Ensure that the total query count is now a flat constant (2-3 queries total, regardless of the books count).

---

### Homework Deliverables:
Submit a link to your repository containing:
1. The `.sql` up/down migration files.
2. The updated model structures and the fully integrated GORM repository.
3. A **`README.md` file with a "Lesson 8 Report"** section that includes:
   * The prompt you used to generate the migrations in Part 1.
   * **N+1 Proof:** Plain-text terminal query logs "BEFORE" the fix (highlighting repeating SQL selects) and "AFTER" the fix (highlighting the batched `SELECT ... WHERE id IN (...)` queries).
   * Local endpoint latency measurements in milliseconds (Before vs After).

---

## Lesson 8 Report

### Як запустити

```bash
docker compose up -d --wait     # PostgreSQL 16 на localhost:5434
cp .env.example .env            # змінні середовища
```

### Частина 1: промпт для генерації міграцій

````text
You are a PostgreSQL 16 and golang-migrate expert. Generate SQL migration files for the LibraTrack library management system.

## Output
Four files, each in its own code block with the file name above it:
- migrations/000001_init_schema.up.sql
- migrations/000001_init_schema.down.sql
- migrations/000002_seed_demo_data.up.sql
- migrations/000002_seed_demo_data.down.sql
Use English only, including SQL comments. Do not use any Go code or GORM AutoMigrate: the schema must be created only by SQL.

## Context
The previous version stored author and category as plain string fields on the book. Normalize the schema to 3NF.

## Tables (000001_init_schema)
- categories(id, name)
- authors(id, full_name, bio)
- books(id, title, isbn, category_id, publisher, published_year, description, created_at)
- book_authors(author_id, book_id): join table for Many-to-Many between books and authors
- copies(id, book_id, inventory_no, status)
- users(id, email, password_hash, full_name, role, created_at)

## Requirements
Keys and types:
- id columns: BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY.
- book_authors has a composite PRIMARY KEY (author_id, book_id).
- created_at columns: TIMESTAMPTZ NOT NULL DEFAULT now().
- Required fields are NOT NULL: title, isbn, category_id, full_name, name, inventory_no, email, password_hash, role, status.
- Optional fields stay nullable (no default empty strings): bio, publisher, published_year, description.

Constraints:
- UNIQUE: books.isbn, categories.name, copies.inventory_no, users.email.
- CHECK on copies.status: available, borrowed, lost, under_repair. Default 'available'.
- CHECK on users.role: admin, librarian, reader. Default 'reader'.
- Add a CHECK that published_year is reasonable when not null.

Foreign keys (explicit named constraints):
- books.category_id -> categories(id) ON DELETE RESTRICT
- book_authors.author_id -> authors(id) ON DELETE RESTRICT
- book_authors.book_id -> books(id) ON DELETE CASCADE
- copies.book_id -> books(id) ON DELETE CASCADE
Deleting a book must remove its copies and book_authors rows automatically.

Indexes (for filters, pagination and to avoid slow N+1-style lookups):
- books(category_id)
- book_authors(author_id)
- book_authors(book_id)
- copies(book_id)
- ordering by id is covered by the primary key, no extra index needed.
Use explicit index names, for example idx_books_category_id.

## Seed data (000002_seed_demo_data)
- At least 3 categories, 6 authors, 10 books, 24 copies.
- At least one book has 2 authors and at least one author has 2 books.
- Copies have a mix of statuses.
- Do NOT hardcode ids. Use subselects by unique natural keys, for example
  (SELECT id FROM categories WHERE name = 'Programming') and (SELECT id FROM books WHERE isbn = '...').
  This avoids identity sequence conflicts.
- Do not insert users with real passwords: either insert none or use an obviously fake hash.

## Down migrations
- Drop everything created by the up file in the exact reverse dependency order: copies, book_authors, books, users, authors, categories. Drop indexes and constraints explicitly where applicable.
- The seed down file must delete only the rows inserted by the seed (by natural keys), in reverse dependency order.
- Use IF EXISTS so a down migration is safe to re-run.

## Self-check
After the files, add a short checklist confirming: each up has a matching down; FK behavior matches the list above; the database rejects an invalid copies.status, an invalid users.role, a duplicate ISBN, and a book with a non-existent category_id.
````

### Частина 3: N+1

**ДО виправлення** (лог SQL):

```text
$ set -a; . ./.env; set +a; go run ./cmd/api
2026/10/12 15:51:10 listening on :8080
2026/10/12 15:51:37 GET "/swagger/index.html" -> 200 (0s)
2026/10/12 15:51:38 GET "/openapi.yaml" -> 200 (0s)
2026/10/12 15:51:38 GET "/favicon.ico" -> 404 (0s)
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:149
[20.753ms] [rows:10] SELECT * FROM "books" ORDER BY books.id ASC LIMIT 10
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[1.703ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[1.539ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 1 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[1.518ms] [rows:2] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 2
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[1.212ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 1 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[1.524ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 3
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[1.052ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 1 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[1.063ms] [rows:2] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 4
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[1.114ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 1 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[1.069ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 5
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[1.409ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 2 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[0.520ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 6
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[0.506ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 2 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[2.381ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 7
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[1.122ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 2 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[1.245ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 8
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[1.628ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 2 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[0.561ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 9
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[0.693ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 3 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:156
[1.502ms] [rows:1] SELECT "authors"."id","authors"."full_name","authors"."bio" FROM "authors" JOIN "book_authors" ON "book_authors"."author_id" = "authors"."id" AND "book_authors"."book_id" = 10
2026/10/12 15:58:53 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:160
[0.677ms] [rows:1] SELECT * FROM "categories" WHERE "categories"."id" = 3 ORDER BY "categories"."id" LIMIT 1
2026/10/12 15:58:53 GET "/api/v1/books" -> 200 (54.9367ms)
```

**ПІСЛЯ виправлення** (`Preload`):
```text
Vadym@DESKTOP-FO5I1GK MINGW64 /c/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8 (main)
$ set -a; . ./.env; set +a; go run ./cmd/api
2026/10/12 16:17:43 listening on :8080
2026/10/12 16:18:56 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:149
[5.112ms] [rows:12] SELECT * FROM "book_authors" WHERE "book_authors"."book_id" IN (1,2,3,4,5,6,7,8,9,10)
2026/10/12 16:18:56 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:149
[2.647ms] [rows:10] SELECT * FROM "authors" WHERE "authors"."id" IN (1,2,3,4,5,6,7,8,9,10)
2026/10/12 16:18:56 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:149
[1.143ms] [rows:3] SELECT * FROM "categories" WHERE "categories"."id" IN (1,2,3)
2026/10/12 16:18:56 C:/workSpace/project/softservienc/lesson-8/Go-with-GenAI-lessons-8/internal/repository/book_repository.go:149
[15.258ms] [rows:10] SELECT * FROM "books" ORDER BY books.id ASC LIMIT 10
2026/10/12 16:18:56 GET "/api/v1/books" -> 200 (15.2577ms)
```

**Затримка `GET /api/v1/books?limit=100`** (10 книжок у базі, локальний Postgres у Docker, curl):
один прогрівальний запит, далі 5 вимірювань, `time_total` у мілісекундах.

| | Запитів SELECT | Затримка по 5 запитах, мс | Медіана, мс |
|---|---|---|---|
| ДО (цикл у `List`, N+1) | 21 (1 + 10 авторів + 10 категорій) | 10.1, 23.5, 10.1, 13.8, 13.4 | 13.4 |
| ПІСЛЯ (`Preload("Authors").Preload("Category")`) | 4 (books, book_authors, authors, categories) | 9.2, 4.4, 4.9, 3.8, 3.4 | 4.4 |

Кількість запитів для списку тепер стала (4) і не залежить від числа книжок, а раніше росла як 1 + 2N.
Перший запит одразу після старту повільніший за решту (холодний пул з'єднань): у логу «ДО» це 54.9 мс, у логу «ПІСЛЯ» 15.3 мс.
