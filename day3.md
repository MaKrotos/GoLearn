# День 3: Стандартная библиотека (12 часов)

## Обязательные пакеты

### net/http — Handler, HandlerFunc, middleware цепочки

#### Основы HTTP сервера
```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    // Простой обработчик
    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Привет, мир!")
    })
    
    // Запуск сервера
    http.ListenAndServe(":8080", nil)
}
```

#### Handler интерфейс
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
    http.Handle("/hello", handler)
    http.ListenAndServe(":8080", nil)
}
```

#### HandlerFunc
HandlerFunc - это адаптер для использования обычных функций как HTTP обработчиков.

```go
func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Привет от HandlerFunc!")
}

func main() {
    http.HandleFunc("/hello", helloHandler)
    http.ListenAndServe(":8080", nil)
}
```

#### Middleware паттерн
Middleware - это функции, которые оборачивают обработчики для добавления дополнительной функциональности.

```go
// Middleware паттерн
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Проверка авторизации
        if r.Header.Get("Authorization") == "" {
            http.Error(w, "Не авторизован", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Публичная страница")
    })
    
    // Применяем middleware
    handler := loggingMiddleware(authMiddleware(mux))
    http.ListenAndServe(":8080", handler)
}
```

#### Цепочка middleware
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
    mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "API endpoint")
    })
    
    // Создаем цепочку middleware
    chain := chainMiddleware(loggingMiddleware, authMiddleware)
    handler := chain(mux)
    
    http.ListenAndServe(":8080", handler)
}
```

### encoding/json — теги структур, маршалинг/анмаршалинг

#### Основы работы с JSON
```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
)

type Person struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Email   string `json:"email,omitempty"`
    Address string `json:"-"` // Игнорировать это поле
}

func main() {
    // Маршалинг (преобразование в JSON)
    person := Person{
        Name:    "Иван",
        Age:     30,
        Email:   "ivan@example.com",
        Address: "Москва",
    }
    
    data, err := json.Marshal(person)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(data))
    // Вывод: {"name":"Иван","age":30,"email":"ivan@example.com"}
    
    // Анмаршалинг (преобразование из JSON)
    jsonStr := `{"name":"Мария","age":25,"email":"maria@example.com"}`
    var newPerson Person
    
    err = json.Unmarshal([]byte(jsonStr), &newPerson)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Имя: %s, Возраст: %d\n", newPerson.Name, newPerson.Age)
}
```

#### Работа с массивами и слайсами
```go
type People []Person

func main() {
    people := People{
        {Name: "Иван", Age: 30},
        {Name: "Мария", Age: 25},
    }
    
    data, err := json.Marshal(people)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(string(data))
    // Вывод: [{"name":"Иван","age":30},{"name":"Мария","age":25}]
}
```

#### Теги структур
```go
type User struct {
    ID        int    `json:"id"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    Password  string `json:"-"`                    // Не сериализовать
    Email     string `json:"email,omitempty"`      // Опустить если пустой
    CreatedAt string `json:"created_at,omitempty"` // Опустить если пустой
}
```

### database/sql — подключение, пулы, работа с контекстом

#### Основы подключения к базе данных
```go
package main

import (
    "context"
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/lib/pq" // Драйвер PostgreSQL
)

func main() {
    // Подключение к базе данных
    db, err := sql.Open("postgres", "user=user dbname=test sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Проверка подключения
    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Подключение к базе данных успешно!")
}
```

#### Работа с контекстом
```go
func getUserByID(ctx context.Context, db *sql.DB, id int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    row := db.QueryRowContext(ctx, query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func main() {
    // Создаем контекст с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    user, err := getUserByID(ctx, db, 1)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Пользователь: %+v\n", user)
}
```

#### Пулы соединений
```go
func main() {
    db, err := sql.Open("postgres", "user=user dbname=test sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Настройка пула соединений
    db.SetMaxOpenConns(25) // Максимальное количество открытых соединений
    db.SetMaxIdleConns(25) // Максимальное количество простаивающих соединений
    db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения
    
    // Использование соединения
    var name string
    err = db.QueryRow("SELECT name FROM users WHERE id = $1", 1).Scan(&name)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Имя пользователя:", name)
}
```

### testing — табличные тесты, бенчмарки, httptest

#### Основы тестирования
```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}

func Multiply(a, b int) int {
    return a * b
}
```

```go
// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5
    
    if result != expected {
        t.Errorf("Add(2, 3) = %d; expected %d", result, expected)
    }
}
```

#### Табличные тесты
```go
func TestAddTable(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -1, 1, 0},
        {"zero", 0, 0, 0},
        {"large numbers", 1000000, 2000000, 3000000},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

#### Бенчмарки
```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(1, 2)
    }
}

func BenchmarkMultiply(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Multiply(1, 2)
    }
}
```

#### httptest для тестирования HTTP обработчиков
```go
func TestHelloHandler(t *testing.T) {
    // Создаем тестовый сервер
    req := httptest.NewRequest("GET", "/hello", nil)
    w := httptest.NewRecorder()
    
    // Вызываем обработчик
    helloHandler(w, req)
    
    // Проверяем результат
    resp := w.Result()
    body, _ := io.ReadAll(resp.Body)
    
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK; got %v", resp.Status)
    }
    
    expected := "Привет, мир!"
    if string(body) != expected {
        t.Errorf("Expected %s; got %s", expected, string(body))
    }
}
```

## Паттерны для заучивания

### Middleware паттерн
```go
// Middleware паттерн
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // логика
        next.ServeHTTP(w, r)
    })
}
```

## Практические задания

1. Создайте HTTP сервер с несколькими маршрутами и middleware для логирования и аутентификации.
2. Реализуйте структуру с тегами JSON и выполните маршалинг/анмаршалинг.
3. Подключитесь к базе данных и выполните несколько запросов с использованием контекста.
4. Напишите табличные тесты для нескольких функций.
5. Создайте бенчмарки для сравнения производительности различных операций.