# Тестирование в Go: Полная теория

## Введение в тестирование

### Что такое тестирование?

Тестирование в Go - это **процесс проверки**, что ваш код работает правильно. Go имеет **встроенную поддержку** тестирования.

### Типы тестов

1. **Модульные тесты** (Unit tests) - тестирование отдельных функций
2. **Интеграционные тесты** (Integration tests) - тестирование взаимодействия компонентов
3. **Бенчмарки** (Benchmarks) - измерение производительности
4. **Примеры** (Examples) - документация в виде исполняемого кода

## Основы тестирования

### Структура тестов

Тесты пишутся в файлах с суффиксом `_test.go`:

```
project/
├── math.go          # Основной код
├── math_test.go     # Тесты для math.go
└── calc/
    ├── calc.go
    └── calc_test.go
```

### Базовый тест

```go
// math.go
package math

import "errors"

func Add(a, b int) int {
    return a + b
}

func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("деление на ноль")
    }
    return a / b, nil
}
```

```go
// math_test.go
package math

import (
    "strings"
    "testing"
)

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5
    
    if result != expected {
        t.Errorf("Add(2, 3) = %d; expected %d", result, expected)
    }
}

func TestDivide(t *testing.T) {
    result, err := Divide(10, 2)
    if err != nil {
        t.Fatalf("Неожиданная ошибка: %v", err)
    }
    
    expected := 5.0
    if result != expected {
        t.Errorf("Divide(10, 2) = %f; expected %f", result, expected)
    }
}

func TestDivideByZero(t *testing.T) {
    _, err := Divide(10, 0)
    if err == nil {
        t.Error("Ожидалась ошибка деления на ноль")
    }
}
```

### Запуск тестов

```bash
# Запустить все тесты
go test

# Запустить с подробным выводом
go test -v

# Запустить тесты в определенном файле
go test math_test.go math.go

# Запустить определенные тесты
go test -run TestAdd

# Запустить тесты в подкаталогах
go test ./...

# Запустить с покрытием кода
go test -cover

# Запустить с детальным покрытием
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Табличные тесты

### Что такое табличные тесты?

Табличные тесты - это **один тест для множества случаев**:

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

### Расширенные табличные тесты

```go
func TestDivideTable(t *testing.T) {
    tests := []struct {
        name        string
        a, b        float64
        expected    float64
        expectError bool
        errorText   string
    }{
        {"normal division", 10, 2, 5, false, ""},
        {"division by zero", 10, 0, 0, true, "деление на ноль"},
        {"negative numbers", -10, 2, -5, false, ""},
        {"zero dividend", 0, 5, 0, false, ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Divide(tt.a, tt.b)
            
            // Проверяем ошибку
            if tt.expectError {
                if err == nil {
                    t.Fatal("Ожидалась ошибка, но ее нет")
                }
                if !strings.Contains(err.Error(), tt.errorText) {
                    t.Errorf("Ожидалась ошибка с текстом '%s', получено '%s'", 
                        tt.errorText, err.Error())
                }
            } else {
                if err != nil {
                    t.Fatalf("Неожиданная ошибка: %v", err)
                }
                if result != tt.expected {
                    t.Errorf("Divide(%f, %f) = %f; expected %f", 
                        tt.a, tt.b, result, tt.expected)
                }
            }
        })
    }
}
```

## httptest для HTTP обработчиков

### Что такое httptest?

`httptest` - это **инструмент для тестирования HTTP обработчиков** без запуска реального сервера.

### Тестирование HTTP обработчиков

```go
// handlers/user.go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    if userID == "" {
        http.Error(w, "ID обязателен", http.StatusBadRequest)
        return
    }
    
    id, err := strconv.Atoi(userID)
    if err != nil {
        http.Error(w, "Неверный ID", http.StatusBadRequest)
        return
    }
    
    // Здесь была бы логика получения пользователя
    user := User{ID: id, Name: "Иван Иванов"}
    
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
    tests := []struct {
        name           string
        url            string
        expectedStatus int
        expectedName   string
    }{
        {"valid user", "/user?id=1", http.StatusOK, "Иван Иванов"},
        {"missing id", "/user", http.StatusBadRequest, ""},
        {"invalid id", "/user?id=abc", http.StatusBadRequest, ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Создаем тестовый запрос
            req := httptest.NewRequest("GET", tt.url, nil)
            
            // Создаем тестовый ResponseRecorder
            w := httptest.NewRecorder()
            
            // Вызываем обработчик
            GetUserHandler(w, req)
            
            // Проверяем результат
            resp := w.Result()
            
            // Проверяем статус
            if resp.StatusCode != tt.expectedStatus {
                t.Errorf("Expected status %d; got %d", tt.expectedStatus, resp.StatusCode)
            }
            
            // Проверяем тело ответа, если ожидается успех
            if tt.expectedStatus == http.StatusOK {
                var user User
                err := json.NewDecoder(resp.Body).Decode(&user)
                if err != nil {
                    t.Fatalf("Failed to decode response: %v", err)
                }
                
                if user.Name != tt.expectedName {
                    t.Errorf("Expected name '%s'; got '%s'", tt.expectedName, user.Name)
                }
            }
        })
    }
}
```

## Моки (Mocks)

### Что такое моки?

Моки - это **подделки** реальных объектов для тестирования.

### Создание моков

```go
// repository/user.go
package repository

import "errors"

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
    return nil, errors.New("не реализовано")
}

func (r *DBUserRepository) SaveUser(user *User) error {
    // Реальная реализация с базой данных
    return errors.New("не реализовано")
}
```

```go
// mocks/user_repository.go
package mocks

import (
    "errors"
    "yourproject/repository"
)

type MockUserRepository struct {
    users map[int]*repository.User
    GetError error
    SaveError error
}

func NewMockUserRepository() *MockUserRepository {
    return &MockUserRepository{
        users: make(map[int]*repository.User),
    }
}

func (m *MockUserRepository) WithUser(user *repository.User) *MockUserRepository {
    m.users[user.ID] = user
    return m
}

func (m *MockUserRepository) WithGetError(err error) *MockUserRepository {
    m.GetError = err
    return m
}

func (m *MockUserRepository) WithSaveError(err error) *MockUserRepository {
    m.SaveError = err
    return m
}

func (m *MockUserRepository) GetUserByID(id int) (*repository.User, error) {
    if m.GetError != nil {
        return nil, m.GetError
    }
    
    user, exists := m.users[id]
    if !exists {
        return nil, errors.New("пользователь не найден")
    }
    
    return user, nil
}

func (m *MockUserRepository) SaveUser(user *repository.User) error {
    if m.SaveError != nil {
        return m.SaveError
    }
    
    m.users[user.ID] = user
    return nil
}
```

### Использование моков в тестах

```go
// service/user.go
package service

import "yourproject/repository"

type UserService struct {
    repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*repository.User, error) {
    return s.repo.GetUserByID(id)
}

func (s *UserService) CreateUser(name string) (*repository.User, error) {
    user := &repository.User{Name: name}
    err := s.repo.SaveUser(user)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

```go
// service/user_test.go
package service

import (
    "testing"
    "yourproject/mocks"
    "yourproject/repository"
)

func TestUserService_GetUser(t *testing.T) {
    // Создаем мок репозитория с тестовыми данными
    mockRepo := mocks.NewMockUserRepository().
        WithUser(&repository.User{ID: 1, Name: "Иван"})
    
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

func TestUserService_GetUser_NotFound(t *testing.T) {
    // Создаем мок репозитория без данных
    mockRepo := mocks.NewMockUserRepository()
    
    service := NewUserService(mockRepo)
    
    _, err := service.GetUser(999)
    if err == nil {
        t.Fatal("Expected error, got nil")
    }
    
    if err.Error() != "пользователь не найден" {
        t.Errorf("Expected 'пользователь не найден' error; got '%s'", err.Error())
    }
}

func TestUserService_CreateUser(t *testing.T) {
    mockRepo := mocks.NewMockUserRepository()
    service := NewUserService(mockRepo)
    
    user, err := service.CreateUser("Мария")
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if user.Name != "Мария" {
        t.Errorf("Expected name 'Мария'; got '%s'", user.Name)
    }
    
    // Проверяем, что пользователь сохранился в репозитории
    savedUser, err := mockRepo.GetUserByID(user.ID)
    if err != nil {
        t.Fatalf("User not saved in repository: %v", err)
    }
    
    if savedUser.Name != "Мария" {
        t.Errorf("Expected saved name 'Мария'; got '%s'", savedUser.Name)
    }
}
```

## Бенчмарки

### Что такое бенчмарки?

Бенчмарки - это **тесты производительности**, которые измеряют:
- **Время выполнения**
- **Использование памяти**
- **Аллокации**

### Создание бенчмарков

```go
// math.go
package math

import "strings"

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

func ConcatenateStrings(strs []string) string {
    result := ""
    for _, s := range strs {
        result += s
    }
    return result
}

func ConcatenateStringsBuilder(strs []string) string {
    var builder strings.Builder
    for _, s := range strs {
        builder.WriteString(s)
    }
    return builder.String()
}
```

```go
// math_test.go
package math

import (
    "strings"
    "testing"
)

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

func BenchmarkConcatenateStrings(b *testing.B) {
    strs := []string{"hello", " ", "world", "!", " ", "Go", " ", "is", " ", "awesome"}
    
    for i := 0; i < b.N; i++ {
        ConcatenateStrings(strs)
    }
}

func BenchmarkConcatenateStringsBuilder(b *testing.B) {
    strs := []string{"hello", " ", "world", "!", " ", "Go", " ", "is", " ", "awesome"}
    
    for i := 0; i < b.N; i++ {
        ConcatenateStringsBuilder(strs)
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

# Запустить с профилированием CPU
go test -bench=. -cpuprofile=cpu.prof

# Запустить с профилированием памяти
go test -bench=. -memprofile=mem.prof
```

### Результаты бенчмарков

```
BenchmarkFibonacci-8                    30000         40200 ns/op
BenchmarkFibonacciIterative-8        20000000      65.2 ns/op
BenchmarkConcatenateStrings-8          500000      3200 ns/op     480 B/op     10 allocs/op
BenchmarkConcatenateStringsBuilder-8  5000000       320 ns/op      64 B/op      1 allocs/op
```

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
// counter.go
package counter

var globalCounter int

func Increment() {
    globalCounter++ // Гонка данных!
}

func Get() int {
    return globalCounter
}
```

```go
// counter_test.go
package counter

import (
    "sync"
    "testing"
)

func TestRace(t *testing.T) {
    globalCounter = 0
    
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            Increment() // Гонка данных!
        }()
    }
    
    wg.Wait()
    
    // Результат будет разным каждый раз
    t.Logf("Counter: %d", Get())
}

func TestNoRace(t *testing.T) {
    globalCounter = 0
    
    var wg sync.WaitGroup
    var mutex sync.Mutex
    
    incrementSafe := func() {
        mutex.Lock()
        globalCounter++
        mutex.Unlock()
    }
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            incrementSafe()
        }()
    }
    
    wg.Wait()
    
    if Get() != 1000 {
        t.Errorf("Expected 1000, got %d", Get())
    }
}
```

Запуск с детектором гонок:
```bash
go test -race -v
```

## Тестирование с контекстом

### Тестирование отмены операций

```go
// service/long_running.go
package service

import (
    "context"
    "time"
)

func LongRunningOperation(ctx context.Context, duration time.Duration) error {
    select {
    case <-time.After(duration):
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

```go
// service/long_running_test.go
package service

import (
    "context"
    "testing"
    "time"
)

func TestLongRunningOperation_Success(t *testing.T) {
    ctx := context.Background()
    
    err := LongRunningOperation(ctx, 10*time.Millisecond)
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}

func TestLongRunningOperation_Cancel(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
    defer cancel()
    
    err := LongRunningOperation(ctx, 100*time.Millisecond)
    if err == nil {
        t.Fatal("Expected error, got nil")
    }
    
    if err != context.DeadlineExceeded {
        t.Errorf("Expected DeadlineExceeded, got %v", err)
    }
}
```

## Покрытие кода тестами

### Измерение покрытия

```bash
# Запустить тесты с измерением покрытия
go test -cover

# Детальное покрытие
go test -coverprofile=coverage.out

# Просмотр покрытия в браузере
go tool cover -html=coverage.out

# Покрытие в процентах
go tool cover -func=coverage.out
```

### Пример вывода покрытия

```
coverage: 85.7% of statements
```

### Улучшение покрытия

```go
// math.go
package math

import "errors"

func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("деление на ноль")
    }
    // Добавим еще один случай
    if a == 0 {
        return 0, nil
    }
    return a / b, nil
}
```

```go
// math_test.go - добавляем тест для нового случая
func TestDivide_ZeroDividend(t *testing.T) {
    result, err := Divide(0, 5)
    if err != nil {
        t.Fatalf("Неожиданная ошибка: %v", err)
    }
    
    if result != 0 {
        t.Errorf("Divide(0, 5) = %f; expected 0", result)
    }
}
```

## Расширенные техники тестирования

### 1. Тестирование с таблицами подтестов

```go
func TestComplexFunction(t *testing.T) {
    tests := map[string]struct {
        input    string
        expected string
        wantErr  bool
    }{
        "valid input": {
            input:    "hello world",
            expected: "HELLO WORLD",
            wantErr:  false,
        },
        "empty input": {
            input:    "",
            expected: "",
            wantErr:  false,
        },
        "invalid input": {
            input:    string([]byte{0xff, 0xfe, 0xfd}),
            expected: "",
            wantErr:  true,
        },
    }
    
    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            result, err := ComplexFunction(tc.input)
            
            if tc.wantErr {
                if err == nil {
                    t.Fatal("expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            
            if result != tc.expected {
                t.Errorf("expected %q, got %q", tc.expected, result)
            }
        })
    }
}
```

### 2. Тестирование с временными файлами

```go
func TestFileProcessing(t *testing.T) {
    // Создаем временный файл
    tmpFile, err := ioutil.TempFile("", "test_*.txt")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpFile.Name()) // Удаляем файл после теста
    defer tmpFile.Close()
    
    // Записываем тестовые данные
    content := "hello world"
    if _, err := tmpFile.Write([]byte(content)); err != nil {
        t.Fatal(err)
    }
    
    // Тестируем функцию
    result, err := ProcessFile(tmpFile.Name())
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if result != strings.ToUpper(content) {
        t.Errorf("expected %q, got %q", strings.ToUpper(content), result)
    }
}
```

### 3. Тестирование с сетевыми моками

```go
// network/client.go
package network

import "net/http"

type HTTPClient interface {
    Get(url string) (*http.Response, error)
}

type RealHTTPClient struct{}

func (c *RealHTTPClient) Get(url string) (*http.Response, error) {
    return http.Get(url)
}

// network/client_test.go
type MockHTTPClient struct {
    Response *http.Response
    Error    error
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
    return m.Response, m.Error
}

func TestFetchData(t *testing.T) {
    // Создаем мок ответа
    mockResponse := &http.Response{
        StatusCode: 200,
        Body:       ioutil.NopCloser(strings.NewReader(`{"status": "ok"}`)),
    }
    
    mockClient := &MockHTTPClient{
        Response: mockResponse,
        Error:    nil,
    }
    
    data, err := FetchData(mockClient, "http://example.com")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if data.Status != "ok" {
        t.Errorf("expected status 'ok', got %q", data.Status)
    }
}
```

## Лучшие практики

### 1. Именование тестов

```go
// ХОРОШО - описательные имена
func TestUserService_CreateUser_Success(t *testing.T) { ... }
func TestUserService_CreateUser_InvalidName(t *testing.T) { ... }
func TestUserService_CreateUser_DatabaseError(t *testing.T) { ... }

// ПЛОХО - неинформативные имена
func TestCreateUser1(t *testing.T) { ... }
func TestCreateUser2(t *testing.T) { ... }
```

### 2. Изоляция тестов

```go
// ХОРОШО - каждый тест независим
func TestAdd_PositiveNumbers(t *testing.T) {
    // Начинаем с чистого состояния
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("...")
    }
}

// ПЛОХО - тесты зависят друг от друга
var globalState int

func TestIncrement1(t *testing.T) {
    globalState++
}

func TestIncrement2(t *testing.T) {
    if globalState != 1 { // Зависит от предыдущего теста
        t.Errorf("...")
    }
}
```

### 3. Использование testify для удобства

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

## Распространенные ошибки

### 1. Игнорирование ошибок

```go
// ПЛОХО
func TestBad(t *testing.T) {
    result, _ := Divide(10, 2) // Игнорируем ошибку
    if result != 5 {
        t.Errorf("...")
    }
}

// ХОРОШО
func TestGood(t *testing.T) {
    result, err := Divide(10, 2)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != 5 {
        t.Errorf("...")
    }
}
```

### 2. Недостаточное покрытие граничных случаев

```go
// ПЛОХО - только основной случай
func TestDivide(t *testing.T) {
    result, _ := Divide(10, 2)
    if result != 5 {
        t.Errorf("...")
    }
}

// ХОРОШО - все граничные случаи
func TestDivide_Comprehensive(t *testing.T) {
    tests := []struct {
        name string
        a, b float64
        want float64
        err  bool
    }{
        {"normal", 10, 2, 5, false},
        {"zero dividend", 0, 5, 0, false},
        {"zero divisor", 5, 0, 0, true},
        {"negative", -10, 2, -5, false},
        {"both negative", -10, -2, 5, false},
        {"fractional", 1, 3, 1.0/3.0, false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Divide(tt.a, tt.b)
            
            if tt.err {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want, result)
        })
    }
}
```

### 3. Гонки данных в тестах

```go
// ПЛОХО - гонка данных
func TestCounter(t *testing.T) {
    var counter int
    var wg sync.WaitGroup
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // Гонка данных!
        }()
    }
    
    wg.Wait()
    if counter != 100 {
        t.Errorf("Expected 100, got %d", counter)
    }
}

// ХОРОШО - с синхронизацией
func TestCounterSafe(t *testing.T) {
    var counter int
    var mutex sync.Mutex
    var wg sync.WaitGroup
    
    increment := func() {
        mutex.Lock()
        counter++
        mutex.Unlock()
    }
    
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            increment()
        }()
    }
    
    wg.Wait()
    if counter != 100 {
        t.Errorf("Expected 100, got %d", counter)
    }
}
```

## Тестирование архитектуры приложений

### Тестирование трехслойной архитектуры

```go
// handlers/user_handler.go
type UserHandler struct {
    userUseCase usecases.UserUseCase
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Query().Get("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Неверный ID", http.StatusBadRequest)
        return
    }
    
    user, err := h.userUseCase.GetUserByID(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// handlers/user_handler_test.go
type mockUserUseCase struct {
    GetUserByIDFunc func(ctx context.Context, id int) (*models.User, error)
}

func (m *mockUserUseCase) GetUserByID(ctx context.Context, id int) (*models.User, error) {
    return m.GetUserByIDFunc(ctx, id)
}

func TestUserHandler_GetUserByID(t *testing.T) {
    tests := []struct {
        name           string
        userID         string
        useCaseResult  *models.User
        useCaseError   error
        expectedStatus int
    }{
        {
            name:           "success",
            userID:         "1",
            useCaseResult:  &models.User{ID: 1, Name: "Иван"},
            useCaseError:   nil,
            expectedStatus: http.StatusOK,
        },
        {
            name:           "invalid id",
            userID:         "abc",
            expectedStatus: http.StatusBadRequest,
        },
        {
            name:           "use case error",
            userID:         "1",
            useCaseError:   errors.New("database error"),
            expectedStatus: http.StatusInternalServerError,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Создаем мок use case
            mockUseCase := &mockUserUseCase{
                GetUserByIDFunc: func(ctx context.Context, id int) (*models.User, error) {
                    return tt.useCaseResult, tt.useCaseError
                },
            }
            
            // Создаем handler
            handler := &UserHandler{userUseCase: mockUseCase}
            
            // Создаем запрос
            url := "/user"
            if tt.userID != "" {
                url += "?id=" + tt.userID
            }
            req := httptest.NewRequest("GET", url, nil)
            w := httptest.NewRecorder()
            
            // Вызываем handler
            handler.GetUserByID(w, req)
            
            // Проверяем результат
            resp := w.Result()
            if resp.StatusCode != tt.expectedStatus {
                t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
            }
        })
    }
}
```

## См. также

- [Тестирование для чайников](../concepts/testing.md) - базовое объяснение
- [HTTP серверы](../concepts/http-server.md) - как тестировать веб-приложения
- [Базы данных](../concepts/database.md) - как тестировать работу с БД
- [Профилирование](../concepts/profiling.md) - как использовать бенчмарки
- [Практические примеры](../examples/testing) - примеры кода