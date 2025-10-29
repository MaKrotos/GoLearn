# Контекст (Context) в Go: Полная теория

## Введение в контекст

### Что такое контекст?

Контекст в Go - это **механизм передачи сигналов отмены, таймаутов и метаданных** между функциями и горутинами. Он реализует интерфейс `context.Context`.

### Зачем нужен контекст?

Контекст решает три основные задачи:
1. **Отмена операций** - возможность прервать выполнение
2. **Таймауты** - ограничение времени выполнения
3. **Передача метаданных** - передача значений между функциями

## Интерфейс Context

### Определение интерфейса

```go
type Context interface {
    Deadline() (deadline time.Time, ok bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
}
```

### Методы интерфейса

#### 1. Deadline()

Возвращает время, когда контекст будет отменен:

```go
func exampleDeadline(ctx context.Context) {
    if deadline, ok := ctx.Deadline(); ok {
        fmt.Printf("Контекст будет отменен в: %v\n", deadline)
        timeLeft := time.Until(deadline)
        fmt.Printf("Осталось времени: %v\n", timeLeft)
    } else {
        fmt.Println("Контекст без дедлайна")
    }
}
```

#### 2. Done()

Возвращает канал, который закрывается при отмене контекста:

```go
func longRunningOperation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // Возвращаем причину отмены
        default:
            // Продолжаем работу
            time.Sleep(100 * time.Millisecond)
            fmt.Println("Работаю...")
        }
    }
}
```

#### 3. Err()

Возвращает ошибку, которая привела к отмене контекста:

```go
func handleContextError(ctx context.Context) {
    select {
    case <-ctx.Done():
        switch ctx.Err() {
        case context.Canceled:
            fmt.Println("Контекст отменен")
        case context.DeadlineExceeded:
            fmt.Println("Превышен дедлайн")
        }
    case <-time.After(1 * time.Second):
        fmt.Println("Операция завершена успешно")
    }
}
```

#### 4. Value()

Позволяет передавать метаданные через контекст:

```go
type contextKey string

const (
    userIDKey contextKey = "userID"
    roleKey   contextKey = "role"
)

func withMetadata(ctx context.Context) context.Context {
    ctx = context.WithValue(ctx, userIDKey, "12345")
    ctx = context.WithValue(ctx, roleKey, "admin")
    return ctx
}

func useMetadata(ctx context.Context) {
    if userID, ok := ctx.Value(userIDKey).(string); ok {
        fmt.Printf("User ID: %s\n", userID)
    }
    
    if role, ok := ctx.Value(roleKey).(string); ok {
        fmt.Printf("Role: %s\n", role)
    }
}
```

## Типы контекстов

### 1. context.Background()

Базовый контекст, который никогда не отменяется:

```go
func main() {
    ctx := context.Background()
    
    // Используется как родитель для других контекстов
    childCtx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    // Работа с childCtx
}
```

### 2. context.TODO()

Заглушка для мест, где контекст будет добавлен позже:

```go
func someFunction() {
    // Пока не знаем, какой контекст использовать
    ctx := context.TODO()
    
    // Позже заменим на реальный контекст
    doSomething(ctx)
}
```

### 3. context.WithCancel()

Создает контекст, который можно отменить вручную:

```go
func withCancelExample() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel() // Всегда вызывайте cancel
    
    // Запускаем горутину
    go func() {
        select {
        case <-time.After(2 * time.Second):
            fmt.Println("Работа завершена")
        case <-ctx.Done():
            fmt.Println("Работа отменена")
        }
    }()
    
    // Отменяем через 1 секунду
    time.Sleep(1 * time.Second)
    cancel()
    
    time.Sleep(1 * time.Second)
}
```

### 4. context.WithTimeout()

Создает контекст с автоматической отменой по таймауту:

```go
func withTimeoutExample() {
    // Контекст отменится через 2 секунды
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    
    // Долгая операция
    err := longRunningOperation(ctx)
    if err != nil {
        switch ctx.Err() {
        case context.DeadlineExceeded:
            fmt.Println("Операция превысила таймаут")
        case context.Canceled:
            fmt.Println("Операция была отменена")
        }
        return
    }
    
    fmt.Println("Операция завершена успешно")
}
```

### 5. context.WithDeadline()

Создает контекст, который отменяется в определенное время:

```go
func withDeadlineExample() {
    // Контекст отменится в 15:04:05
    deadline := time.Now().Add(5 * time.Second)
    ctx, cancel := context.WithDeadline(context.Background(), deadline)
    defer cancel()
    
    // Работа с контекстом
    select {
    case <-time.After(10 * time.Second):
        fmt.Println("Работа завершена")
    case <-ctx.Done():
        fmt.Printf("Работа отменена: %v\n", ctx.Err())
    }
}
```

### 6. context.WithValue()

Передает метаданные через контекст:

```go
type User struct {
    ID   string
    Name string
}

type contextKey string

const userKey contextKey = "user"

func authenticateMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Извлекаем токен и проверяем пользователя
        token := r.Header.Get("Authorization")
        user := validateToken(token)
        
        // Добавляем пользователя в контекст
        ctx := context.WithValue(r.Context(), userKey, user)
        
        // Передаем контекст дальше
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    // Извлекаем пользователя из контекста
    user, ok := r.Context().Value(userKey).(*User)
    if !ok {
        http.Error(w, "Не авторизован", http.StatusUnauthorized)
        return
    }
    
    fmt.Fprintf(w, "Профиль пользователя: %s", user.Name)
}
```

## Практическое применение

### 1. HTTP серверы

```go
func apiHandler(w http.ResponseWriter, r *http.Request) {
    // Создаем контекст с таймаутом
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    // Выполняем операцию с контекстом
    result, err := fetchData(ctx)
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            http.Error(w, "Таймаут запроса", http.StatusRequestTimeout)
            return
        }
        http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
        return
    }
    
    // Отправляем результат
    json.NewEncoder(w).Encode(result)
}
```

### 2. Работа с базами данных

```go
func getUserByID(ctx context.Context, db *sql.DB, id int) (*User, error) {
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
```

### 3. Микросервисы

```go
func callExternalService(ctx context.Context, client *http.Client, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    return io.ReadAll(resp.Body)
}
```

## Лучшие практики

### 1. Первый параметр в функциях

```go
// ПРАВИЛЬНО
func DoSomething(ctx context.Context, data string) error {
    // ...
}

// НЕПРАВИЛЬНО
func DoSomething(data string, ctx context.Context) error {
    // ...
}
```

### 2. Всегда вызывайте cancel()

```go
// ПРАВИЛЬНО
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // Обязательно!

// НЕПРАВИЛЬНО
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// Забыли cancel() - утечка ресурсов!
```

### 3. Проверяйте ctx.Done()

```go
func longRunningOperation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Продолжаем работу
            if err := doWork(); err != nil {
                return err
            }
        }
    }
}
```

### 4. Используйте типизированные ключи

```go
// ПРАВИЛЬНО
type contextKey string

const userIDKey contextKey = "userID"

ctx := context.WithValue(ctx, userIDKey, "12345")

// НЕПРАВИЛЬНО
ctx := context.WithValue(ctx, "userID", "12345") // Нетипизированный ключ
```

### 5. Не передавайте конфиденциальные данные

```go
// НЕ РЕКОМЕНДУЕТСЯ
ctx = context.WithValue(ctx, "password", "secret123")

// ЛУЧШЕ
// Используйте безопасные механизмы аутентификации
```

## Распространенные ошибки

### 1. Игнорирование контекста

```go
// ОШИБКА
func badFunction(ctx context.Context) error {
    // Игнорируем ctx и выполняем долгую операцию
    time.Sleep(10 * time.Second)
    return nil
}

// ИСПРАВЛЕНО
func goodFunction(ctx context.Context) error {
    select {
    case <-time.After(10 * time.Second):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

### 2. Передача nil контекста

```go
// ОШИБКА
err := DoSomething(nil, data)

// ИСПРАВЛЕНО
err := DoSomething(context.Background(), data)
```

### 3. Забытый defer cancel()

```go
// ОШИБКА
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// Забыли defer cancel()

// ИСПРАВЛЕНО
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

## Профилирование контекста

### Мониторинг таймаутов

```go
func monitoredHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    err := processRequest(ctx)
    
    duration := time.Since(start)
    
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            log.Printf("Таймаут запроса: %v", duration)
        }
        http.Error(w, "Ошибка", http.StatusInternalServerError)
        return
    }
    
    log.Printf("Запрос выполнен за: %v", duration)
    w.WriteHeader(http.StatusOK)
}
```

### Трассировка запросов

```go
type traceIDKey struct{}

func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := generateTraceID()
        
        ctx := context.WithValue(r.Context(), traceIDKey{}, traceID)
        
        log.Printf("Начало запроса %s: %s %s", traceID, r.Method, r.URL.Path)
        
        start := time.Now()
        next.ServeHTTP(w, r.WithContext(ctx))
        duration := time.Since(start)
        
        log.Printf("Завершение запроса %s за %v", traceID, duration)
    })
}
```

## Расширенные примеры

### 1. Fan-out с контекстом

```go
func fanOutWithContext(ctx context.Context, data []int) ([]int, error) {
    results := make(chan int, len(data))
    errChan := make(chan error, len(data))
    
    // Запускаем воркеров
    for _, item := range data {
        go func(val int) {
            result, err := processItem(ctx, val)
            if err != nil {
                errChan <- err
                return
            }
            results <- result
        }(item)
    }
    
    // Собираем результаты
    var processed []int
    for i := 0; i < len(data); i++ {
        select {
        case result := <-results:
            processed = append(processed, result)
        case err := <-errChan:
            return nil, err
        case <-ctx.Done():
            return nil, ctx.Err()
        }
    }
    
    return processed, nil
}
```

### 2. Retry с экспоненциальной задержкой

```go
func retryWithContext(ctx context.Context, operation func() error, maxRetries int) error {
    var backoff = time.Second
    
    for i := 0; i < maxRetries; i++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        // Проверяем контекст
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
        
        // Ждем с экспоненциальной задержкой
        select {
        case <-time.After(backoff):
            backoff *= 2 // Удваиваем задержку
            if backoff > time.Minute {
                backoff = time.Minute // Максимальная задержка
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return errors.New("превышено максимальное количество попыток")
}
```

## См. также

- [Контекст для чайников](../concepts/context.md) - базовое объяснение
- [HTTP серверы](../concepts/http-server.md) - использование контекста в веб-приложениях
- [Базы данных](../concepts/database.md) - работа с контекстом в database/sql
- [Горутины](../concepts/goroutine.md) - что отменять с помощью контекста
- [Практические примеры](../examples/context) - примеры кода