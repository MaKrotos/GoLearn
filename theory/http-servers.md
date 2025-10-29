# HTTP серверы в Go: Полная теория

## Введение в HTTP серверы

### Что такое HTTP сервер?

HTTP сервер в Go - это **приложение**, которое:
- **Слушает** сетевые соединения на определенном порту
- **Принимает** HTTP запросы
- **Обрабатывает** запросы с помощью обработчиков
- **Отправляет** HTTP ответы

### Основные компоненты

```go
import "net/http"

func main() {
    // 1. Регистрация обработчиков
    http.HandleFunc("/", handler)
    
    // 2. Запуск сервера
    http.ListenAndServe(":8080", nil)
}
```

## Обработчики (Handlers)

### Интерфейс Handler

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

#### Реализация Handler

```go
type HelloHandler struct{}

func (h HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Привет от Handler!")
}

func main() {
    var handler HelloHandler
    http.Handle("/hello", handler)
    http.ListenAndServe(":8080", nil)
}
```

### HandlerFunc

HandlerFunc - это **адаптер** для использования обычных функций как HTTP обработчиков:

```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Привет от HandlerFunc!")
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    http.ListenAndServe(":8080", nil)
}
```

### ServeMux

ServeMux - это **мультиплексор** HTTP запросов:

```go
func main() {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/", homeHandler)
    mux.HandleFunc("/users", usersHandler)
    mux.HandleFunc("/users/", userHandler)
    
    http.ListenAndServe(":8080", mux)
}
```

## Middleware

### Что такое middleware?

Middleware - это **функции-обертки**, которые:
- **Выполняются до** или **после** основного обработчика
- **Могут модифицировать** запрос или ответ
- **Могут прервать** обработку запроса

### Базовый middleware

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        fmt.Printf("[%s] %s %s\n", start.Format("2006-01-02 15:04:05"), r.Method, r.URL.Path)
        
        next.ServeHTTP(w, r) // Вызываем следующий обработчик
        
        fmt.Printf("Завершено за %v\n", time.Since(start))
    })
}
```

### Цепочка middleware

```go
func chainMiddleware(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
    return func(final http.Handler) http.Handler {
        for i := len(middlewares) - 1; i >= 0; i-- {
            final = middlewares[i](final)
        }
        return final
    }
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api", apiHandler)
    
    // Создаем цепочку middleware
    chain := chainMiddleware(loggingMiddleware, authMiddleware, recoveryMiddleware)
    handler := chain(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### Распространенные middleware

#### 1. Логирование

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Создаем обертку для ResponseWriter для отслеживания статуса
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(wrapped, r)
        
        log.Printf(
            "%s %s %d %v",
            r.Method,
            r.URL.Path,
            wrapped.statusCode,
            time.Since(start),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```

#### 2. Аутентификация

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Проверяем заголовок авторизации
        auth := r.Header.Get("Authorization")
        if auth == "" {
            http.Error(w, "Не авторизован", http.StatusUnauthorized)
            return
        }
        
        // Проверяем токен
        if !isValidToken(auth) {
            http.Error(w, "Неверный токен", http.StatusUnauthorized)
            return
        }
        
        // Добавляем информацию о пользователе в контекст
        ctx := context.WithValue(r.Context(), "user", extractUserFromToken(auth))
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

#### 3. CORS

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

#### 4. Rate Limiting

```go
type rateLimiter struct {
    visitors map[string]*visitor
    mutex    sync.RWMutex
}

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

func (rl *rateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ip := r.RemoteAddr
        
        rl.mutex.Lock()
        v, exists := rl.visitors[ip]
        if !exists {
            // Ограничиваем 10 запросов в минуту
            limiter := rate.NewLimiter(10, 10)
            rl.visitors[ip] = &visitor{limiter, time.Now()}
            v = rl.visitors[ip]
        } else {
            v.lastSeen = time.Now()
        }
        rl.mutex.Unlock()
        
        if !v.limiter.Allow() {
            http.Error(w, "Слишком много запросов", http.StatusTooManyRequests)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Работа с запросами

### Получение данных из запроса

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Метод запроса
    method := r.Method
    
    // Параметры URL (?name=John&age=30)
    name := r.URL.Query().Get("name")
    age := r.URL.Query().Get("age")
    
    // Заголовки
    contentType := r.Header.Get("Content-Type")
    userAgent := r.Header.Get("User-Agent")
    
    // Тело запроса (для POST, PUT)
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
        return
    }
    
    fmt.Fprintf(w, "Method: %s\nName: %s\nAge: %s\nContent-Type: %s\nBody: %s",
        method, name, age, contentType, body)
}
```

### Работа с формами

```go
func formHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        // Парсим форму
        if err := r.ParseForm(); err != nil {
            http.Error(w, "Ошибка парсинга формы", http.StatusBadRequest)
            return
        }
        
        // Получаем значения
        name := r.FormValue("name")
        email := r.FormValue("email")
        
        fmt.Fprintf(w, "Имя: %s, Email: %s", name, email)
    } else {
        // Показываем форму
        html := `
        <form method="POST">
            <input type="text" name="name" placeholder="Имя">
            <input type="email" name="email" placeholder="Email">
            <button type="submit">Отправить</button>
        </form>
        `
        fmt.Fprint(w, html)
    }
}
```

### Работа с JSON

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        // Отправляем JSON
        user := User{ID: 1, Name: "Иван", Age: 30}
        
        w.Header().Set("Content-Type", "application/json")
        if err := json.NewEncoder(w).Encode(user); err != nil {
            http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
            return
        }
        
    case "POST":
        // Получаем JSON
        var user User
        if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
            http.Error(w, "Неверный JSON", http.StatusBadRequest)
            return
        }
        
        // Обрабатываем данные
        fmt.Fprintf(w, "Получен пользователь: %+v", user)
        
    default:
        http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
    }
}
```

## Отправка ответов

### Простой ответ

```go
func simpleHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK) // Код ответа 200
    fmt.Fprint(w, "OK")
}
```

### JSON ответ

```go
func jsonHandler(w http.ResponseWriter, r *http.Request) {
    // Устанавливаем заголовок Content-Type
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    
    // Создаем JSON
    response := map[string]interface{}{
        "status": "success",
        "data":   []int{1, 2, 3, 4, 5},
    }
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
        return
    }
}
```

### Ошибки

```go
func errorHandler(w http.ResponseWriter, r *http.Request) {
    // 404 Not Found
    http.Error(w, "Страница не найдена", http.StatusNotFound)
    
    // Или с кастомным кодом
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprint(w, "Внутренняя ошибка сервера")
}
```

## Настройка сервера

### Таймауты

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)
    
    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  5 * time.Second,  // Таймаут чтения
        WriteTimeout: 10 * time.Second, // Таймаут записи
        IdleTimeout:  60 * time.Second, // Таймаут простоя
    }
    
    log.Fatal(server.ListenAndServe())
}
```

### Graceful shutdown

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
    }
    
    // Горутина для graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
        <-sigChan
        
        log.Println("Получен сигнал завершения...")
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := server.Shutdown(ctx); err != nil {
            log.Fatalf("Ошибка graceful shutdown: %v", err)
        }
    }()
    
    log.Println("Сервер запущен на :8080")
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Ошибка запуска сервера: %v", err)
    }
}
```

### HTTPS

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)
    
    log.Fatal(http.ListenAndServeTLS(":443", "cert.pem", "key.pem", mux))
}
```

## Практический пример: REST API

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type UserAPI struct {
    users map[int]User
    mutex sync.RWMutex
    nextID int
}

func NewUserAPI() *UserAPI {
    return &UserAPI{
        users:  make(map[int]User),
        nextID: 1,
    }
}

func (api *UserAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        api.getUsers(w, r)
    case "POST":
        api.createUser(w, r)
    case "PUT":
        api.updateUser(w, r)
    case "DELETE":
        api.deleteUser(w, r)
    default:
        http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
    }
}

func (api *UserAPI) getUsers(w http.ResponseWriter, r *http.Request) {
    api.mutex.RLock()
    defer api.mutex.RUnlock()
    
    users := make([]User, 0, len(api.users))
    for _, user := range api.users {
        users = append(users, user)
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}

func (api *UserAPI) createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "Неверный JSON", http.StatusBadRequest)
        return
    }
    
    if user.Name == "" || user.Email == "" {
        http.Error(w, "Имя и email обязательны", http.StatusBadRequest)
        return
    }
    
    api.mutex.Lock()
    user.ID = api.nextID
    api.nextID++
    api.users[user.ID] = user
    api.mutex.Unlock()
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

func main() {
    api := NewUserAPI()
    
    // Добавляем middleware
    handler := loggingMiddleware(corsMiddleware(api))
    
    server := &http.Server{
        Addr:    ":8080",
        Handler: handler,
    }
    
    log.Println("Сервер запущен на :8080")
    log.Fatal(server.ListenAndServe())
}
```

## Тестирование HTTP серверов

### Использование httptest

```go
func TestUserAPI(t *testing.T) {
    api := NewUserAPI()
    server := httptest.NewServer(loggingMiddleware(api))
    defer server.Close()
    
    // Тест создания пользователя
    user := User{Name: "Иван", Email: "ivan@example.com"}
    jsonData, _ := json.Marshal(user)
    
    resp, err := http.Post(server.URL+"/users", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        t.Fatalf("Ошибка запроса: %v", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        t.Errorf("Ожидаемый статус %d, получено %d", http.StatusCreated, resp.StatusCode)
    }
    
    var createdUser User
    if err := json.NewDecoder(resp.Body).Decode(&createdUser); err != nil {
        t.Fatalf("Ошибка декодирования: %v", err)
    }
    
    if createdUser.Name != "Иван" {
        t.Errorf("Ожидаемое имя 'Иван', получено '%s'", createdUser.Name)
    }
}
```

### Моки для внешних сервисов

```go
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

type MockHTTPClient struct {
    DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    if m.DoFunc != nil {
        return m.DoFunc(req)
    }
    return nil, nil
}

func TestExternalServiceCall(t *testing.T) {
    mockClient := &MockHTTPClient{
        DoFunc: func(req *http.Request) (*http.Response, error) {
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
            }, nil
        },
    }
    
    // Используем mockClient в тестируемом коде
}
```

## Распространенные ошибки

### 1. Игнорирование ошибок

```go
// ПЛОХО
json.NewEncoder(w).Encode(data)

// ЛУЧШЕ
if err := json.NewEncoder(w).Encode(data); err != nil {
    log.Printf("Ошибка кодирования JSON: %v", err)
    http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
    return
}
```

### 2. Небезопасная обработка тела запроса

```go
// ПЛОХО
body, _ := io.ReadAll(r.Body) // Игнорируем ошибки

// ЛУЧШЕ
body, err := io.ReadAll(r.Body)
if err != nil {
    http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
    return
}
```

### 3. Отсутствие таймаутов

```go
// ПЛОХО
http.ListenAndServe(":8080", nil) // Нет таймаутов

// ЛУЧШЕ
server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
}
```

## Лучшие практики

### 1. Валидация входных данных

```go
func validateUser(user User) error {
    if user.Name == "" {
        return errors.New("имя не может быть пустым")
    }
    
    if user.Email == "" {
        return errors.New("email не может быть пустым")
    }
    
    if !isValidEmail(user.Email) {
        return errors.New("неверный формат email")
    }
    
    return nil
}
```

### 2. Логирование и мониторинг

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        next.ServeHTTP(wrapped, r)
        
        log.Printf(
            "%s %s %d %v %s",
            r.Method,
            r.URL.Path,
            wrapped.statusCode,
            time.Since(start),
            r.RemoteAddr,
        )
    })
}
```

### 3. Обработка паник

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Паника: %v\n%s", err, debug.Stack())
                http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

## См. также

- [HTTP серверы для чайников](../concepts/http-server.md) - базовое объяснение
- [Контекст](../concepts/context.md) - использование контекста в HTTP обработчиках
- [Тестирование](../concepts/testing.md) - как тестировать HTTP серверы
- [Middleware паттерн](middleware.md) - подробнее о middleware
- [JSON в Go](json.md) - работа с JSON
- [Практические примеры](../examples/http-servers) - примеры кода