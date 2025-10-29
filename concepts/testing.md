# Тестирование в Go - объяснение для чайников

## Что такое тестирование?

Представьте, что вы пекарь. Перед тем как продавать торт, вы **проверяете**:
- Вкусен ли он?
- Правильная ли текстура?
- Все ли ингредиенты на месте?

Тестирование в программировании - это **проверка**, что ваш код работает правильно.

## Основы тестирования в Go

### Структура тестов

В Go тесты пишутся в отдельных файлах с суффиксом `_test.go`:

```
math.go          // Основной код
math_test.go     // Тесты для math.go
```

### Простой тест

```go
// math.go
package math

func Add(a, b int) int {
    return a + b
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

### Как запустить тесты

```bash
# Запустить все тесты
go test

# Запустить с подробным выводом
go test -v

# Запустить тесты в определенном файле
go test math_test.go math.go
```

## Табличные тесты

### Что такое табличные тесты?

Табличные тесты - это **один тест для множества случаев**. Представьте таблицу:

| Входные данные | Ожидаемый результат |
|----------------|---------------------|
| 2 + 3          | 5                   |
| -1 + 1         | 0                   |
| 0 + 0          | 0                   |

### Пример табличного теста

```go
func TestAddTable(t *testing.T) {
    tests := []struct {
        name     string // Имя теста
        a, b     int    // Входные данные
        expected int    // Ожидаемый результат
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

### Преимущества табличных тестов:

1. **Легко добавлять новые случаи**
2. **Единообразие** в тестировании
3. **Читаемость** - все случаи в одном месте
4. **Параллельное выполнение** подтестов

## httptest для HTTP обработчиков

### Что такое httptest?

`httptest` - это **инструмент для тестирования HTTP обработчиков** без запуска реального сервера.

### Пример тестирования HTTP обработчика

```go
// handlers/user.go
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    if userID == "" {
        http.Error(w, "ID обязателен", http.StatusBadRequest)
        return
    }
    
    // Здесь была бы логика получения пользователя
    user := map[string]string{
        "id":   userID,
        "name": "Иван Иванов",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

```go
// handlers/user_test.go
package handlers

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetUserHandler(t *testing.T) {
    // Создаем тестовый запрос
    req := httptest.NewRequest("GET", "/user?id=123", nil)
    
    // Создаем тестовый ResponseRecorder
    w := httptest.NewRecorder()
    
    // Вызываем обработчик
    GetUserHandler(w, req)
    
    // Проверяем результат
    resp := w.Result()
    
    // Проверяем статус
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status OK; got %v", resp.Status)
    }
    
    // Проверяем Content-Type
    if resp.Header.Get("Content-Type") != "application/json" {
        t.Errorf("Expected Content-Type application/json; got %s", resp.Header.Get("Content-Type"))
    }
    
    // Проверяем тело ответа
    var user map[string]string
    err := json.NewDecoder(resp.Body).Decode(&user)
    if err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }
    
    if user["id"] != "123" {
        t.Errorf("Expected user ID 123; got %s", user["id"])
    }
}
```

## Моки (Mocks)

### Что такое моки?

Моки - это **подделки** реальных объектов для тестирования. Представьте:
- Вместо настоящего сервера - **мок сервер**
- Вместо базы данных - **мок хранилище**
- Вместо внешнего API - **мок клиента**

### Пример мока репозитория

```go
// repository/user.go
type User struct {
    ID   int
    Name string
}

type UserRepository interface {
    GetUserByID(id int) (*User, error)
    SaveUser(user *User) error
}

type DBUserRepository struct {
    db *sql.DB
}

func (r *DBUserRepository) GetUserByID(id int) (*User, error) {
    // Реальная реализация с базой данных
    // ...
}
```

```go
// mocks/user_repository.go
package mocks

import "yourproject/repository"

type MockUserRepository struct {
    GetUserByIDFunc func(id int) (*repository.User, error)
    SaveUserFunc    func(user *repository.User) error
}

func (m *MockUserRepository) GetUserByID(id int) (*repository.User, error) {
    if m.GetUserByIDFunc != nil {
        return m.GetUserByIDFunc(id)
    }
    return nil, nil
}

func (m *MockUserRepository) SaveUser(user *repository.User) error {
    if m.SaveUserFunc != nil {
        return m.SaveUserFunc(user)
    }
    return nil
}
```

### Использование мока в тестах

```go
// service/user_test.go
func TestUserService_GetUser(t *testing.T) {
    // Создаем мок репозитория
    mockRepo := &mocks.MockUserRepository{
        GetUserByIDFunc: func(id int) (*repository.User, error) {
            if id == 1 {
                return &repository.User{ID: 1, Name: "Иван"}, nil
            }
            return nil, errors.New("Пользователь не найден")
        },
    }
    
    // Создаем сервис с моком
    service := NewUserService(mockRepo)
    
    // Тестируем
    user, err := service.GetUser(1)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if user.Name != "Иван" {
        t.Errorf("Expected name 'Иван'; got '%s'", user.Name)
    }
}
```

## Бенчмарки

### Что такое бенчмарки?

Бенчмарки - это **тесты производительности**. Они измеряют:
- **Время выполнения**
- **Использование памяти**
- **Аллокации**

### Пример бенчмарка

```go
// math.go
func Fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return Fibonacci(n-1) + Fibonacci(n-2)
}

func FibonacciIterative(n int) int {
    if n <= 1 {
        return n
    }
    
    a, b := 0, 1
    for i := 2; i <= n; i++ {
        a, b = b, a+b
    }
    return b
}
```

```go
// math_test.go
func BenchmarkFibonacci(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Fibonacci(20)
    }
}

func BenchmarkFibonacciIterative(b *testing.B) {
    for i := 0; i < b.N; i++ {
        FibonacciIterative(20)
    }
}
```

### Запуск бенчмарков

```bash
# Запустить бенчмарки
go test -bench=.

# Запустить с измерением памяти
go test -bench=. -benchmem

# Запустить определенный бенчмарк
go test -bench=BenchmarkFibonacci
```

### Результаты бенчмарков

```
BenchmarkFibonacci-8              30000         40200 ns/op
BenchmarkFibonacciIterative-8    20000000      65.2 ns/op
```

Это означает:
- `Fibonacci` выполняется 30000 раз, среднее время 40200 наносекунд
- `FibonacciIterative` выполняется 20000000 раз, среднее время 65.2 наносекунды

## Тестирование с гонками данных

### Что такое гонки данных?

Гонки данных происходят когда **несколько горутин** одновременно:
- **Читают и пишут** одну переменную
- **Без синхронизации**

### Как обнаружить гонки?

```bash
go test -race
```

### Пример кода с гонкой

```go
// ПЛОХОЙ код с гонкой
var counter int

func TestRace(t *testing.T) {
    for i := 0; i < 1000; i++ {
        go func() {
            counter++ // Гонка данных!
        }()
    }
    
    time.Sleep(time.Second)
    fmt.Println(counter) // Результат будет разным каждый раз
}
```

### Исправленный код

```go
// ХОРОШИЙ код без гонки
var (
    counter int
    mutex   sync.Mutex
)

func TestNoRace(t *testing.T) {
    var wg sync.WaitGroup
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mutex.Lock()
            counter++
            mutex.Unlock()
        }()
    }
    
    wg.Wait()
    fmt.Println(counter) // Всегда будет 1000
}
```

## Практические советы

### 1. Пишите тесты вместе с кодом

```go
// Сначала напишите тест
func TestCalculateTax(t *testing.T) {
    // ...
}

// Потом реализуйте функцию
func CalculateTax(income float64) float64 {
    // ...
}
```

### 2. Используйте вспомогательные функции

```go
func createTestUser(id int, name string) *User {
    return &User{ID: id, Name: name}
}

func assertUserEqual(t *testing.T, expected, actual *User) {
    if expected.ID != actual.ID || expected.Name != actual.Name {
        t.Errorf("Expected user %+v; got %+v", expected, actual)
    }
}
```

### 3. Тестируйте граничные случаи

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name string
        a, b float64
        want float64
        err  bool
    }{
        {"normal", 10, 2, 5, false},
        {"zero dividend", 0, 5, 0, false},
        {"zero divisor", 5, 0, 0, true}, // Ошибка деления на ноль
        {"negative", -10, 2, -5, false},
    }
    
    // ...
}
```

### 4. Используйте testify для удобства

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestWithTestify(t *testing.T) {
    result := Add(2, 3)
    
    // assert не останавливает тест при ошибке
    assert.Equal(t, 5, result, "2 + 3 should equal 5")
    
    // require останавливает тест при ошибке
    require.Equal(t, 5, result, "2 + 3 should equal 5")
    
    // Более читаемые сообщения об ошибках
    assert.NoError(t, err, "Should not return an error")
    assert.NotNil(t, user, "User should not be nil")
}
```

## См. также

- [Интерфейсы](interface.md) - как использовать для моков
- [Контекст](context.md) - как тестировать код с контекстами
- [HTTP серверы](http-server.md) - как тестировать веб-приложения
- [Базы данных](database.md) - как тестировать работу с БД