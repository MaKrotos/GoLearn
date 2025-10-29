# Архитектура приложений в Go - объяснение для чайников

## Что такое архитектура?

Представьте архитектуру как **план дома**:
- **Фундамент** - основа всего здания
- **Стены** - разделяют пространство на комнаты
- **Комнаты** - каждая для своей цели (кухня, спальня, ванная)
- **Коридоры** - соединяют комнаты

Архитектура приложения - это **структура кода**, которая:
- **Разделяет** ответственность
- **Упрощает** понимание
- **Облегчает** тестирование
- **Упрощает** поддержку

## Чистая архитектура (Clean Architecture)

### Три слоя архитектуры

```
┌─────────────────────────────────────┐
│           Delivery Layer            │  ← Внешний мир (HTTP, gRPC)
├─────────────────────────────────────┤
│          Business Logic             │  ← Правила бизнеса
├─────────────────────────────────────┤
│           Repository Layer          │  ← Работа с данными
└─────────────────────────────────────┘
```

### 1. Delivery Layer (Слой доставки)

**Назначение**: Общение с внешним миром
- HTTP запросы
- gRPC вызовы
- CLI интерфейсы
- WebSockets

**Пример**:
```go
// handlers/user_handler.go
type UserHandler struct {
    userUseCase usecases.UserUseCase
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    // 1. Получаем данные из запроса
    idStr := r.URL.Query().Get("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Неверный ID", http.StatusBadRequest)
        return
    }
    
    // 2. Вызываем бизнес-логику
    user, err := h.userUseCase.GetUserByID(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 3. Отправляем ответ
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

### 2. Business Logic Layer (Слой бизнес-логики)

**Назначение**: Правила и логика приложения
- Валидация данных
- Бизнес-правила
- Координация между слоями

**Пример**:
```go
// usecases/user_usecase.go
type userUseCase struct {
    userRepo repositories.UserRepository
}

func (uc *userUseCase) GetUserByID(ctx context.Context, id int) (*models.User, error) {
    // 1. Бизнес-валидация
    if id <= 0 {
        return nil, errors.New("Неверный ID пользователя")
    }
    
    // 2. Получаем данные из репозитория
    user, err := uc.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
    }
    
    // 3. Бизнес-логика (например, проверка прав доступа)
    // ...
    
    return user, nil
}
```

### 3. Repository Layer (Слой репозиториев)

**Назначение**: Работа с данными
- Базы данных
- Внешние API
- Файловая система

**Пример**:
```go
// repositories/postgres_user_repository.go
type postgresUserRepository struct {
    db *sql.DB
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
```

## Dependency Injection (Внедрение зависимостей)

### Что такое DI?

Dependency Injection - это **способ передачи зависимостей** в компоненты вместо их создания внутри.

### Пример без DI (плохо):

```go
// ПЛОХО - жесткая зависимость
type UserService struct{}

func (s *UserService) CreateUser(user *User) error {
    // Жестко завязаны на конкретные реализации
    db := connectToDatabase()        // Прямое создание
    emailer := NewSMTPMailer()       // Прямое создание
    validator := NewUserValidator()  // Прямое создание
    
    if !validator.Validate(user) {
        return errors.New("Неверные данные пользователя")
    }
    
    return db.Save(user)
}
```

### Пример с DI (хорошо):

```go
// ХОРОШО - зависимости передаются извне
type UserService struct {
    db        Database
    emailer   Emailer
    validator Validator
}

func NewUserService(db Database, emailer Emailer, validator Validator) *UserService {
    return &UserService{
        db:        db,
        emailer:   emailer,
        validator: validator,
    }
}

func (s *UserService) CreateUser(user *User) error {
    if !s.validator.Validate(user) {
        return errors.New("Неверные данные пользователя")
    }
    
    return s.db.Save(user)
}
```

### DI контейнер

```go
// di/container.go
type Container struct {
    db *sql.DB
}

func NewContainer(db *sql.DB) *Container {
    return &Container{db: db}
}

func (c *Container) GetUserRepository() repositories.UserRepository {
    return repositories.NewUserRepository(c.db)
}

func (c *Container) GetUserUseCase() usecases.UserUseCase {
    return usecases.NewUserUseCase(c.GetUserRepository())
}

func (c *Container) GetUserHandler() *handlers.UserHandler {
    return handlers.NewUserHandler(c.GetUserUseCase())
}
```

## Интерфейсы для тестирования

### Зачем нужны интерфейсы?

Интерфейсы позволяют **заменять реальные реализации** на тестовые (моки).

### Пример:

```go
// repositories/user_repository.go
type UserRepository interface {
    GetByID(ctx context.Context, id int) (*models.User, error)
    Create(ctx context.Context, user *models.User) error
}

// Реальная реализация
type postgresUserRepository struct {
    db *sql.DB
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
    // Реальный запрос к базе данных
}

// Мок для тестирования
type mockUserRepository struct {
    users map[int]*models.User
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
    user, exists := m.users[id]
    if !exists {
        return nil, errors.New("Пользователь не найден")
    }
    return user, nil
}
```

### Использование в тестах:

```go
func TestUserUseCase_GetUserByID(t *testing.T) {
    // Создаем мок
    mockRepo := &mockUserRepository{
        users: map[int]*models.User{
            1: {ID: 1, Name: "Иван", Email: "ivan@example.com"},
        },
    }
    
    // Создаем use case с моком
    useCase := NewUserUseCase(mockRepo)
    
    // Тестируем
    user, err := useCase.GetUserByID(context.Background(), 1)
    if err != nil {
        t.Fatalf("Неожиданная ошибка: %v", err)
    }
    
    if user.Name != "Иван" {
        t.Errorf("Ожидалось имя 'Иван', получено '%s'", user.Name)
    }
}
```

## Практические примеры структуры проекта

### Маленький проект:

```
project/
├── main.go                 # Точка входа
├── handlers/               # HTTP обработчики
│   └── user_handler.go
├── usecases/               # Бизнес-логика
│   └── user_usecase.go
├── repositories/           # Репозитории
│   └── user_repository.go
├── models/                 # Модели данных
│   └── user.go
└── di/                     # DI контейнер
    └── container.go
```

### Средний проект:

```
project/
├── cmd/                    # Точки входа
│   └── api/
│       └── main.go
├── internal/               # Внутренний код
│   ├── handlers/           # HTTP обработчики
│   ├── usecases/           # Бизнес-логика
│   ├── repositories/       # Репозитории
│   ├── models/             # Модели данных
│   ├── di/                 # DI контейнер
│   └── config/             # Конфигурация
├── pkg/                    # Переиспользуемый код
│   └── utils/
├── migrations/             # Миграции базы данных
└── configs/                # Файлы конфигурации
```

## Лучшие практики

### 1. Разделяйте ответственность

```go
// ПЛОХО - всё в одном файле
func HandleUserRequest(w http.ResponseWriter, r *http.Request) {
    // Валидация
    // Бизнес-логика
    // Работа с БД
    // Отправка email
    // Логирование
}

// ХОРОШО - каждый слой отвечает за свою часть
func (h *UserHandler) HandleUserRequest(w http.ResponseWriter, r *http.Request) {
    // Только обработка HTTP
    userID := h.extractUserID(r)
    user, err := h.userUseCase.GetUser(userID)
    h.sendResponse(w, user, err)
}
```

### 2. Используйте интерфейсы на уровне бизнес-логики

```go
// usecases/user_usecase.go
type UserUseCase struct {
    userRepo    UserRepository    // Интерфейс, не конкретная реализация
    emailSender EmailSender       // Интерфейс
    logger      Logger            // Интерфейс
}
```

### 3. Минимизируйте зависимости

```go
// ПЛОХО - много зависимостей
type UserService struct {
    db        *sql.DB
    emailer   *SMTPMailer
    logger    *Logger
    cache     *RedisClient
    metrics   *MetricsCollector
    // ... еще 10 зависимостей
}

// ЛУЧШЕ - только необходимые зависимости
type UserService struct {
    userRepo    UserRepository
    emailSender EmailSender
}
```

### 4. Используйте контекст для отмены и таймаутов

```go
func (uc *UserUseCase) GetUserByID(ctx context.Context, id int) (*User, error) {
    // Контекст передается во все вызовы
    return uc.userRepo.GetByID(ctx, id)
}
```

## Распространенные ошибки

### 1. Смешивание слоев

```go
// ПЛОХО - бизнес-логика в обработчике
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Валидация
    if user.Name == "" {
        http.Error(w, "Имя обязательно", http.StatusBadRequest)
        return
    }
    
    // Бизнес-логика
    if len(user.Name) < 2 {
        http.Error(w, "Имя слишком короткое", http.StatusBadRequest)
        return
    }
    
    // Работа с БД
    _, err := h.db.Exec("INSERT INTO users (name) VALUES (?)", user.Name)
    // ...
}
```

### 2. Жесткие зависимости

```go
// ПЛОХО - жесткая зависимость от реализации
type OrderService struct {
    db *sql.DB  // Конкретная реализация
}

// ЛУЧШЕ - зависимость от интерфейса
type OrderService struct {
    orderRepo OrderRepository  // Интерфейс
}
```

## См. также

- [Интерфейсы](interface.md) - основа для архитектуры
- [Тестирование](testing.md) - как тестировать архитектуру
- [Dependency Injection](../theory/di.md) - подробнее о внедрении зависимостей
- [HTTP серверы](http-server.md) - delivery layer
- [Базы данных](database.md) - repository layer