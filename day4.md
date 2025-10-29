# День 4: Архитектура и проектирование (12 часов)

## Чистая архитектура за 3 слоя

### 1. Delivery слой (HTTP/gRPC) — хендлеры

Delivery слой отвечает за взаимодействие с внешним миром: HTTP запросы, gRPC вызовы, CLI интерфейсы и т.д.

#### HTTP хендлеры
```go
// handlers/user_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/yourproject/internal/usecases"
)

type UserHandler struct {
    userUseCase usecases.UserUseCase
}

func NewUserHandler(uc usecases.UserUseCase) *UserHandler {
    return &UserHandler{
        userUseCase: uc,
    }
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    // Получаем ID из URL
    idStr := r.URL.Query().Get("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Неверный ID", http.StatusBadRequest)
        return
    }
    
    // Вызываем use case
    user, err := h.userUseCase.GetUserByID(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Отправляем ответ
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

#### Роутинг
```go
// main.go
package main

import (
    "net/http"
    
    "github.com/yourproject/internal/handlers"
    "github.com/yourproject/internal/repositories"
    "github.com/yourproject/internal/usecases"
)

func main() {
    // Создаем репозиторий
    userRepo := repositories.NewUserRepository()
    
    // Создаем use case
    userUseCase := usecases.NewUserUseCase(userRepo)
    
    // Создаем хендлер
    userHandler := handlers.NewUserHandler(userUseCase)
    
    // Настраиваем маршруты
    http.HandleFunc("/users", userHandler.GetUserByID)
    
    // Запускаем сервер
    http.ListenAndServe(":8080", nil)
}
```

### 2. Business logic слой — use cases, сервисы

Business logic слой содержит бизнес-логику приложения, реализованную в виде use case'ов и сервисов.

#### Use Case
```go
// usecases/user_usecase.go
package usecases

import (
    "context"
    
    "github.com/yourproject/internal/models"
    "github.com/yourproject/internal/repositories"
)

type UserUseCase interface {
    GetUserByID(ctx context.Context, id int) (*models.User, error)
    CreateUser(ctx context.Context, user *models.User) error
}

type userUseCase struct {
    userRepo repositories.UserRepository
}

func NewUserUseCase(repo repositories.UserRepository) UserUseCase {
    return &userUseCase{
        userRepo: repo,
    }
}

func (uc *userUseCase) GetUserByID(ctx context.Context, id int) (*models.User, error) {
    // Бизнес-логика: проверка прав доступа, валидация и т.д.
    if id <= 0 {
        return nil, errors.New("Неверный ID пользователя")
    }
    
    // Вызываем репозиторий
    user, err := uc.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}

func (uc *userUseCase) CreateUser(ctx context.Context, user *models.User) error {
    // Бизнес-логика: валидация, проверка уникальности и т.д.
    if user.Name == "" {
        return errors.New("Имя пользователя не может быть пустым")
    }
    
    // Вызываем репозиторий
    err := uc.userRepo.Create(ctx, user)
    if err != nil {
        return err
    }
    
    return nil
}
```

#### Модели
```go
// models/user.go
package models

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

### 3. Repository слой — работа с данными

Repository слой отвечает за взаимодействие с базой данных или другими внешними системами хранения данных.

#### Интерфейс репозитория
```go
// repositories/user_repository.go
package repositories

import (
    "context"
    
    "github.com/yourproject/internal/models"
)

type UserRepository interface {
    GetByID(ctx context.Context, id int) (*models.User, error)
    Create(ctx context.Context, user *models.User) error
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id int) error
}
```

#### Реализация репозитория
```go
// repositories/postgres_user_repository.go
package repositories

import (
    "context"
    "database/sql"
    
    "github.com/yourproject/internal/models"
    _ "github.com/lib/pq"
)

type postgresUserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
    return &postgresUserRepository{
        db: db,
    }
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    row := r.db.QueryRowContext(ctx, query, id)
    
    var user models.User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, errors.New("Пользователь не найден")
        }
        return nil, err
    }
    
    return &user, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, user *models.User) error {
    query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
    
    err := r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
    if err != nil {
        return err
    }
    
    return nil
}
```

## Dependency Injection

Dependency Injection (DI) - это паттерн, который позволяет передавать зависимости в компоненты вместо того, чтобы создавать их внутри.

### Конструктор внедрения зависимостей

#### Простой DI контейнер
```go
// di/container.go
package di

import (
    "database/sql"
    
    "github.com/yourproject/internal/handlers"
    "github.com/yourproject/internal/repositories"
    "github.com/yourproject/internal/usecases"
)

type Container struct {
    db *sql.DB
}

func NewContainer(db *sql.DB) *Container {
    return &Container{
        db: db,
    }
}

func (c *Container) GetUserHandler() *handlers.UserHandler {
    userRepo := repositories.NewUserRepository(c.db)
    userUseCase := usecases.NewUserUseCase(userRepo)
    return handlers.NewUserHandler(userUseCase)
}
```

#### Использование DI контейнера
```go
// main.go
package main

import (
    "database/sql"
    "net/http"
    
    "github.com/yourproject/internal/di"
    _ "github.com/lib/pq"
)

func main() {
    // Подключение к базе данных
    db, err := sql.Open("postgres", "user=user dbname=test sslmode=disable")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // Создаем DI контейнер
    container := di.NewContainer(db)
    
    // Получаем хендлеры из контейнера
    userHandler := container.GetUserHandler()
    
    // Настраиваем маршруты
    http.HandleFunc("/users", userHandler.GetUserByID)
    
    // Запускаем сервер
    http.ListenAndServe(":8080", nil)
}
```

### Интерфейсы для моков в тестах

Интерфейсы позволяют создавать моки для тестирования, заменяя реальные реализации на фиктивные.

#### Мок репозитория
```go
// mocks/user_repository_mock.go
package mocks

import (
    "context"
    
    "github.com/yourproject/internal/models"
    "github.com/yourproject/internal/repositories"
)

type MockUserRepository struct {
    GetByIDFunc func(ctx context.Context, id int) (*models.User, error)
    CreateFunc  func(ctx context.Context, user *models.User) error
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, user)
    }
    return nil
}

// ... остальные методы
```

#### Тестирование с моками
```go
// usecases/user_usecase_test.go
package usecases

import (
    "context"
    "testing"
    
    "github.com/yourproject/internal/mocks"
    "github.com/yourproject/internal/models"
)

func TestGetUserByID(t *testing.T) {
    // Создаем мок репозитория
    mockRepo := &mocks.MockUserRepository{
        GetByIDFunc: func(ctx context.Context, id int) (*models.User, error) {
            if id == 1 {
                return &models.User{ID: 1, Name: "Иван", Email: "ivan@example.com"}, nil
            }
            return nil, errors.New("Пользователь не найден")
        },
    }
    
    // Создаем use case с моком
    userUseCase := NewUserUseCase(mockRepo)
    
    // Тестируем
    user, err := userUseCase.GetUserByID(context.Background(), 1)
    if err != nil {
        t.Fatalf("Неожиданная ошибка: %v", err)
    }
    
    if user.Name != "Иван" {
        t.Errorf("Ожидалось имя 'Иван', получено '%s'", user.Name)
    }
}
```

## Практические задания

1. Создайте простое веб-приложение с тремя слоями архитектуры (delivery, use case, repository).
2. Реализуйте CRUD операции для сущности (например, пользователь).
3. Настройте dependency injection для управления зависимостями.
4. Напишите моки для репозиториев и протестируйте use case'ы.
5. Добавьте валидацию и обработку ошибок на всех уровнях.