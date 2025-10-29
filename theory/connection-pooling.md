# Пул соединений в Go: Полная теория

## Введение в пулы соединений

### Что такое пул соединений?

Пул соединений - это **механизм повторного использования** соединений с базами данных, чтобы:
- **Снизить накладные расходы** на создание/закрытие соединений
- **Ограничить количество одновременных соединений**
- **Улучшить производительность** приложений
- **Управлять ресурсами** базы данных

### Зачем нужны пулы соединений?

1. **Производительность** - создание соединения занимает время
2. **Ресурсы** - базы данных имеют ограничения на количество соединений
3. **Стабильность** - предотвращение исчерпания соединений
4. **Масштабируемость** - эффективное использование ресурсов

## Основы пулов соединений в Go

### database/sql и пулы соединений

Go стандартная библиотека `database/sql` **автоматически управляет** пулом соединений:

```go
// Создание пула соединений
db, err := sql.Open("postgres", "user=postgres dbname=test sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Пул автоматически создается при первом использовании
```

### Конфигурация пула соединений

```go
// Настройка параметров пула
db.SetMaxOpenConns(25)     // Максимальное количество открытых соединений
db.SetMaxIdleConns(25)     // Максимальное количество простаивающих соединений
db.SetConnMaxLifetime(5 * time.Minute)  // Максимальное время жизни соединения
db.SetConnMaxIdleTime(5 * time.Minute)  // Максимальное время простоя соединения
```

### Понимание параметров

1. **MaxOpenConns** - ограничивает общее количество соединений
2. **MaxIdleConns** - количество соединений, которые остаются открытыми для повторного использования
3. **ConnMaxLifetime** - время, после которого соединение будет закрыто
4. **ConnMaxIdleTime** - время, после которого простаивающее соединение будет закрыто

## Практическая реализация

### 1. Базовая настройка

```go
// database/connection.go
package database

import (
    "database/sql"
    "time"
    _ "github.com/lib/pq" // PostgreSQL driver
)

type DB struct {
    *sql.DB
}

func NewConnection(connectionString string) (*DB, error) {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    // Настройка пула соединений
    db.SetMaxOpenConns(25)                 // Максимум 25 соединений
    db.SetMaxIdleConns(25)                 // 25 простаивающих соединений
    db.SetConnMaxLifetime(5 * time.Minute) // Закрывать соединения через 5 минут
    db.SetConnMaxIdleTime(5 * time.Minute) // Закрывать простаивающие через 5 минут
    
    // Проверка соединения
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, err
    }
    
    return &DB{db}, nil
}

func (db *DB) Close() error {
    return db.DB.Close()
}
```

### 2. Использование пула соединений

```go
// repository/user.go
package repository

import (
    "context"
    "database/sql"
    "yourproject/database"
)

type User struct {
    ID    int
    Name  string
    Email string
}

type UserRepository struct {
    db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    var user User
    err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil // Пользователь не найден
        }
        return nil, err
    }
    
    return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
    query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
    
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *User) error {
    query := "UPDATE users SET name = $1, email = $2 WHERE id = $3"
    
    result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }
    
    return nil
}
```

### 3. Мониторинг пула соединений

```go
// database/stats.go
package database

import (
    "database/sql"
    "fmt"
    "time"
)

type PoolStats struct {
    OpenConnections int           // Количество открытых соединений
    InUse           int           // Соединения в использовании
    Idle            int           // Простаивающие соединения
    WaitCount       int64         // Количество ожиданий
    WaitDuration    time.Duration // Общее время ожидания
    MaxIdleClosed   int64         // Закрытые из-за превышения MaxIdleConns
    MaxLifetimeClosed int64       // Закрытые из-за превышения ConnMaxLifetime
}

func (db *DB) GetStats() PoolStats {
    stats := db.DB.Stats()
    
    return PoolStats{
        OpenConnections: stats.OpenConnections,
        InUse:           stats.InUse,
        Idle:            stats.Idle,
        WaitCount:       stats.WaitCount,
        WaitDuration:    stats.WaitDuration,
        MaxIdleClosed:   stats.MaxIdleClosed,
        MaxLifetimeClosed: stats.MaxLifetimeClosed,
    }
}

func (db *DB) PrintStats() {
    stats := db.GetStats()
    
    fmt.Printf("=== Pool Statistics ===\n")
    fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
    fmt.Printf("In Use: %d\n", stats.InUse)
    fmt.Printf("Idle: %d\n", stats.Idle)
    fmt.Printf("Wait Count: %d\n", stats.WaitCount)
    fmt.Printf("Wait Duration: %v\n", stats.WaitDuration)
    fmt.Printf("Max Idle Closed: %d\n", stats.MaxIdleClosed)
    fmt.Printf("Max Lifetime Closed: %d\n", stats.MaxLifetimeClosed)
    fmt.Printf("========================\n")
}
```

## Расширенные техники пулов соединений

### 1. Динамическая настройка пула

```go
// database/dynamic_pool.go
package database

import (
    "context"
    "database/sql"
    "sync"
    "time"
)

type DynamicPool struct {
    *sql.DB
    mu           sync.RWMutex
    config       PoolConfig
    statsMonitor *time.Ticker
}

type PoolConfig struct {
    MaxOpenConns    int
    MaxIdleConns    int
    MaxLifetime     time.Duration
    MaxIdleTime     time.Duration
    MonitorInterval time.Duration
}

func NewDynamicPool(connectionString string, config PoolConfig) (*DynamicPool, error) {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    pool := &DynamicPool{
        DB:     db,
        config: config,
    }
    
    // Применяем начальную конфигурацию
    pool.applyConfig()
    
    // Запускаем мониторинг
    if config.MonitorInterval > 0 {
        pool.startMonitoring()
    }
    
    return pool, nil
}

func (p *DynamicPool) applyConfig() {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    p.SetMaxOpenConns(p.config.MaxOpenConns)
    p.SetMaxIdleConns(p.config.MaxIdleConns)
    p.SetConnMaxLifetime(p.config.MaxLifetime)
    p.SetConnMaxIdleTime(p.config.MaxIdleTime)
}

func (p *DynamicPool) UpdateConfig(config PoolConfig) {
    p.mu.Lock()
    p.config = config
    p.mu.Unlock()
    
    p.applyConfig()
}

func (p *DynamicPool) startMonitoring() {
    p.statsMonitor = time.NewTicker(p.config.MonitorInterval)
    
    go func() {
        for range p.statsMonitor.C {
            stats := p.DB.Stats()
            
            // Адаптивная настройка на основе статистики
            if stats.WaitCount > 100 {
                // Увеличиваем пул при высокой нагрузке
                p.mu.Lock()
                newConfig := p.config
                newConfig.MaxOpenConns = min(newConfig.MaxOpenConns*2, 100)
                newConfig.MaxIdleConns = min(newConfig.MaxIdleConns*2, 50)
                p.config = newConfig
                p.mu.Unlock()
                
                p.applyConfig()
            }
        }
    }()
}

func (p *DynamicPool) Close() error {
    if p.statsMonitor != nil {
        p.statsMonitor.Stop()
    }
    return p.DB.Close()
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### 2. Пул соединений с retry логикой

```go
// database/retry_pool.go
package database

import (
    "context"
    "database/sql"
    "time"
)

type RetryPool struct {
    *sql.DB
    maxRetries int
    retryDelay time.Duration
}

func NewRetryPool(connectionString string, maxRetries int, retryDelay time.Duration) (*RetryPool, error) {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    return &RetryPool{
        DB:         db,
        maxRetries: maxRetries,
        retryDelay: retryDelay,
    }, nil
}

func (p *RetryPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    for i := 0; i <= p.maxRetries; i++ {
        row := p.DB.QueryRowContext(ctx, query, args...)
        
        // Проверяем ошибку через Scan
        var dummy interface{}
        err := row.Scan(&dummy)
        if err == nil || err == sql.ErrNoRows {
            // Возвращаем оригинальный row для нормальной работы
            return p.DB.QueryRowContext(ctx, query, args...)
        }
        
        if i < p.maxRetries {
            time.Sleep(p.retryDelay)
        }
    }
    
    return p.DB.QueryRowContext(ctx, query, args...)
}

func (p *RetryPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    var lastErr error
    
    for i := 0; i <= p.maxRetries; i++ {
        result, err := p.DB.ExecContext(ctx, query, args...)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        if i < p.maxRetries {
            time.Sleep(p.retryDelay)
        }
    }
    
    return nil, lastErr
}
```

### 3. Пул соединений с метриками

```go
// database/metrics_pool.go
package database

import (
    "context"
    "database/sql"
    "time"
)

type MetricsPool struct {
    *sql.DB
    metrics *PoolMetrics
}

type PoolMetrics struct {
    TotalQueries     int64
    TotalErrors      int64
    TotalQueryTime   time.Duration
    SlowQueries      int64 // Запросы дольше 1 секунды
    mu               sync.RWMutex
}

func NewMetricsPool(connectionString string) (*MetricsPool, error) {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    return &MetricsPool{
        DB:      db,
        metrics: &PoolMetrics{},
    }, nil
}

func (p *MetricsPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    start := time.Now()
    
    rows, err := p.DB.QueryContext(ctx, query, args...)
    
    p.recordQuery(start, err)
    return rows, err
}

func (p *MetricsPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    start := time.Now()
    
    row := p.DB.QueryRowContext(ctx, query, args...)
    
    // Для QueryRow мы не можем напрямую измерить время,
    // поэтому создаем обертку
    return &timedRow{
        Row:   row,
        start: start,
        pool:  p,
    }
}

func (p *MetricsPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    start := time.Now()
    
    result, err := p.DB.ExecContext(ctx, query, args...)
    
    p.recordQuery(start, err)
    return result, err
}

func (p *MetricsPool) recordQuery(start time.Time, err error) {
    duration := time.Since(start)
    
    p.metrics.mu.Lock()
    defer p.metrics.mu.Unlock()
    
    p.metrics.TotalQueries++
    p.metrics.TotalQueryTime += duration
    
    if err != nil {
        p.metrics.TotalErrors++
    }
    
    if duration > time.Second {
        p.metrics.SlowQueries++
    }
}

func (p *MetricsPool) GetMetrics() *PoolMetrics {
    p.metrics.mu.RLock()
    defer p.metrics.mu.RUnlock()
    
    // Возвращаем копию метрик
    return &PoolMetrics{
        TotalQueries:   p.metrics.TotalQueries,
        TotalErrors:    p.metrics.TotalErrors,
        TotalQueryTime: p.metrics.TotalQueryTime,
        SlowQueries:    p.metrics.SlowQueries,
    }
}

func (p *MetricsPool) PrintMetrics() {
    metrics := p.GetMetrics()
    
    avgTime := time.Duration(0)
    if metrics.TotalQueries > 0 {
        avgTime = metrics.TotalQueryTime / time.Duration(metrics.TotalQueries)
    }
    
    fmt.Printf("=== Database Metrics ===\n")
    fmt.Printf("Total Queries: %d\n", metrics.TotalQueries)
    fmt.Printf("Total Errors: %d\n", metrics.TotalErrors)
    fmt.Printf("Average Query Time: %v\n", avgTime)
    fmt.Printf("Slow Queries: %d\n", metrics.SlowQueries)
    if metrics.TotalQueries > 0 {
        errorRate := float64(metrics.TotalErrors) / float64(metrics.TotalQueries) * 100
        fmt.Printf("Error Rate: %.2f%%\n", errorRate)
    }
    fmt.Printf("========================\n")
}

// Обертка для sql.Row с измерением времени
type timedRow struct {
    *sql.Row
    start time.Time
    pool  *MetricsPool
}

func (tr *timedRow) Scan(dest ...interface{}) error {
    err := tr.Row.Scan(dest...)
    tr.pool.recordQuery(tr.start, err)
    return err
}
```

## Тестирование пулов соединений

### 1. Модульные тесты

```go
// database/connection_test.go
package database

import (
    "context"
    "database/sql"
    "testing"
    "time"
)

func TestConnectionPool(t *testing.T) {
    // Используем in-memory SQLite для тестов
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    
    // Создаем тестовую таблицу
    _, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
    if err != nil {
        t.Fatal(err)
    }
    
    // Настройка пула
    pool := &DB{db}
    pool.SetMaxOpenConns(5)
    pool.SetMaxIdleConns(2)
    
    // Тест вставки
    ctx := context.Background()
    _, err = pool.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "Test User")
    if err != nil {
        t.Fatal(err)
    }
    
    // Тест выборки
    var name string
    err = pool.QueryRowContext(ctx, "SELECT name FROM users WHERE id = 1").Scan(&name)
    if err != nil {
        t.Fatal(err)
    }
    
    if name != "Test User" {
        t.Errorf("Expected 'Test User', got '%s'", name)
    }
    
    // Проверка статистики
    stats := pool.Stats()
    if stats.OpenConnections == 0 {
        t.Error("Expected open connections")
    }
}

func TestConnectionPoolConcurrency(t *testing.T) {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()
    
    _, err = db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
    if err != nil {
        t.Fatal(err)
    }
    
    pool := &DB{db}
    pool.SetMaxOpenConns(3)
    pool.SetMaxIdleConns(1)
    
    // Параллельные запросы
    const numGoroutines = 10
    done := make(chan bool, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func(id int) {
            ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            
            _, err := pool.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", 
                fmt.Sprintf("User %d", id))
            if err != nil {
                t.Errorf("Goroutine %d failed: %v", id, err)
            }
            
            done <- true
        }(i)
    }
    
    // Ждем завершения всех горутин
    for i := 0; i < numGoroutines; i++ {
        select {
        case <-done:
        case <-time.After(10 * time.Second):
            t.Fatal("Test timeout")
        }
    }
    
    // Проверяем, что все записи созданы
    var count int
    err = pool.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count)
    if err != nil {
        t.Fatal(err)
    }
    
    if count != numGoroutines {
        t.Errorf("Expected %d users, got %d", numGoroutines, count)
    }
}
```

### 2. Интеграционные тесты

```go
// integration/database_test.go
package integration

import (
    "context"
    "testing"
    "time"
    "yourproject/database"
)

func TestConnectionPoolIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Подключение к тестовой базе данных
    connectionString := "user=test dbname=test_pool sslmode=disable"
    pool, err := database.NewConnection(connectionString)
    if err != nil {
        t.Fatal(err)
    }
    defer pool.Close()
    
    // Настройка пула для теста
    pool.SetMaxOpenConns(5)
    pool.SetMaxIdleConns(2)
    pool.SetConnMaxLifetime(30 * time.Second)
    
    ctx := context.Background()
    
    // Создаем тестовую таблицу
    _, err = pool.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS test_users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        t.Fatal(err)
    }
    
    // Очищаем таблицу
    _, err = pool.ExecContext(ctx, "TRUNCATE test_users")
    if err != nil {
        t.Fatal(err)
    }
    
    // Тест высокой нагрузки
    const numOperations = 100
    done := make(chan error, numOperations)
    
    for i := 0; i < numOperations; i++ {
        go func(id int) {
            opCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
            defer cancel()
            
            // Выполняем операции вставки и выборки
            _, err := pool.ExecContext(opCtx, 
                "INSERT INTO test_users (name) VALUES ($1)", 
                fmt.Sprintf("User %d", id))
            
            if err != nil {
                done <- err
                return
            }
            
            var name string
            err = pool.QueryRowContext(opCtx, 
                "SELECT name FROM test_users WHERE name = $1", 
                fmt.Sprintf("User %d", id)).Scan(&name)
            
            done <- err
        }(i)
    }
    
    // Собираем результаты
    for i := 0; i < numOperations; i++ {
        select {
        case err := <-done:
            if err != nil {
                t.Errorf("Operation %d failed: %v", i, err)
            }
        case <-time.After(10 * time.Second):
            t.Fatalf("Operation %d timeout", i)
        }
    }
    
    // Проверяем статистику пула
    stats := pool.GetStats()
    if stats.OpenConnections > 5 {
        t.Errorf("Too many open connections: %d", stats.OpenConnections)
    }
    
    if stats.WaitCount > 0 {
        t.Logf("Had to wait for connections: %d times", stats.WaitCount)
    }
}
```

## Лучшие практики пулов соединений

### 1. Настройка параметров

```go
// Рекомендуемые настройки для разных сценариев

// Веб-приложение с умеренной нагрузкой
func setupWebAppPool(db *sql.DB) {
    db.SetMaxOpenConns(25)        // Ограничение на количество соединений
    db.SetMaxIdleConns(5)         // Небольшой пул простаивающих соединений
    db.SetConnMaxLifetime(10 * time.Minute)  // Закрывать старые соединения
    db.SetConnMaxIdleTime(5 * time.Minute)   // Закрывать долго простаивающие
}

// Высоконагруженное приложение
func setupHighLoadPool(db *sql.DB) {
    db.SetMaxOpenConns(100)       // Больше соединений для высокой нагрузки
    db.SetMaxIdleConns(25)        // Больше простаивающих для быстрого ответа
    db.SetConnMaxLifetime(5 * time.Minute)   // Быстрее закрываем соединения
    db.SetConnMaxIdleTime(2 * time.Minute)   // Быстрее закрываем простаивающие
}

// Микросервис с низкой нагрузкой
func setupMicroservicePool(db *sql.DB) {
    db.SetMaxOpenConns(10)        // Меньше соединений
    db.SetMaxIdleConns(2)         // Минимальный пул простаивающих
    db.SetConnMaxLifetime(30 * time.Minute)  // Дольше держим соединения
    db.SetConnMaxIdleTime(10 * time.Minute)  // Дольше держим простаивающие
}
```

### 2. Обработка ошибок соединений

```go
// database/error_handling.go
package database

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "time"
)

type ConnectionError struct {
    Op  string
    Err error
}

func (e *ConnectionError) Error() string {
    return fmt.Sprintf("connection error during %s: %v", e.Op, e.Err)
}

func (e *ConnectionError) Unwrap() error {
    return e.Err
}

func IsConnectionError(err error) bool {
    var connErr *ConnectionError
    return errors.As(err, &connErr)
}

func (db *DB) ExecWithRetry(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    const maxRetries = 3
    const retryDelay = 100 * time.Millisecond
    
    var lastErr error
    
    for i := 0; i <= maxRetries; i++ {
        result, err := db.ExecContext(ctx, query, args...)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // Проверяем, можно ли повторить попытку
        if !isRetryableError(err) {
            return nil, &ConnectionError{Op: "exec", Err: err}
        }
        
        if i < maxRetries {
            select {
            case <-ctx.Done():
                return nil, ctx.Err()
            case <-time.After(retryDelay):
                // Продолжаем повторные попытки
            }
        }
    }
    
    return nil, &ConnectionError{Op: "exec", Err: lastErr}
}

func isRetryableError(err error) bool {
    // Список ошибок, которые можно повторить
    retryableErrors := []string{
        "connection refused",
        "connection reset",
        "broken pipe",
        "i/o timeout",
        "deadline exceeded",
    }
    
    errStr := err.Error()
    for _, retryable := range retryableErrors {
        if strings.Contains(strings.ToLower(errStr), retryable) {
            return true
        }
    }
    
    return false
}
```

### 3. Graceful shutdown

```go
// database/shutdown.go
package database

import (
    "context"
    "time"
)

func (db *DB) GracefulClose() error {
    // Создаем контекст с таймаутом для закрытия
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Закрываем новые соединения
    db.SetMaxOpenConns(0)
    
    // Ждем завершения активных соединений
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            // Принудительно закрываем
            return db.Close()
        case <-ticker.C:
            stats := db.Stats()
            if stats.InUse == 0 {
                // Все соединения освобождены
                return db.Close()
            }
            // Продолжаем ждать
        }
    }
}
```

## Распространенные ошибки и их решение

### 1. Исчерпание соединений

```go
// ПЛОХО - утечка соединений
func BadConnectionHandling() {
    db, _ := sql.Open("postgres", "connection_string")
    
    for i := 0; i < 1000; i++ {
        // Запрос без закрытия rows
        rows, _ := db.Query("SELECT * FROM users")
        // Забыли вызвать rows.Close() - утечка соединений!
        
        // Обработка результатов...
    }
}

// ХОРОШО - правильное закрытие ресурсов
func GoodConnectionHandling() {
    db, _ := sql.Open("postgres", "connection_string")
    
    for i := 0; i < 1000; i++ {
        rows, err := db.Query("SELECT * FROM users")
        if err != nil {
            // Обработка ошибки
            continue
        }
        
        // Обязательно закрываем rows
        defer rows.Close()
        
        // Обработка результатов...
        for rows.Next() {
            // Обработка каждой строки
        }
        
        // Проверяем ошибки итерации
        if err := rows.Err(); err != nil {
            // Обработка ошибки
        }
    }
}
```

### 2. Неправильная настройка пула

```go
// ПЛОХО - слишком большие значения
func BadPoolConfig(db *sql.DB) {
    db.SetMaxOpenConns(1000)  // Слишком много соединений
    db.SetMaxIdleConns(500)   // Слишком много простаивающих
    db.SetConnMaxLifetime(0)  // Соединения никогда не закрываются
}

// ХОРОШО - сбалансированная конфигурация
func GoodPoolConfig(db *sql.DB) {
    db.SetMaxOpenConns(25)                 // Ограничение на основе нагрузки
    db.SetMaxIdleConns(5)                  // Минимальный пул простаивающих
    db.SetConnMaxLifetime(10 * time.Minute) // Разумное время жизни
    db.SetConnMaxIdleTime(5 * time.Minute)  // Разумное время простоя
}
```

### 3. Игнорирование контекстов

```go
// ПЛОХО - игнорирование контекста
func BadWithContext() {
    db, _ := sql.Open("postgres", "connection_string")
    
    // Запрос без таймаута - может висеть бесконечно
    rows, _ := db.Query("SELECT * FROM large_table")
    defer rows.Close()
    
    // Обработка результатов...
}

// ХОРОШО - использование контекста с таймаутом
func GoodWithContext() {
    db, _ := sql.Open("postgres", "connection_string")
    
    // Контекст с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    rows, err := db.QueryContext(ctx, "SELECT * FROM large_table")
    if err != nil {
        // Обработка ошибки
        return
    }
    defer rows.Close()
    
    // Обработка результатов...
}
```

## Мониторинг и отладка

### 1. Логирование пула соединений

```go
// database/logging.go
package database

import (
    "log"
    "time"
)

func (db *DB) StartLogging(interval time.Duration) {
    ticker := time.NewTicker(interval)
    
    go func() {
        for range ticker.C {
            stats := db.Stats()
            
            log.Printf("DB Pool Stats - Open: %d, InUse: %d, Idle: %d, Wait: %d",
                stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
            
            if stats.WaitCount > 0 {
                log.Printf("WARNING: Connection waiting detected! WaitDuration: %v",
                    stats.WaitDuration)
            }
        }
    }()
}
```

### 2. Prometheus метрики

```go
// database/prometheus.go
package database

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    dbOpenConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "db_open_connections",
        Help: "Number of open database connections",
    })
    
    dbInUseConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "db_in_use_connections",
        Help: "Number of database connections currently in use",
    })
    
    dbWaitCount = promauto.NewCounter(prometheus.CounterOpts{
        Name: "db_wait_count",
        Help: "Total number of connections waited for",
    })
)

func (db *DB) StartPrometheusMonitoring(interval time.Duration) {
    ticker := time.NewTicker(interval)
    
    go func() {
        for range ticker.C {
            stats := db.Stats()
            
            dbOpenConnections.Set(float64(stats.OpenConnections))
            dbInUseConnections.Set(float64(stats.InUse))
            dbWaitCount.Add(float64(stats.WaitCount))
        }
    }()
}
```

## См. также

- [Работа с базами данных](../concepts/database.md) - основы работы с БД
- [Тестирование](../concepts/testing.md) - как тестировать код
- [Контекст](../concepts/context.md) - управление жизненным циклом операций
- [Профилирование](../concepts/profiling.md) - как измерять производительность
- [Практические примеры](../examples/connection-pooling) - примеры кода