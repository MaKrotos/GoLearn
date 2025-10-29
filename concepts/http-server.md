# HTTP сервер в Go - объяснение для чайников

## Что такое HTTP сервер?

Представьте HTTP сервер как **почтальона**, который:
- **Получает письма** (HTTP запросы) от клиентов
- **Сортирует их** (определяет, куда доставить)
- **Доставляет по адресам** (вызывает нужные функции)
- **Отправляет ответы** (возвращает результаты клиентам)

## Базовый HTTP сервер

### Простейший сервер

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    // Регистрируем обработчик для пути "/"
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Привет, мир!") // Отправляем ответ
    })
    
    // Запускаем сервер на порту 8080
    http.ListenAndServe(":8080", nil)
}
```

### Как это работает:

1. **http.HandleFunc("/", ...)** - регистрируем функцию для обработки запросов по пути "/"
2. **http.ListenAndServe(":8080", nil)** - запускаем сервер на порту 8080
3. Когда кто-то заходит на http://localhost:8080/, вызывается наша функция

## Handler и HandlerFunc

### Handler интерфейс

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Пример реализации:

```go
type HelloHandler struct{}

func (h HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Привет от Handler!")
}

func main() {
    var handler HelloHandler
    http.Handle("/hello", handler) // Регистрируем Handler
    http.ListenAndServe(":8080", nil)
}
```

### HandlerFunc

HandlerFunc - это **адаптер**, который позволяет использовать обычные функции как Handler:

```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Привет от HandlerFunc!")
}

func main() {
    http.HandleFunc("/hello", helloHandler) // Регистрируем HandlerFunc
    http.ListenAndServe(":8080", nil)
}
```

## Middleware

### Что такое middleware?

Middleware - это **функции-обертки**, которые:
- **Выполняются до** основного обработчика
- **Могут модифицировать** запрос или ответ
- **Могут прервать** обработку запроса
- **Вызывают** следующий обработчик в цепочке

### Примеры middleware:

#### Логирование

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

#### Аутентификация

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Проверяем заголовок авторизации
        if r.Header.Get("Authorization") == "" {
            http.Error(w, "Не авторизован", http.StatusUnauthorized)
            return // Прерываем обработку
        }
        
        next.ServeHTTP(w, r) // Продолжаем обработку
    })
}
```

#### Восстановление после паники

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                fmt.Printf("Паника: %v\n", err)
                http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
            }
        }()
        
        next.ServeHTTP(w, r)
    })
}
```

### Цепочка middleware

```go
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api", apiHandler)
    
    // Создаем цепочку middleware
    var handler http.Handler = mux
    handler = recoveryMiddleware(handler)
    handler = loggingMiddleware(handler)
    handler = authMiddleware(handler)
    
    http.ListenAndServe(":8080", handler)
}
```

## Router (Маршрутизатор)

### http.ServeMux

```go
func main() {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/", homeHandler)
    mux.HandleFunc("/users", usersHandler)
    mux.HandleFunc("/users/", userHandler) // Обрабатывает /users/123
    
    http.ListenAndServe(":8080", mux)
}
```

### Параметры в URL

```go
func userHandler(w http.ResponseWriter, r *http.Request) {
    // Для пути /users/123 получаем "123"
    userID := strings.TrimPrefix(r.URL.Path, "/users/")
    fmt.Fprintf(w, "Пользователь ID: %s", userID)
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
    body, _ := io.ReadAll(r.Body)
    
    fmt.Fprintf(w, "Method: %s\nName: %s\nAge: %s\nContent-Type: %s\nBody: %s",
        method, name, age, contentType, body)
}
```

### Работа с формами

```go
func formHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        // Парсим форму
        err := r.ParseForm()
        if err != nil {
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
    
    json.NewEncoder(w).Encode(response)
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

## Практический пример: REST API

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

var users = []User{
    {ID: 1, Name: "Иван", Age: 30},
    {ID: 2, Name: "Мария", Age: 25},
}

func main() {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/users", usersHandler)
    mux.HandleFunc("/users/", userHandler)
    
    // Добавляем middleware
    handler := loggingMiddleware(mux)
    
    fmt.Println("Сервер запущен на :8080")
    http.ListenAndServe(":8080", handler)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
        
    case "POST":
        var user User
        if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
            http.Error(w, "Неверный JSON", http.StatusBadRequest)
            return
        }
        
        user.ID = len(users) + 1
        users = append(users, user)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(user)
        
    default:
        http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
    }
}

func userHandler(w http.ResponseWriter, r *http.Request) {
    // Извлекаем ID из URL
    path := strings.TrimPrefix(r.URL.Path, "/users/")
    id, err := strconv.Atoi(path)
    if err != nil {
        http.Error(w, "Неверный ID", http.StatusBadRequest)
        return
    }
    
    // Ищем пользователя
    var user *User
    for i := range users {
        if users[i].ID == id {
            user = &users[i]
            break
        }
    }
    
    if user == nil {
        http.Error(w, "Пользователь не найден", http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

## Настройка сервера

### Таймауты

```go
func main() {
    server := &http.Server{
        Addr:         ":8080",
        Handler:      mux,
        ReadTimeout:  5 * time.Second,  // Таймаут чтения
        WriteTimeout: 10 * time.Second, // Таймаут записи
        IdleTimeout:  60 * time.Second, // Таймаут простоя
    }
    
    server.ListenAndServe()
}
```

### Graceful shutdown

```go
func main() {
    server := &http.Server{Addr: ":8080", Handler: mux}
    
    // Горутина для graceful shutdown
    go func() {
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, os.Interrupt)
        <-sigChan
        
        fmt.Println("Получен сигнал завершения...")
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        server.Shutdown(ctx)
    }()
    
    server.ListenAndServe()
}
```

## См. также

- [Context](context.md) - как использовать контексты в HTTP обработчиках
- [Middleware паттерн](../theory/middleware.md) - более подробно о middleware
- [JSON в Go](../theory/json.md) - работа с JSON
- [Тестирование HTTP серверов](../theory/http-testing.md) - как тестировать серверы