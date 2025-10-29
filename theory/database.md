# Работа с базами данных в Go: Полная теория

## Введение в database/sql

### Что такое database/sql?

Пакет `database/sql` в Go предоставляет **универсальный интерфейс** для работы с различными СУБД:
- **PostgreSQL**, MySQL, SQLite
- **Microsoft SQL Server**, Oracle
- **Другие** через драйверы

### Архитектура database/sql

```
Приложение Go
    ↓
database/sql (универсальный интерфейс)
    ↓
Драйверы (github.com/lib/pq, go-sql-driver/mysql, и т.д.)
    ↓
СУБД (PostgreSQL, MySQL, и т.д.)
```

## Подключение к базе данных

### Импорт драйверов

```go
import (
    "database/sql"
    _ "github.com/lib/pq"        // PostgreSQL
    _ "github.com/go-sql-driver/mysql" // MySQL
    _ "github.com/mattn/go-sqlite3"    // SQLite
)
```

### Открытие соединения

```go
func connectToPostgreSQL() (*sql.DB, error) {
    // Строка подключения для PostgreSQL
    connStr := "user=username dbname=mydb sslmode=disable"
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("ошибка подключения: %w", err)
    }
    
    // Проверяем подключение
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("ошибка пинга: %w", err)
    }
    
    return db, nil
}

func connectToMySQL() (*sql.DB, error) {
    // Строка подключения для MySQL
    connStr := "user:password@tcp(localhost:3306)/dbname"
    
    db, err := sql.Open("mysql", connStr)
    if err != nil {
        return nil, fmt.Errorf("ошибка подключения: %w", err)
    }
    
    return db, nil
}

func connectToSQLite() (*sql.DB, error) {
    // Строка подключения для SQLite
    db, err := sql.Open("sqlite3", "./test.db")
    if err != nil {
        return nil, fmt.Errorf("ошибка подключения: %w", err)
    }
    
    return db, nil
}
```

## Пулы соединений

### Настройка пула соединений

```go
func setupConnectionPool(db *sql.DB) {
    // Максимальное количество открытых соединений
    db.SetMaxOpenConns(25)
    
    // Максимальное количество простаивающих соединений
    db.SetMaxIdleConns(25)
    
    // Максимальное время жизни соединения
    db.SetConnMaxLifetime(5 * time.Minute)
    
    // Максимальное время простоя соединения
    db.SetConnMaxIdleTime(5 * time.Minute)
}
```

### Мониторинг пула соединений

```go
func printConnectionStats(db *sql.DB) {
    stats := db.Stats()
    
    fmt.Printf("Открыто соединений: %d\n", stats.OpenConnections)
    fmt.Printf("Занято: %d\n", stats.InUse)
    fmt.Printf("Свободно: %d\n", stats.Idle)
    fmt.Printf("Всего открыто: %d\n", stats.TotalOpenConnections)
}
```

## Выполнение запросов

### Запросы без возвращаемых данных

```go
func insertUser(db *sql.DB, name string, email string) error {
    query := "INSERT INTO users (name, email) VALUES ($1, $2)"
    
    result, err := db.Exec(query, name, email)
    if err != nil {
        return fmt.Errorf("ошибка вставки: %w", err)
    }
    
    // Получаем количество затронутых строк
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("ошибка получения количества строк: %w", err)
    }
    
    fmt.Printf("Вставлено %d строк\n", rowsAffected)
    return nil
}

func updateUser(db *sql.DB, id int, name string) error {
    query := "UPDATE users SET name = $1 WHERE id = $2"
    
    result, err := db.Exec(query, name, id)
    if err != nil {
        return fmt.Errorf("ошибка обновления: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("ошибка получения количества строк: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("пользователь с ID %d не найден", id)
    }
    
    return nil
}

func deleteUser(db *sql.DB, id int) error {
    query := "DELETE FROM users WHERE id = $1"
    
    result, err := db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("ошибка удаления: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("ошибка получения количества строк: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("пользователь с ID %d не найден", id)
    }
    
    return nil
}
```

### Запросы с одной строкой результата

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func getUserByID(db *sql.DB, id int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    row := db.QueryRow(query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("пользователь с ID %d не найден", id)
        }
        return nil, fmt.Errorf("ошибка сканирования: %w", err)
    }
    
    return &user, nil
}

func getUserByEmail(db *sql.DB, email string) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE email = $1"
    
    row := db.QueryRow(query, email)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("пользователь с email %s не найден", email)
        }
        return nil, fmt.Errorf("ошибка сканирования: %w", err)
    }
    
    return &user, nil
}
```

### Запросы с множеством строк

```go
func getAllUsers(db *sql.DB) ([]User, error) {
    query := "SELECT id, name, email FROM users ORDER BY id"
    
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса: %w", err)
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, &user.Email)
        if err != nil {
            return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
        }
        users = append(users, user)
    }
    
    // Проверяем ошибки итерации
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("ошибка итерации: %w", err)
    }
    
    return users, nil
}

func getUsersByAgeRange(db *sql.DB, minAge, maxAge int) ([]User, error) {
    query := "SELECT id, name, email FROM users WHERE age BETWEEN $1 AND $2"
    
    rows, err := db.Query(query, minAge, maxAge)
    if err != nil {
        return nil, fmt.Errorf("ошибка запроса: %w", err)
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, &user.Email)
        if err != nil {
            return nil, fmt.Errorf("ошибка сканирования строки: %w", err)
        }
        users = append(users, user)
    }
    
    return users, nil
}
```

## Транзакции

### Основы транзакций

```go
func transferMoney(db *sql.DB, fromID, toID int, amount float64) error {
    // Начинаем транзакцию
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("ошибка начала транзакции: %w", err)
    }
    
    // Откладываем откат на случай ошибки
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    // Списываем деньги
    _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
    if err != nil {
        return fmt.Errorf("ошибка списания: %w", err)
    }
    
    // Зачисляем деньги
    _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
    if err != nil {
        return fmt.Errorf("ошибка зачисления: %w", err)
    }
    
    // Подтверждаем транзакцию
    err = tx.Commit()
    if err != nil {
        return fmt.Errorf("ошибка подтверждения транзакции: %w", err)
    }
    
    return nil
}
```

### Транзакции с контекстом

```go
func transferMoneyWithContext(ctx context.Context, db *sql.DB, fromID, toID int, amount float64) error {
    // Начинаем транзакцию с контекстом
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("ошибка начала транзакции: %w", err)
    }
    
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    // Списываем деньги
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
    if err != nil {
        return fmt.Errorf("ошибка списания: %w", err)
    }
    
    // Зачисляем деньги
    _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
    if err != nil {
        return fmt.Errorf("ошибка зачисления: %w", err)
    }
    
    // Подтверждаем транзакцию
    err = tx.Commit()
    if err != nil {
        return fmt.Errorf("ошибка подтверждения транзакции: %w", err)
    }
    
    return nil
}
```

## Подготовленные выражения

### Когда использовать prepared statements?

Prepared statements полезны когда:
- Один запрос выполняется **много раз**
- Нужно **повысить производительность**
- Нужно **улучшить безопасность**

### Пример prepared statements

```go
type UserRepository struct {
    db *sql.DB
    insertStmt *sql.Stmt
    updateStmt *sql.Stmt
    deleteStmt *sql.Stmt
}

func NewUserRepository(db *sql.DB) (*UserRepository, error) {
    // Подготавливаем выражения
    insertStmt, err := db.Prepare("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id")
    if err != nil {
        return nil, fmt.Errorf("ошибка подготовки insert: %w", err)
    }
    
    updateStmt, err := db.Prepare("UPDATE users SET name = $1, email = $2 WHERE id = $3")
    if err != nil {
        return nil, fmt.Errorf("ошибка подготовки update: %w", err)
    }
    
    deleteStmt, err := db.Prepare("DELETE FROM users WHERE id = $1")
    if err != nil {
        return nil, fmt.Errorf("ошибка подготовки delete: %w", err)
    }
    
    return &UserRepository{
        db:         db,
        insertStmt: insertStmt,
        updateStmt: updateStmt,
        deleteStmt: deleteStmt,
    }, nil
}

func (r *UserRepository) CreateUser(user *User) error {
    err := r.insertStmt.QueryRow(user.Name, user.Email).Scan(&user.ID)
    if err != nil {
        return fmt.Errorf("ошибка создания пользователя: %w", err)
    }
    return nil
}

func (r *UserRepository) UpdateUser(user *User) error {
    _, err := r.updateStmt.Exec(user.Name, user.Email, user.ID)
    if err != nil {
        return fmt.Errorf("ошибка обновления пользователя: %w", err)
    }
    return nil
}

func (r *UserRepository) DeleteUser(id int) error {
    _, err := r.deleteStmt.Exec(id)
    if err != nil {
        return fmt.Errorf("ошибка удаления пользователя: %w", err)
    }
    return nil
}

func (r *UserRepository) Close() error {
    // Закрываем подготовленные выражения
    if err := r.insertStmt.Close(); err != nil {
        return err
    }
    if err := r.updateStmt.Close(); err != nil {
        return err
    }
    if err := r.deleteStmt.Close(); err != nil {
        return err
    }
    return nil
}
```

## Работа с NULL значениями

### Использование sql.Null*

```go
type User struct {
    ID        int            `json:"id"`
    Name      string         `json:"name"`
    Email     string         `json:"email"`
    Phone     sql.NullString `json:"phone,omitempty"`
    Age       sql.NullInt64  `json:"age,omitempty"`
    IsActive  sql.NullBool   `json:"is_active,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
}

func (u User) MarshalJSON() ([]byte, error) {
    type Alias User
    aux := struct {
        Alias
        Phone *string `json:"phone,omitempty"`
        Age   *int64  `json:"age,omitempty"`
    }{
        Alias: Alias(u),
    }
    
    if u.Phone.Valid {
        aux.Phone = &u.Phone.String
    }
    
    if u.Age.Valid {
        aux.Age = &u.Age.Int64
    }
    
    return json.Marshal(aux)
}

func getUserWithNulls(db *sql.DB, id int) (*User, error) {
    query := "SELECT id, name, email, phone, age, is_active, created_at FROM users WHERE id = $1"
    
    row := db.QueryRow(query, id)
    
    var user User
    err := row.Scan(
        &user.ID,
        &user.Name,
        &user.Email,
        &user.Phone,
        &user.Age,
        &user.IsActive,
        &user.CreatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("ошибка сканирования: %w", err)
    }
    
    return &user, nil
}
```

## Использование контекста

### Зачем нужен контекст?

Контекст позволяет:
- **Отменять** долгие запросы
- **Устанавливать таймауты**
- **Передавать метаданные**

### Примеры с контекстом

```go
func getUserWithTimeout(ctx context.Context, db *sql.DB, id int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    row := db.QueryRowContext(ctx, query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("пользователь не найден")
        }
        return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
    }
    
    return &user, nil
}

func main() {
    // Создаем контекст с таймаутом 5 секунд
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    user, err := getUserWithTimeout(ctx, db, 1)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            log.Fatal("Запрос превысил таймаут")
        }
        log.Fatal("Ошибка:", err)
    }
    
    fmt.Printf("Пользователь: %+v\n", user)
}
```

## Обработка ошибок

### Типичные ошибки баз данных

```go
import (
    "database/sql"
    "errors"
    "strings"
)

func handleDatabaseErrors(err error) error {
    if err == nil {
        return nil
    }
    
    // Ошибки от database/sql
    if errors.Is(err, sql.ErrNoRows) {
        return fmt.Errorf("запись не найдена: %w", err)
    }
    
    if errors.Is(err, sql.ErrTxDone) {
        return fmt.Errorf("транзакция уже завершена: %w", err)
    }
    
    // Ошибки от драйверов
    errMsg := err.Error()
    
    // Ошибки уникальности
    if strings.Contains(errMsg, "duplicate key") || 
       strings.Contains(errMsg, "UNIQUE constraint failed") {
        return fmt.Errorf("нарушение уникальности: %w", err)
    }
    
    // Ошибки соединения
    if strings.Contains(errMsg, "connection refused") ||
       strings.Contains(errMsg, "no such host") {
        return fmt.Errorf("нет соединения с базой данных: %w", err)
    }
    
    // Ошибки синтаксиса
    if strings.Contains(errMsg, "syntax error") {
        return fmt.Errorf("ошибка синтаксиса SQL: %w", err)
    }
    
    return fmt.Errorf("ошибка базы данных: %w", err)
}
```

### Retry механизм

```go
func executeWithRetry(db *sql.DB, query string, args ...interface{}) error {
    maxRetries := 3
    backoff := time.Second
    
    for i := 0; i < maxRetries; i++ {
        _, err := db.Exec(query, args...)
        if err == nil {
            return nil
        }
        
        // Проверяем, можно ли повторить
        if !isRetryableError(err) {
            return err
        }
        
        // Ждем перед повтором
        time.Sleep(backoff)
        backoff *= 2 // Экспоненциальная задержка
    }
    
    return fmt.Errorf("превышено максимальное количество попыток")
}

func isRetryableError(err error) bool {
    errMsg := err.Error()
    
    // Временные ошибки
    retryableErrors := []string{
        "connection refused",
        "connection reset",
        "timeout",
        "deadlock",
    }
    
    for _, retryable := range retryableErrors {
        if strings.Contains(errMsg, retryable) {
            return true
        }
    }
    
    return false
}
```

## Миграции баз данных

### Простая система миграций

```go
type Migration struct {
    Version int
    Query   string
}

var migrations = []Migration{
    {
        Version: 1,
        Query: `
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    },
    {
        Version: 2,
        Query: `
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    content TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
    },
}

func runMigrations(db *sql.DB) error {
    // Создаем таблицу для отслеживания миграций
    _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS migrations (
    version INTEGER PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)
    if err != nil {
        return fmt.Errorf("ошибка создания таблицы миграций: %w", err)
    }
    
    // Получаем последнюю примененную миграцию
    var lastVersion int
    err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM migrations").Scan(&lastVersion)
    if err != nil {
        return fmt.Errorf("ошибка получения последней миграции: %w", err)
    }
    
    // Применяем новые миграции
    for _, migration := range migrations {
        if migration.Version > lastVersion {
            fmt.Printf("Применяем миграцию v%d...\n", migration.Version)
            
            tx, err := db.Begin()
            if err != nil {
                return fmt.Errorf("ошибка начала транзакции: %w", err)
            }
            
            // Выполняем миграцию
            _, err = tx.Exec(migration.Query)
            if err != nil {
                tx.Rollback()
                return fmt.Errorf("ошибка выполнения миграции v%d: %w", migration.Version, err)
            }
            
            // Записываем информацию о миграции
            _, err = tx.Exec("INSERT INTO migrations (version) VALUES ($1)", migration.Version)
            if err != nil {
                tx.Rollback()
                return fmt.Errorf("ошибка записи миграции v%d: %w", migration.Version, err)
            }
            
            err = tx.Commit()
            if err != nil {
                return fmt.Errorf("ошибка коммита миграции v%d: %w", migration.Version, err)
            }
            
            fmt.Printf("Миграция v%d применена успешно\n", migration.Version)
        }
    }
    
    return nil
}
```

## Распространенные ошибки

### 1. Забытый rows.Close()

```go
// ПЛОХО
rows, _ := db.Query("SELECT * FROM users")
for rows.Next() {
    // ...
}
// Забыли rows.Close() - утечка соединений!

// ХОРОШО
rows, _ := db.Query("SELECT * FROM users")
defer rows.Close() // Автоматическое закрытие
for rows.Next() {
    // ...
}
```

### 2. Необработанные ошибки

```go
// ПЛОХО
db.Exec("INSERT INTO users (name) VALUES ('John')")

// ХОРОШО
_, err := db.Exec("INSERT INTO users (name) VALUES ('John')")
if err != nil {
    log.Printf("Ошибка вставки: %v", err)
    return err
}
```

### 3. SQL инъекции

```go
// ОЧЕНЬ ПЛОХО
name := "John'; DROP TABLE users; --"
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
db.Query(query) // Удалит всю таблицу!

// ХОРОШО
name := "John'; DROP TABLE users; --"
query := "SELECT * FROM users WHERE name = $1"
db.Query(query, name) // Безопасно
```

## Лучшие практики

### 1. Использование репозиториев

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id int) error
    List(ctx context.Context, limit, offset int) ([]User, error)
}

type postgresUserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *User) error {
    query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
    
    err := r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
    if err != nil {
        return fmt.Errorf("ошибка создания пользователя: %w", err)
    }
    
    return nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    row := r.db.QueryRowContext(ctx, query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("пользователь не найден")
        }
        return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
    }
    
    return &user, nil
}
```

### 2. Graceful shutdown

```go
func main() {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        fmt.Println("Закрываем соединение с базой данных...")
        db.Close()
        os.Exit(0)
    }()
    
    // Используем базу данных...
}
```

### 3. Мониторинг и логирование

```go
func loggedExec(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
    start := time.Now()
    
    result, err := db.Exec(query, args...)
    
    duration := time.Since(start)
    log.Printf("Запрос: %s, Аргументы: %v, Время: %v, Ошибка: %v", 
        query, args, duration, err)
    
    return result, err
}
```

## См. также

- [Базы данных для чайников](../concepts/database.md) - базовое объяснение
- [Контекст](../concepts/context.md) - использование контекста с базами данных
- [Тестирование](../concepts/testing.md) - как тестировать код работы с БД
- [Пулы соединений](connection-pooling.md) - подробнее о пулах
- [Транзакции](transactions.md) - подробнее о транзакциях
- [Практические примеры](../examples/database) - примеры кода