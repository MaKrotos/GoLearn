# Middleware в Go: Полная теория

## Введение в middleware

### Что такое middleware?

Middleware в Go - это **функции промежуточного слоя**, которые:
- **Перехватывают** HTTP запросы и ответы
- **Выполняют** общие операции (логирование, аутентификация, CORS)
- **Передают** управление следующему обработчику
- **Могут** модифицировать запросы и ответы

### Зачем нужен middleware?

1. **Повторное использование** кода
2. **Разделение ответственности** (separation of concerns)
3. **Централизованная** обработка общих задач
4. **Композиция** функциональности
5. **Упрощение** бизнес-логики

## Основы middleware в Go

### Базовая структура middleware

```go
// Тип middleware функции
type Middleware func(http.Handler) http.Handler

// Базовый пример middleware
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Предварительная обработка
        log.Printf("Запрос: %s %s", r.Method, r.URL.Path)
        
        // Передаем управление следующему обработчику
        next.ServeHTTP(w, r)
        
        // Постобработка (если нужна)
        log.Printf("Запрос завершен: %s %s", r.Method, r.URL.Path)
    })
}
```

### Цепочка middleware

```go
// Создание цепочки middleware
func ChainMiddleware(middlewares ...Middleware) Middleware {
    return func(final http.Handler) http.Handler {
        // Применяем middleware в обратном порядке
        for i := len(middlewares) - 1; i >= 0; i-- {
            final = middlewares[i](final)
        }
        return final
    }
}

// Использование цепочки
func main() {
    mux := http.NewServeMux()
    
    // Определяем обработчики
    mux.HandleFunc("/api/users", usersHandler)
    mux.HandleFunc("/api/orders", ordersHandler)
    
    // Создаем цепочку middleware
    chain := ChainMiddleware(
        LoggingMiddleware,
        AuthMiddleware,
        CORSMiddleware,
    )
    
    // Применяем цепочку к маршрутизатору
    wrappedMux := chain(mux)
    
    log.Fatal(http.ListenAndServe(":8080", wrappedMux))
}
```

## Практическая реализация middleware

### 1. Логирование middleware

```go
// middleware/logging.go
package middleware

import (
    "log"
    "net/http"
    "time"
)

// ResponseWriter с поддержкой захвата статуса
type loggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
    size       int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
    size, err := lrw.ResponseWriter.Write(b)
    lrw.size += size
    return size, err
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
    return &loggingResponseWriter{
        ResponseWriter: w,
        statusCode:     http.StatusOK,
    }
}

// LoggingMiddleware логирует все HTTP запросы
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Создаем обертку для ResponseWriter
        lrw := newLoggingResponseWriter(w)
        
        // Передаем управление следующему обработчику
        next.ServeHTTP(lrw, r)
        
        // Логируем информацию о запросе
        log.Printf(
            "%s %s %s %d %d %v",
            r.RemoteAddr,
            r.Method,
            r.URL.Path,
            lrw.statusCode,
            lrw.size,
            time.Since(start),
        )
    })
}
```

### 2. Аутентификационный middleware

```go
// middleware/auth.go
package middleware

import (
    "context"
    "errors"
    "net/http"
    "strings"
)

// Ключи контекста
type contextKey string

const UserContextKey contextKey = "user"

// User структура пользователя
type User struct {
    ID    int
    Name  string
    Email string
    Role  string
}

// AuthMiddleware проверяет JWT токен и устанавливает пользователя в контекст
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Получаем заголовок Authorization
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header required", http.StatusUnauthorized)
            return
        }
        
        // Проверяем формат Bearer token
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
            return
        }
        
        // Валидируем токен (здесь упрощенная реализация)
        user, err := validateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Добавляем пользователя в контекст
        ctx := context.WithValue(r.Context(), UserContextKey, user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// GetUserFromContext получает пользователя из контекста
func GetUserFromContext(ctx context.Context) (*User, error) {
    user, ok := ctx.Value(UserContextKey).(*User)
    if !ok {
        return nil, errors.New("user not found in context")
    }
    return user, nil
}

// Упрощенная валидация токена (в реальном приложении используйте JWT библиотеки)
func validateToken(tokenString string) (*User, error) {
    // Здесь должна быть реальная валидация токена
    // Для примера возвращаем фиктивного пользователя
    if tokenString == "valid-token" {
        return &User{
            ID:    1,
            Name:  "John Doe",
            Email: "john@example.com",
            Role:  "user",
        }, nil
    }
    
    return nil, errors.New("invalid token")
}

// AdminMiddleware проверяет, что пользователь имеет роль администратора
func AdminMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, err := GetUserFromContext(r.Context())
        if err != nil {
            http.Error(w, "User not authenticated", http.StatusUnauthorized)
            return
        }
        
        if user.Role != "admin" {
            http.Error(w, "Admin access required", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### 3. CORS middleware

```go
// middleware/cors.go
package middleware

import (
    "net/http"
    "strings"
)

// CORSMiddleware добавляет CORS заголовки
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Добавляем CORS заголовки
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        // Обрабатываем preflight запросы
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// AdvancedCORSMiddleware с настройками
type CORSConfig struct {
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
    AllowCredentials bool
}

func NewCORSMiddleware(config CORSConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Определяем разрешенный origin
            origin := r.Header.Get("Origin")
            if origin != "" {
                if isOriginAllowed(origin, config.AllowedOrigins) {
                    w.Header().Set("Access-Control-Allow-Origin", origin)
                }
            }
            
            // Устанавливаем другие заголовки
            if len(config.AllowedMethods) > 0 {
                w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
            }
            
            if len(config.AllowedHeaders) > 0 {
                w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
            }
            
            if config.AllowCredentials {
                w.Header().Set("Access-Control-Allow-Credentials", "true")
            }
            
            // Обрабатываем preflight запросы
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
    if len(allowedOrigins) == 0 {
        return true // Разрешаем все origins
    }
    
    for _, allowed := range allowedOrigins {
        if allowed == "*" || allowed == origin {
            return true
        }
    }
    
    return false
}
```

### 4. Rate limiting middleware

```go
// middleware/ratelimit.go
package middleware

import (
    "net/http"
    "sync"
    "time"
)

// TokenBucket реализация алгоритма токенов
type TokenBucket struct {
    tokens       int
    maxTokens    int
    refillRate   time.Duration
    lastRefill   time.Time
    mu           sync.Mutex
}

func NewTokenBucket(maxTokens int, refillRate time.Duration) *TokenBucket {
    return &TokenBucket{
        tokens:     maxTokens,
        maxTokens:  maxTokens,
        refillRate: refillRate,
        lastRefill: time.Now(),
    }
}

func (tb *TokenBucket) Take() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()
    
    // Пополняем токены
    now := time.Now()
    tokensToAdd := int(now.Sub(tb.lastRefill) / tb.refillRate)
    if tokensToAdd > 0 {
        tb.tokens = min(tb.tokens+tokensToAdd, tb.maxTokens)
        tb.lastRefill = now
    }
    
    // Проверяем наличие токенов
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    
    return false
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// RateLimitMiddleware ограничивает количество запросов
func RateLimitMiddleware(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
    bucket := NewTokenBucket(maxRequests, window/time.Duration(maxRequests))
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !bucket.Take() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// IP-based rate limiting
type IPRateLimiter struct {
    limiters map[string]*TokenBucket
    mu       sync.RWMutex
    maxRequests int
    window   time.Duration
}

func NewIPRateLimiter(maxRequests int, window time.Duration) *IPRateLimiter {
    return &IPRateLimiter{
        limiters:    make(map[string]*TokenBucket),
        maxRequests: maxRequests,
        window:      window,
    }
}

func (irl *IPRateLimiter) GetLimiter(ip string) *TokenBucket {
    irl.mu.RLock()
    limiter, exists := irl.limiters[ip]
    irl.mu.RUnlock()
    
    if !exists {
        irl.mu.Lock()
        limiter, exists = irl.limiters[ip]
        if !exists {
            limiter = NewTokenBucket(irl.maxRequests, irl.window/time.Duration(irl.maxRequests))
            irl.limiters[ip] = limiter
        }
        irl.mu.Unlock()
    }
    
    return limiter
}

func (irl *IPRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := getClientIP(r)
        limiter := irl.GetLimiter(ip)
        
        if !limiter.Take() {
            http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

func getClientIP(r *http.Request) string {
    // Проверяем заголовки прокси
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        return strings.Split(forwarded, ",")[0]
    }
    
    // Проверяем другие заголовки
    if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
        return realIP
    }
    
    // Используем RemoteAddr
    return r.RemoteAddr
}
```

## Расширенные техники middleware

### 1. Middleware с конфигурацией

```go
// middleware/configurable.go
package middleware

import (
    "log"
    "net/http"
    "time"
)

// Config структура конфигурации middleware
type Config struct {
    SkipLoggingPaths []string
    LogLevel         string
    Timeout          time.Duration
}

// ConfigurableLoggingMiddleware логирование с настройками
func ConfigurableLoggingMiddleware(config Config) func(http.Handler) http.Handler {
    skipPaths := make(map[string]bool)
    for _, path := range config.SkipLoggingPaths {
        skipPaths[path] = true
    }
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Пропускаем определенные пути
            if skipPaths[r.URL.Path] {
                next.ServeHTTP(w, r)
                return
            }
            
            start := time.Now()
            lrw := newLoggingResponseWriter(w)
            
            next.ServeHTTP(lrw, r)
            
            // Логируем в зависимости от уровня
            switch config.LogLevel {
            case "debug":
                log.Printf(
                    "DEBUG: %s %s %s %d %d %v",
                    r.RemoteAddr, r.Method, r.URL.Path, lrw.statusCode, lrw.size, time.Since(start),
                )
            case "info":
                log.Printf(
                    "INFO: %s %s %d %v",
                    r.URL.Path, r.Method, lrw.statusCode, time.Since(start),
                )
            default:
                log.Printf(
                    "%s %s %d",
                    r.URL.Path, r.Method, lrw.statusCode,
                )
            }
        })
    }
}
```

### 2. Middleware с метриками

```go
// middleware/metrics.go
package middleware

import (
    "net/http"
    "sync"
    "time"
)

// Metrics структура для сбора метрик
type Metrics struct {
    RequestCount   map[string]int
    ResponseTimes  map[string][]time.Duration
    mu             sync.RWMutex
}

func NewMetrics() *Metrics {
    return &Metrics{
        RequestCount:  make(map[string]int),
        ResponseTimes: make(map[string][]time.Duration),
    }
}

func (m *Metrics) RecordRequest(method string, duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.RequestCount[method]++
    m.ResponseTimes[method] = append(m.ResponseTimes[method], duration)
}

func (m *Metrics) GetMetrics() (map[string]int, map[string][]time.Duration) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    // Возвращаем копии данных
    countCopy := make(map[string]int)
    timesCopy := make(map[string][]time.Duration)
    
    for k, v := range m.RequestCount {
        countCopy[k] = v
    }
    
    for k, v := range m.ResponseTimes {
        timesCopy[k] = append([]time.Duration(nil), v...)
    }
    
    return countCopy, timesCopy
}

// MetricsMiddleware собирает метрики запросов
func MetricsMiddleware(metrics *Metrics) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            lrw := newLoggingResponseWriter(w)
            next.ServeHTTP(lrw, r)
            
            duration := time.Since(start)
            metrics.RecordRequest(r.Method, duration)
        })
    }
}
```

### 3. Middleware с контекстом и таймаутами

```go
// middleware/timeout.go
package middleware

import (
    "context"
    "net/http"
    "time"
)

// TimeoutMiddleware добавляет таймаут к запросам
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Создаем контекст с таймаутом
            ctx, cancel := context.WithTimeout(r.Context(), timeout)
            defer cancel()
            
            // Создаем новый запрос с контекстом
            r = r.WithContext(ctx)
            
            // Создаем ResponseWriter, который отслеживает отмену контекста
            tw := &timeoutWriter{ResponseWriter: w, ctx: ctx}
            
            next.ServeHTTP(tw, r)
        })
    }
}

type timeoutWriter struct {
    http.ResponseWriter
    ctx context.Context
    written bool
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
    if tw.written {
        return tw.ResponseWriter.Write(b)
    }
    
    select {
    case <-tw.ctx.Done():
        tw.ResponseWriter.WriteHeader(http.StatusGatewayTimeout)
        tw.written = true
        return 0, tw.ctx.Err()
    default:
        tw.written = true
        return tw.ResponseWriter.Write(b)
    }
}

func (tw *timeoutWriter) WriteHeader(code int) {
    if tw.written {
        return
    }
    
    select {
    case <-tw.ctx.Done():
        tw.ResponseWriter.WriteHeader(http.StatusGatewayTimeout)
        tw.written = true
    default:
        tw.ResponseWriter.WriteHeader(code)
        tw.written = true
    }
}
```

## Тестирование middleware

### 1. Модульные тесты middleware

```go
// middleware/logging_test.go
package middleware

import (
    "bytes"
    "log"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestLoggingMiddleware(t *testing.T) {
    // Создаем буфер для логов
    var buf bytes.Buffer
    log.SetOutput(&buf)
    
    // Создаем тестовый обработчик
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })
    
    // Оборачиваем обработчик в middleware
    wrappedHandler := LoggingMiddleware(handler)
    
    // Создаем тестовый запрос
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    
    // Выполняем запрос
    wrappedHandler.ServeHTTP(w, req)
    
    // Проверяем результат
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    if w.Body.String() != "Hello, World!" {
        t.Errorf("Expected body 'Hello, World!', got '%s'", w.Body.String())
    }
    
    // Проверяем, что лог записан
    if buf.Len() == 0 {
        t.Error("Expected log output, got none")
    }
    
    logOutput := buf.String()
    if !bytes.Contains([]byte(logOutput), []byte("/test")) {
        t.Errorf("Expected log to contain '/test', got '%s'", logOutput)
    }
}

func TestCORSMiddleware(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    wrappedHandler := CORSMiddleware(handler)
    
    // Тест OPTIONS запроса
    req := httptest.NewRequest("OPTIONS", "/test", nil)
    w := httptest.NewRecorder()
    
    wrappedHandler.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
    }
    
    // Проверяем CORS заголовки
    allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
    if allowOrigin != "*" {
        t.Errorf("Expected Access-Control-Allow-Origin '*', got '%s'", allowOrigin)
    }
    
    allowMethods := w.Header().Get("Access-Control-Allow-Methods")
    expectedMethods := "GET, POST, PUT, DELETE, OPTIONS"
    if allowMethods != expectedMethods {
        t.Errorf("Expected Access-Control-Allow-Methods '%s', got '%s'", expectedMethods, allowMethods)
    }
}

func TestAuthMiddleware(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    wrappedHandler := AuthMiddleware(handler)
    
    // Тест без заголовка Authorization
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    
    wrappedHandler.ServeHTTP(w, req)
    
    if w.Code != http.StatusUnauthorized {
        t.Errorf("Expected status 401, got %d", w.Code)
    }
    
    // Тест с невалидным токеном
    req = httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    w = httptest.NewRecorder()
    
    wrappedHandler.ServeHTTP(w, req)
    
    if w.Code != http.StatusUnauthorized {
        t.Errorf("Expected status 401, got %d", w.Code)
    }
    
    // Тест с валидным токеном
    req = httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    w = httptest.NewRecorder()
    
    wrappedHandler.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
```

### 2. Интеграционные тесты middleware

```go
// integration/middleware_test.go
package integration

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    "yourproject/middleware"
)

func TestMiddlewareChain(t *testing.T) {
    // Создаем тестовый обработчик
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, err := middleware.GetUserFromContext(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(user.Name))
    })
    
    // Создаем цепочку middleware
    chain := middleware.ChainMiddleware(
        middleware.LoggingMiddleware,
        middleware.AuthMiddleware,
        middleware.CORSMiddleware,
    )
    
    wrappedHandler := chain(handler)
    
    // Создаем тестовый запрос с валидным токеном
    req := httptest.NewRequest("GET", "/api/users", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    w := httptest.NewRecorder()
    
    wrappedHandler.ServeHTTP(w, req)
    
    // Проверяем результат
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    if w.Body.String() != "John Doe" {
        t.Errorf("Expected body 'John Doe', got '%s'", w.Body.String())
    }
    
    // Проверяем CORS заголовки
    allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
    if allowOrigin != "*" {
        t.Errorf("Expected CORS header, got '%s'", allowOrigin)
    }
}

func TestRateLimitingMiddleware(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    // Создаем rate limiter: 2 запроса в секунду
    rateLimiter := middleware.NewIPRateLimiter(2, time.Second)
    wrappedHandler := rateLimiter.Middleware(handler)
    
    // Создаем тестовый запрос
    req := httptest.NewRequest("GET", "/api/test", nil)
    req.RemoteAddr = "127.0.0.1:12345"
    
    // Выполняем 3 запроса быстро
    for i := 0; i < 3; i++ {
        w := httptest.NewRecorder()
        wrappedHandler.ServeHTTP(w, req)
        
        if i < 2 {
            // Первые два должны пройти
            if w.Code != http.StatusOK {
                t.Errorf("Request %d: Expected status 200, got %d", i+1, w.Code)
            }
        } else {
            // Третий должен быть заблокирован
            if w.Code != http.StatusTooManyRequests {
                t.Errorf("Request %d: Expected status 429, got %d", i+1, w.Code)
            }
        }
    }
}

func TestTimeoutMiddleware(t *testing.T) {
    // Создаем обработчик, который спит дольше таймаута
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(100 * time.Millisecond)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Done"))
    })
    
    // Применяем middleware с таймаутом 50ms
    wrappedHandler := middleware.TimeoutMiddleware(50 * time.Millisecond)(handler)
    
    req := httptest.NewRequest("GET", "/slow", nil)
    w := htthttptest.NewRecorder()
    
    wrappedHandler.ServeHTTP(w, req)
    
    // Должен вернуть таймаут
    if w.Code != http.StatusGatewayTimeout {
        t.Errorf("Expected status 504, got %d", w.Code)
    }
    
    // Создаем быстрый обработчик
    fastHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Fast"))
    })
    
    fastWrappedHandler := middleware.TimeoutMiddleware(50 * time.Millisecond)(fastHandler)
    
    req = httptest.NewRequest("GET", "/fast", nil)
    w = httptest.NewRecorder()
    
    fastWrappedHandler.ServeHTTP(w, req)
    
    // Должен вернуть нормальный ответ
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    
    if w.Body.String() != "Fast" {
        t.Errorf("Expected body 'Fast', got '%s'", w.Body.String())
    }
}
```

## Лучшие практики middleware

### 1. Порядок применения middleware

```go
// Правильный порядок middleware
func setupRouter() http.Handler {
    mux := http.NewServeMux()
    
    // Определяем маршруты
    mux.HandleFunc("/api/users", usersHandler)
    mux.HandleFunc("/api/admin", adminHandler)
    
    // Цепочка middleware в правильном порядке:
    chain := middleware.ChainMiddleware(
        // 1. Восстановление после паники (первый для защиты)
        RecoveryMiddleware,
        
        // 2. Логирование (рано для захвата всех запросов)
        LoggingMiddleware,
        
        // 3. CORS (до аутентификации для preflight запросов)
        CORSMiddleware,
        
        // 4. Таймауты (до тяжелых операций)
        TimeoutMiddleware(30 * time.Second),
        
        // 5. Rate limiting (до аутентификации для защиты)
        RateLimitMiddleware(100, time.Minute),
        
        // 6. Аутентификация (после rate limiting)
        AuthMiddleware,
        
        // 7. Авторизация (после аутентификации)
        AdminMiddleware,
    )
    
    return chain(mux)
}
```

### 2. Обработка ошибок в middleware

```go
// middleware/error.go
package middleware

import (
    "log"
    "net/http"
)

// RecoveryMiddleware восстанавливается после паники
func RecoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                
                // Отправляем 500 ошибку
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}

// ErrorMiddleware централизованная обработка ошибок
func ErrorMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Создаем обертку для ResponseWriter, которая отслеживает ошибки
        ew := &errorWriter{ResponseWriter: w}
        
        next.ServeHTTP(ew, r)
        
        // Если была ошибка, логируем её
        if ew.statusCode >= 400 {
            log.Printf("HTTP %d error for %s %s", ew.statusCode, r.Method, r.URL.Path)
        }
    })
}

type errorWriter struct {
    http.ResponseWriter
    statusCode int
}

func (ew *errorWriter) WriteHeader(code int) {
    ew.statusCode = code
    ew.ResponseWriter.WriteHeader(code)
}
```

### 3. Конфигурация middleware

```go
// middleware/config.go
package middleware

import (
    "encoding/json"
    "io/ioutil"
    "time"
)

// MiddlewareConfig конфигурация всех middleware
type MiddlewareConfig struct {
    Logging struct {
        Enabled         bool     `json:"enabled"`
        SkipPaths       []string `json:"skip_paths"`
        LogLevel        string   `json:"log_level"`
    } `json:"logging"`
    
    Auth struct {
        Enabled     bool   `json:"enabled"`
        SecretKey   string `json:"secret_key"`
        SkipPaths   []string `json:"skip_paths"`
    } `json:"auth"`
    
    CORS struct {
        Enabled        bool     `json:"enabled"`
        AllowedOrigins []string `json:"allowed_origins"`
        AllowedMethods []string `json:"allowed_methods"`
        AllowedHeaders []string `json:"allowed_headers"`
    } `json:"cors"`
    
    RateLimit struct {
        Enabled     bool          `json:"enabled"`
        Requests    int           `json:"requests"`
        Window      time.Duration `json:"window"`
    } `json:"rate_limit"`
    
    Timeout struct {
        Enabled bool          `json:"enabled"`
        Duration time.Duration `json:"duration"`
    } `json:"timeout"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(filename string) (*MiddlewareConfig, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config MiddlewareConfig
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}

// ApplyConfig применяет конфигурацию к маршрутизатору
func ApplyConfig(mux http.Handler, config *MiddlewareConfig) http.Handler {
    var middlewares []Middleware
    
    // Добавляем middleware в зависимости от конфигурации
    if config.Logging.Enabled {
        middlewares = append(middlewares, 
            ConfigurableLoggingMiddleware(Config{
                SkipLoggingPaths: config.Logging.SkipPaths,
                LogLevel:         config.Logging.LogLevel,
            }))
    }
    
    if config.CORS.Enabled {
        corsMiddleware := NewCORSMiddleware(CORSConfig{
            AllowedOrigins: config.CORS.AllowedOrigins,
            AllowedMethods: config.CORS.AllowedMethods,
            AllowedHeaders: config.CORS.AllowedHeaders,
        })
        middlewares = append(middlewares, corsMiddleware)
    }
    
    if config.RateLimit.Enabled {
        middlewares = append(middlewares, 
            RateLimitMiddleware(config.RateLimit.Requests, config.RateLimit.Window))
    }
    
    if config.Timeout.Enabled {
        middlewares = append(middlewares, 
            TimeoutMiddleware(config.Timeout.Duration))
    }
    
    if config.Auth.Enabled {
        middlewares = append(middlewares, AuthMiddleware)
    }
    
    // Применяем цепочку middleware
    if len(middlewares) > 0 {
        chain := ChainMiddleware(middlewares...)
        return chain(mux)
    }
    
    return mux
}
```

## Распространенные ошибки и их решение

### 1. Неправильный порядок middleware

```go
// ПЛОХО - неправильный порядок
func BadMiddlewareOrder() http.Handler {
    mux := http.NewServeMux()
    
    chain := middleware.ChainMiddleware(
        AuthMiddleware,      // Сначала аутентификация
        LoggingMiddleware,   // Потом логирование (пропустит неавторизованные запросы)
        CORSMiddleware,      // И в конце CORS
    )
    
    return chain(mux)
}

// ХОРОШО - правильный порядок
func GoodMiddlewareOrder() http.Handler {
    mux := http.NewServeMux()
    
    chain := middleware.ChainMiddleware(
        CORSMiddleware,      // Сначала CORS для preflight запросов
        LoggingMiddleware,   // Потом логирование всех запросов
        AuthMiddleware,      // И в конце аутентификация
    )
    
    return chain(mux)
}
```

### 2. Забытый вызов next.ServeHTTP

```go
// ПЛОХО - забыт next.ServeHTTP
func BadMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        // Забыли вызвать next.ServeHTTP - запрос не будет обработан!
        // next.ServeHTTP(w, r)
    })
}

// ХОРОШО - правильный вызов
func GoodMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r) // Не забываем вызвать следующий обработчик
    })
}
```

### 3. Модификация оригинального запроса

```go
// ПЛОХО - модификация оригинального запроса
func BadRequestModification(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Неправильно - модифицируем оригинальный запрос
        r.Header.Set("X-Modified", "true")
        next.ServeHTTP(w, r)
    })
}

// ХОРОШО - создание нового запроса
func GoodRequestModification(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Правильно - создаем новый запрос с модифицированным контекстом
        ctx := context.WithValue(r.Context(), "modified", true)
        newReq := r.WithContext(ctx)
        next.ServeHTTP(w, newReq)
    })
}
```

## Мониторинг и отладка middleware

### 1. Логирование производительности middleware

```go
// middleware/profiling.go
package middleware

import (
    "log"
    "net/http"
    "runtime"
    "time"
)

// ProfilingMiddleware профилирует производительность
func ProfilingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Получаем начальные метрики
        var m1, m2 runtime.MemStats
        runtime.ReadMemStats(&m1)
        
        // Выполняем обработчик
        lrw := newLoggingResponseWriter(w)
        next.ServeHTTP(lrw, r)
        
        // Получаем финальные метрики
        runtime.ReadMemStats(&m2)
        duration := time.Since(start)
        
        // Логируем профилировочную информацию
        log.Printf(
            "PERF: %s %s %d - Duration: %v, Alloc: %d KB, GC: %d",
            r.Method, r.URL.Path, lrw.statusCode, duration,
            (m2.Alloc-m1.Alloc)/1024, m2.NumGC-m1.NumGC,
        )
    })
}
```

### 2. Prometheus метрики для middleware

```go
// middleware/prometheus.go
package middleware

import (
    "net/http"
    "strconv"
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    requestCount = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)

// PrometheusMiddleware собирает метрики для Prometheus
func PrometheusMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        lrw := newLoggingResponseWriter(w)
        next.ServeHTTP(lrw, r)
        
        duration := time.Since(start).Seconds()
        
        // Записываем метрики
        requestCount.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(lrw.statusCode),
        ).Inc()
        
        requestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
        ).Observe(duration)
    })
}
```

## См. также

- [HTTP серверы](../concepts/http-server.md) - основы работы с HTTP
- [Контекст](../concepts/context.md) - управление жизненным циклом запросов
- [Тестирование](../concepts/testing.md) - как тестировать middleware
- [Профилирование](../concepts/profiling.md) - как измерять производительность
- [Практические примеры](../examples/middleware) - примеры кода