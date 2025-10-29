# Продвинутые лучшие практики и архитектурные подходы в Go

## 1. Чистая архитектура (Clean Architecture) - углубленное изучение

### Расширенная структура проекта

#### Масштабируемая архитектура для больших проектов:
```
project/
├── cmd/                    # Точки входа
│   ├── api/               # HTTP API сервер
│   │   └── main.go
│   ├── worker/            # Background worker
│   │   └── main.go
│   └── cli/               # CLI инструменты
│       └── main.go
├── internal/              # Внутренний код (не экспортируется)
│   ├── app/               # Бизнес-логика
│   │   ├── user/          # Модуль пользователей
│   │   │   ├── service.go # Сервисный слой
│   │   │   ├── model.go   # Модели данных
│   │   │   └── errors.go  # Ошибки модуля
│   │   └── order/         # Модуль заказов
│   │       ├── service.go
│   │       ├── model.go
│   │       └── errors.go
│   ├── adapters/          # Адаптеры для внешних систем
│   │   ├── postgres/      # Адаптер для PostgreSQL
│   │   ├── redis/         # Адаптер для Redis
│   │   └── http/          # HTTP клиенты для внешних API
│   ├── ports/             # Интерфейсы для взаимодействия
│   │   ├── primary/       # Входящие порты (handlers)
│   │   └── secondary/     # Исходящие порты (repositories)
│   └── shared/            # Общие компоненты
│       ├── kernel/        # Ядро приложения
│       ├── config/        # Конфигурация
│       └── logger/        # Логирование
├── pkg/                   # Переиспользуемые библиотеки
│   ├── validator/         # Валидация данных
│   ├── middleware/        # HTTP middleware
│   └── utils/             # Утилиты общего назначения
├── deployments/           # Файлы для деплоя
│   ├── docker/            # Docker файлы
│   ├── kubernetes/        # Kubernetes манифесты
│   └── scripts/           # Скрипты деплоя
├── docs/                  # Документация
├── migrations/            # Миграции базы данных
├── configs/               # Файлы конфигурации
└── test/                  # Тестовые данные и интеграционные тесты
```

### Продвинутые техники разделения слоев

#### Domain Driven Design (DDD) в Go:
```go
// domain/user/entity.go
package user

// User представляет сущность пользователя
type User struct {
    id    UserID
    name  string
    email Email
    role  Role
}

// UserID идентификатор пользователя
type UserID string

// Email электронная почта пользователя
type Email string

// Role роль пользователя
type Role string

const (
    RoleUser  Role = "user"
    RoleAdmin Role = "admin"
)

// domain/user/repository.go
package user

// Repository интерфейс репозитория пользователей
type Repository interface {
    Save(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id UserID) (*User, error)
    FindByEmail(ctx context.Context, email Email) (*User, error)
    Delete(ctx context.Context, id UserID) error
}

// domain/user/service.go
package user

// Service сервис пользователей
type Service struct {
    repo Repository
}

// NewService создает новый сервис пользователей
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// Register регистрирует нового пользователя
func (s *Service) Register(ctx context.Context, name string, email Email) (*User, error) {
    // Бизнес-логика: проверка уникальности email
    existing, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return nil, err
    }
    
    if existing != nil {
        return nil, errors.New("user with this email already exists")
    }
    
    // Создание нового пользователя
    user := &User{
        id:    generateUserID(),
        name:  name,
        email: email,
        role:  RoleUser,
    }
    
    // Сохранение пользователя
    if err := s.repo.Save(ctx, user); err != nil {
        return nil, err
    }
    
    return user, nil
}

// adapters/postgres/user_repository.go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "yourproject/internal/app/user"
)

// UserRepository реализация репозитория пользователей для PostgreSQL
type UserRepository struct {
    db *sql.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Save сохраняет пользователя в базе данных
func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
    query := `
        INSERT INTO users (id, name, email, role)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (id) DO UPDATE
        SET name = $2, email = $3, role = $4
    `
    
    _, err := r.db.ExecContext(ctx, query, u.ID(), u.Name(), u.Email(), u.Role())
    if err != nil {
        return fmt.Errorf("failed to save user: %w", err)
    }
    
    return nil
}

// FindByID находит пользователя по ID
func (r *UserRepository) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
    query := "SELECT id, name, email, role FROM users WHERE id = $1"
    
    row := r.db.QueryRowContext(ctx, query, id)
    
    var u user.User
    err := row.Scan(&u.id, &u.name, &u.email, &u.role)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, fmt.Errorf("failed to find user by ID: %w", err)
    }
    
    return &u, nil
}
```

### Паттерн CQRS (Command Query Responsibility Segregation)

#### Разделение команд и запросов:
```go
// ports/primary/user_handler.go
package primary

import (
    "context"
    "encoding/json"
    "net/http"
)

// UserHandler обработчик HTTP запросов для пользователей
type UserHandler struct {
    commandService UserCommandService
    queryService   UserQueryService
}

// NewUserHandler создает новый обработчик
func NewUserHandler(commandService UserCommandService, queryService UserQueryService) *UserHandler {
    return &UserHandler{
        commandService: commandService,
        queryService:   queryService,
    }
}

// CreateUser создает нового пользователя
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }
    
    cmd := CreateUserCommand{
        Name:  req.Name,
        Email: req.Email,
    }
    
    user, err := h.commandService.CreateUser(r.Context(), cmd)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(user)
}

// GetUser получает пользователя по ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    userID := getUserIDFromPath(r.URL.Path)
    
    user, err := h.queryService.GetUser(r.Context(), userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    if user == nil {
        http.Error(w, "user not found", http.StatusNotFound)
        return
    }
    
    json.NewEncoder(w).Encode(user)
}

// ports/secondary/user_command_service.go
package secondary

import "context"

// UserCommandService сервис для выполнения команд над пользователями
type UserCommandService interface {
    CreateUser(ctx context.Context, cmd CreateUserCommand) (*User, error)
    UpdateUser(ctx context.Context, cmd UpdateUserCommand) (*User, error)
    DeleteUser(ctx context.Context, userID string) error
}

// CreateUserCommand команда создания пользователя
type CreateUserCommand struct {
    Name  string
    Email string
}

// UpdateUserCommand команда обновления пользователя
type UpdateUserCommand struct {
    ID    string
    Name  string
    Email string
}

// ports/secondary/user_query_service.go
package secondary

import "context"

// UserQueryService сервис для выполнения запросов к пользователям
type UserQueryService interface {
    GetUser(ctx context.Context, userID string) (*User, error)
    ListUsers(ctx context.Context, filter UserFilter) ([]*User, error)
    GetUserByEmail(ctx context.Context, email string) (*User, error)
}

// UserFilter фильтр для списка пользователей
type UserFilter struct {
    Limit  int
    Offset int
    Name   string
}
```

## 2. Тестирование - углубленное изучение

### Расширенные техники тестирования

#### Property-based тестирование:
```go
// go.mod: require github.com/stretchr/testify v1.8.0
// go.mod: require github.com/leanovate/gopter v0.2.9

import (
    "testing"
    "github.com/leanovate/gopter"
    "github.com/leanovate/gopter/gen"
    "github.com/leanovate/gopter/prop"
)

func TestUserService_PropertyBased(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("Email validation", prop.ForAll(
        func(email string) bool {
            // Тестируем валидацию email
            _, err := validateEmail(email)
            // Проверяем, что валидация работает корректно
            return (isValidEmail(email) && err == nil) || 
                   (!isValidEmail(email) && err != nil)
        },
        gen.Email(),
    ))
    
    properties.Property("User creation with valid data", prop.ForAll(
        func(name string, email string) bool {
            // Тестируем создание пользователя с валидными данными
            service := NewUserService(mockRepo)
            user, err := service.CreateUser(context.Background(), name, email)
            
            // Проверяем, что пользователь создается успешно
            return user != nil && err == nil && 
                   user.Name == name && user.Email == email
        },
        gen.AlphaString(), // Генерируем случайные имена
        gen.Email(),       // Генерируем случайные email
    ))
    
    properties.TestingRun(t)
}
```

#### Fuzz тестирование:
```go
// go.mod: require github.com/google/gofuzz v1.2.0

import (
    "testing"
    "github.com/google/gofuzz"
)

func FuzzUserService_CreateUser(f *testing.F) {
    // Добавляем примеры входных данных
    f.Add("John Doe", "john@example.com")
    f.Add("Jane Smith", "jane@test.org")
    
    f.Fuzz(func(t *testing.T, name string, email string) {
        // Проверяем, что сервис корректно обрабатывает различные входные данные
        service := NewUserService(mockRepo)
        user, err := service.CreateUser(context.Background(), name, email)
        
        // Если email валиден, пользователь должен быть создан
        if isValidEmail(email) && name != "" {
            if user == nil || err != nil {
                t.Errorf("Expected user to be created, got error: %v", err)
            }
        } else {
            // Если данные невалидны, должна быть ошибка
            if user != nil && err == nil {
                t.Error("Expected error for invalid input")
            }
        }
    })
}
```

#### Mock генерация:
```go
//go:generate mockgen -source=user_service.go -destination=mocks/user_service_mock.go

// app/user/service.go
package user

import "context"

// Service интерфейс сервиса пользователей
type Service interface {
    CreateUser(ctx context.Context, name, email string) (*User, error)
    GetUser(ctx context.Context, id string) (*User, error)
    UpdateUser(ctx context.Context, id, name, email string) (*User, error)
    DeleteUser(ctx context.Context, id string) error
}

// Mock генерируется автоматически с помощью mockgen
// mocks/user_service_mock.go будет содержать реализацию мока
```

### Тестирование конкурентности

#### Тестирование гонок данных:
```go
func TestConcurrentUserAccess(t *testing.T) {
    // Включаем детектор гонок
    if testing.Short() {
        t.Skip("Skipping race test in short mode")
    }
    
    service := NewUserService(NewInMemoryUserRepository())
    
    // Создаем пользователя
    user, err := service.CreateUser(context.Background(), "Test User", "test@example.com")
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    // Запускаем конкурентные операции
    var wg sync.WaitGroup
    errors := make(chan error, 100)
    
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            // Конкурентное чтение
            _, err := service.GetUser(context.Background(), user.ID)
            if err != nil {
                errors <- err
                return
            }
        }()
    }
    
    for i := 0; i < 50; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            // Конкурентная запись
            _, err := service.UpdateUser(context.Background(), user.ID, "Updated User", "updated@example.com")
            if err != nil {
                errors <- err
                return
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    // Проверяем ошибки
    for err := range errors {
        if err != nil {
            t.Errorf("Concurrent operation failed: %v", err)
        }
    }
}
```

## 3. Профилирование и оптимизация - углубленное изучение

### Расширенные техники профилирования

#### Профилирование аллокаций:
```go
// benchmark_test.go
package main

import (
    "testing"
    "github.com/pkg/profile"
)

func BenchmarkUserCreation(b *testing.B) {
    // Включаем профилирование аллокаций
    defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop()
    
    service := NewUserService(NewInMemoryUserRepository())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.CreateUser(context.Background(), 
            fmt.Sprintf("User %d", i), 
            fmt.Sprintf("user%d@example.com", i))
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkUserCreationWithPool(b *testing.B) {
    // Включаем профилирование аллокаций
    defer profile.Start(profile.MemProfile, profile.MemProfileRate(1)).Stop()
    
    service := NewUserServiceWithPool(NewInMemoryUserRepository())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.CreateUser(context.Background(), 
            fmt.Sprintf("User %d", i), 
            fmt.Sprintf("user%d@example.com", i))
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### CPU профилирование с pprof:
```go
// main.go
package main

import (
    "net/http"
    _ "net/http/pprof"
    "log"
)

func main() {
    // Включаем pprof сервер
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Основное приложение
    startServer()
}

// Анализ профиля:
// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
// (pprof) top10
// (pprof) web
// (pprof) list functionName
```

#### Блок-профилирование:
```go
// main.go
package main

import (
    "runtime"
    "net/http"
    _ "net/http/pprof"
)

func main() {
    // Включаем блок-профилирование
    runtime.SetBlockProfileRate(1)
    
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    startServer()
}

// Анализ блок-профиля:
// go tool pprof http://localhost:6060/debug/pprof/block
// (pprof) top10
// (pprof) web
```

### Оптимизация производительности

#### Использование sync.Pool:
```go
// user_pool.go
package user

import "sync"

// UserPool пул пользователей для переиспользования
type UserPool struct {
    pool sync.Pool
}

// NewUserPool создает новый пул пользователей
func NewUserPool() *UserPool {
    return &UserPool{
        pool: sync.Pool{
            New: func() interface{} {
                return &User{}
            },
        },
    }
}

// Get получает пользователя из пула
func (p *UserPool) Get() *User {
    return p.pool.Get().(*User)
}

// Put возвращает пользователя в пул
func (p *UserPool) Put(u *User) {
    // Сбрасываем состояние пользователя перед возвратом в пул
    u.Reset()
    p.pool.Put(u)
}

// User модель пользователя с методом сброса
type User struct {
    ID    string
    Name  string
    Email string
}

// Reset сбрасывает состояние пользователя
func (u *User) Reset() {
    u.ID = ""
    u.Name = ""
    u.Email = ""
}

// UserServiceWithPool сервис пользователей с использованием пула
type UserServiceWithPool struct {
    repo User

[Response interrupted by API Error]