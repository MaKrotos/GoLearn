# Работа с базами данных в Go - объяснение для чайников

## Что такое database/sql?

Представьте, что `database/sql` - это **универсальный пульт управления** для разных баз данных. Независимо от того, используете ли вы MySQL, PostgreSQL или SQLite, вы можете управлять ими с помощью одного и того же интерфейса.

## Основы подключения

### Подключение к базе данных

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/lib/pq" // Драйвер для PostgreSQL
)

func main() {
    // Строка подключения
    connStr := "user=username dbname=mydb sslmode=disable"
    
    // Открываем соединение
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close() // Всегда закрываем соединение
    
    // Проверяем подключение
    if err := db.Ping(); err != nil {
        log.Fatal("Не удалось подключиться к базе данных:", err)
    }
    
    fmt.Println("Успешное подключение к базе данных!")
}
```

### Важные моменты:

1. **sql.Open()** не проверяет подключение - только создает объект
2. **db.Ping()** проверяет реальное подключение
3. **defer db.Close()** закрывает соединение при завершении функции

## Пулы соединений

### Что такое пул соединений?

Представьте пул соединений как **группу такси**, которые ждут заказов:
- Когда нужна база данных, берется свободное "такси"
- После использования "такси" возвращается в пул
- Нет необходимости каждый раз создавать новое соединение

### Настройка пула

```go
func main() {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Настройка пула соединений
    db.SetMaxOpenConns(25)     // Максимум 25 открытых соединений
    db.SetMaxIdleConns(25)     // Максимум 25 простаивающих соединений
    db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни 5 минут
    
    // Используем базу данных...
}
```

## Выполнение запросов

### Запросы без возвращаемых данных (INSERT, UPDATE, DELETE)

```go
func createUser(db *sql.DB, name string, age int) error {
    query := "INSERT INTO users (name, age) VALUES ($1, $2)"
    
    // Выполняем запрос
    _, err := db.Exec(query, name, age)
    if err != nil {
        return fmt.Errorf("ошибка создания пользователя: %v", err)
    }
    
    return nil
}

func updateUser(db *sql.DB, id int, name string) error {
    query := "UPDATE users SET name = $1 WHERE id = $2"
    
    result, err := db.Exec(query, name, id)
    if err != nil {
        return fmt.Errorf("ошибка обновления пользователя: %v", err)
    }
    
    // Проверяем, сколько строк было изменено
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("ошибка получения количества измененных строк: %v", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("пользователь с ID %d не найден", id)
    }
    
    return nil
}
```

### Запросы с одной строкой результата (SELECT одного объекта)

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func getUserByID(db *sql.DB, id int) (*User, error) {
    query := "SELECT id, name, age FROM users WHERE id = $1"
    
    row := db.QueryRow(query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Age)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("пользователь с ID %d не найден", id)
        }
        return nil, fmt.Errorf("ошибка получения пользователя: %v", err)
    }
    
    return &user, nil
}
```

### Запросы с множеством строк (SELECT списка объектов)

```go
func getAllUsers(db *sql.DB) ([]User, error) {
    query := "SELECT id, name, age FROM users ORDER BY id"
    
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("ошибка получения пользователей: %v", err)
    }
    defer rows.Close() // Важно закрыть rows!
    
    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Name, &user.Age)
        if err != nil {
            return nil, fmt.Errorf("ошибка сканирования строки: %v", err)
        }
        users = append(users, user)
    }
    
    // Проверяем ошибки итерации
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("ошибка итерации по строкам: %v", err)
    }
    
    return users, nil
}
```

## Работа с транзакциями

### Что такое транзакция?

Транзакция - это **последовательность операций**, которые выполняются как **единое целое**:
- Или **все** операции выполняются успешно
- Или **ни одна** не выполняется (откат)

### Пример транзакции

```go
func transferMoney(db *sql.DB, fromID, toID int, amount float64) error {
    // Начинаем транзакцию
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("ошибка начала транзакции: %v", err)
    }
    
    // Откладываем откат на случай ошибки
    defer func() {
        if err != nil {
            tx.Rollback() // Откатываем транзакцию
        }
    }()
    
    // Списываем деньги
    _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
    if err != nil {
        return fmt.Errorf("ошибка списания: %v", err)
    }
    
    // Зачисляем деньги
    _, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
    if err != nil {
        return fmt.Errorf("ошибка зачисления: %v", err)
    }
    
    // Подтверждаем транзакцию
    err = tx.Commit()
    if err != nil {
        return fmt.Errorf("ошибка подтверждения транзакции: %v", err)
    }
    
    return nil
}
```

## Использование контекста

### Зачем нужен контекст?

Контекст позволяет:
- **Отменять** долгие запросы
- **Устанавливать таймауты**
- **Передавать метаданные** (например, ID пользователя)

### Примеры с контекстом

```go
func getUserWithTimeout(ctx context.Context, db *sql.DB, id int) (*User, error) {
    query := "SELECT id, name, age FROM users WHERE id = $1"
    
    row := db.QueryRowContext(ctx, query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Age)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("пользователь не найден")
        }
        return nil, fmt.Errorf("ошибка получения пользователя: %v", err)
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

## Подготовленные выражения (Prepared Statements)

### Зачем нужны prepared statements?

Prepared statements полезны когда:
- Один запрос выполняется **много раз**
- Нужно **повысить производительность**
- Нужно **улучшить безопасность**

### Пример

```go
func createUsersBatch(db *sql.DB, users []User) error {
    // Подготавливаем выражение
    stmt, err := db.Prepare("INSERT INTO users (name, age) VALUES ($1, $2)")
    if err != nil {
        return fmt.Errorf("ошибка подготовки выражения: %v", err)
    }
    defer stmt.Close()
    
    // Выполняем для каждого пользователя
    for _, user := range users {
        _, err := stmt.Exec(user.Name, user.Age)
        if err != nil {
            return fmt.Errorf("ошибка вставки пользователя %s: %v", user.Name, err)
        }
    }
    
    return nil
}
```

## Обработка ошибок

### Типичные ошибки баз данных

```go
func handleDatabaseErrors() {
    // sql.ErrNoRows - когда SELECT не нашел строк
    if err == sql.ErrNoRows {
        fmt.Println("Запись не найдена")
    }
    
    // Ошибки соединения
    if strings.Contains(err.Error(), "connection refused") {
        fmt.Println("Нет соединения с базой данных")
    }
    
    // Ошибки уникальности
    if strings.Contains(err.Error(), "duplicate key") {
        fmt.Println("Нарушение уникальности")
    }
}
```

## Практические советы

### 1. Используйте миграции

```go
// Создание таблицы
func createUsersTable(db *sql.DB) error {
    query := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        age INTEGER NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`
    
    _, err := db.Exec(query)
    return err
}
```

### 2. Валидация данных

```go
func validateUser(user User) error {
    if user.Name == "" {
        return fmt.Errorf("имя не может быть пустым")
    }
    
    if user.Age < 0 || user.Age > 150 {
        return fmt.Errorf("возраст должен быть от 0 до 150")
    }
    
    return nil
}
```

### 3. Логирование запросов

```go
func logQuery(query string, args ...interface{}) {
    log.Printf("Выполняем запрос: %s с параметрами: %v", query, args)
}
```

### 4. Graceful shutdown

```go
func main() {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }
    
    // Graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    
    go func() {
        <-c
        fmt.Println("Закрываем соединение с базой данных...")
        db.Close()
        os.Exit(0)
    }()
    
    // Используем базу данных...
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

## См. также

- [Context](context.md) - как использовать контексты с базами данных
- [Тестирование с базами данных](../theory/database-testing.md) - как тестировать код работы с БД
- [Пулы соединений](../theory/connection-pooling.md) - подробнее о пулах
- [Транзакции](../theory/transactions.md) - подробнее о транзакциях