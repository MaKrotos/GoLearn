# Context - объяснение для чайников

## Что такое context?

Представьте, что вы планируете поездку. У вас есть **контекст** всей поездки - это время, когда вы должны вернуться домой, ваш бюджет, погодные условия и т.д.

Context в Go - это **способ передачи важной информации** от одной части программы к другой, особенно когда эта информация может **отменить выполнение** или **ограничить время** выполнения.

## Техническое определение

Context - это **интерфейс**, который **несет сигналы отмены, таймауты и метаданные** между функциями и горутинами.

## Основные функции context

### 1. Отмена операций
Когда что-то идет не так, context позволяет **отменить все связанные операции**.

### 2. Таймауты
Context может **ограничивать время выполнения** операций.

### 3. Передача метаданных
Context может **нести дополнительную информацию** (например, ID пользователя, токены авторизации).

## Типы context

### 1. context.Background()

```go
ctx := context.Background()
```

Это **базовый контекст**, который никогда не отменяется и не имеет таймаута. Используется как **родитель для других контекстов**.

### 2. context.WithCancel()

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // Всегда вызывайте cancel

// В другой горутине:
// cancel() // Отменяет контекст
```

Позволяет **вручную отменить** операции.

### 3. context.WithTimeout()

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // Всегда вызывайте cancel
```

Автоматически отменяет операции **по истечении времени**.

### 4. context.WithDeadline()

```go
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()
```

Отменяет операции **в определенное время**.

### 5. context.WithValue()

```go
type key string
const userIDKey key = "userID"

ctx := context.WithValue(context.Background(), userIDKey, "12345")
userID := ctx.Value(userIDKey).(string)
```

Позволяет **передавать метаданные** через контекст.

## Практические примеры

### Пример 1: Отмена HTTP запроса

```go
func fetchData() error {
    // Создаем контекст с таймаутом 5 секунд
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Создаем HTTP запрос с контекстом
    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
    if err != nil {
        return err
    }
    
    // Выполняем запрос
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    // Если запрос занимает больше 5 секунд, он автоматически отменяется
    return nil
}
```

### Пример 2: Отмена долгой операции

```go
func processData(ctx context.Context, data []int) error {
    for i, item := range data {
        // Проверяем, не отменен ли контекст
        select {
        case <-ctx.Done():
            return ctx.Err() // Возвращаем причину отмены
        default:
        }
        
        // Имитация долгой обработки
        time.Sleep(100 * time.Millisecond)
        fmt.Printf("Обработан элемент %d\n", i)
    }
    return nil
}

func main() {
    data := make([]int, 100)
    
    // Создаем контекст с отменой через 3 секунды
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    err := processData(ctx, data)
    if err != nil {
        fmt.Printf("Операция отменена: %v\n", err)
    }
}
```

### Пример 3: Передача метаданных

```go
type contextKey string
const (
    userIDKey contextKey = "userID"
    roleKey   contextKey = "role"
)

func authenticateMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Извлекаем токен из заголовка
        token := r.Header.Get("Authorization")
        
        // Проверяем токен и получаем данные пользователя
        userID, role := validateToken(token)
        
        // Добавляем данные в контекст
        ctx := context.WithValue(r.Context(), userIDKey, userID)
        ctx = context.WithValue(ctx, roleKey, role)
        
        // Передаем контекст дальше
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func getUserProfile(w http.ResponseWriter, r *http.Request) {
    // Извлекаем данные из контекста
    userID := r.Context().Value(userIDKey).(string)
    role := r.Context().Value(roleKey).(string)
    
    fmt.Printf("Пользователь: %s, Роль: %s\n", userID, role)
}
```

## Правила использования context

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

### 2. Всегда проверяйте ctx.Done()

```go
func longRunningOperation(ctx context.Context) error {
    for {
        select {
        case <-ctx.Done():
            return ctx.Err() // Возвращаем причину отмены
        default:
            // Продолжаем работу
        }
        
        // Делаем что-то полезное
        time.Sleep(100 * time.Millisecond)
    }
}
```

### 3. Не передавайте nil context

```go
// НЕПРАВИЛЬНО
err := DoSomething(nil, data)

// ПРАВИЛЬНО
err := DoSomething(context.Background(), data)
```

## Распространенные ошибки

### 1. Забытый cancel()

```go
// ОШИБКА
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// Забыли defer cancel() - ресурсы не освобождаются!

// ПРАВИЛЬНО
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel() // Обязательно!
```

### 2. Использование context.Value() неправильно

```go
// ПЛОХО
userID := ctx.Value("userID").(string) // Нетипизированный ключ

// ЛУЧШЕ
type contextKey string
const userIDKey contextKey = "userID"
userID := ctx.Value(userIDKey).(string) // Типизированный ключ
```

### 3. Передача конфиденциальных данных

```go
// НЕ РЕКОМЕНДУЕТСЯ
ctx = context.WithValue(ctx, "password", "secret123")

// ЛУЧШЕ
// Не передавайте конфиденциальные данные через context
```

## Лучшие практики

1. **Используйте context.Background()** как родитель для других контекстов
2. **Всегда вызывайте cancel()** через defer
3. **Проверяйте ctx.Done()** в долгих операциях
4. **Используйте типизированные ключи** для context.Value()
5. **Не передавайте конфиденциальные данные** через context
6. **Не храните context** в структурах - передавайте как параметр

## См. также

- [Горутины](goroutine.md) - для чего нужны контексты
- [Каналы](channel.md) - альтернативный способ передачи сигналов
- [HTTP серверы](../theory/http-server.md) - где часто используются контексты
- [Базы данных](../theory/database.md) - работа с контекстами в database/sql