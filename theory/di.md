# Внедрение зависимостей (Dependency Injection) в Go: Полная теория

## Введение во внедрение зависимостей

### Что такое внедрение зависимостей?

Внедрение зависимостей (Dependency Injection, DI) - это **паттерн проектирования**, который:
- **Разделяет** создание и использование объектов
- **Уменьшает** связность между компонентами
- **Упрощает** тестирование и поддержку кода
- **Повышает** гибкость и расширяемость приложений

### Зачем нужно внедрение зависимостей?

1. **Тестируемость** - легко подменять зависимости моками
2. **Гибкость** - можно менять реализации без изменения кода
3. **Поддерживаемость** - уменьшается связность компонентов
4. **Расширяемость** - легко добавлять новые реализации
5. **Контроль** - централизованное управление зависимостями

## Основы внедрения зависимостей в Go

### Проблема без DI

```go
// ПЛОХО - высокая связность
type UserService struct {
    db *sql.DB // Прямая зависимость от конкретной реализации
}

func NewUserService() *UserService {
    db, _ := sql.Open("postgres", "connection_string") // Жесткая привязка
    return &UserService{db: db}
}

func (s *UserService) GetUser(id int) (*User, error) {
    // Используем конкретную реализацию БД
    row := s.db.QueryRow("SELECT id, name FROM users WHERE id = $1", id)
    // ...
}
```

### Решение с DI

```go
// ХОРОШО - интерфейсы и внедрение зависимостей
type UserStorage interface {
    GetUser(id int) (*User, error)
    CreateUser(user *User) error
}

type UserService struct {
    storage UserStorage // Зависимость от интерфейса
}

func NewUserService(storage UserStorage) *UserService {
    return &UserService{storage: storage}
}

func (s *UserService) GetUser(id int) (*User, error) {
    return s.storage.GetUser(id) // Используем абстракцию
}
```

## Практическая реализация DI

### 1. Интерфейсы и реализации

```go
// di/interfaces.go
package di

import (
    "context"
    "database/sql"
)

// User модель пользователя
type User struct {
    ID   int
    Name string
    Email string
}

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
    GetUser(ctx context.Context, id int) (*User, error)
    CreateUser(ctx context.Context, user *User) error
    UpdateUser(ctx context.Context, user *User) error
    DeleteUser(ctx context.Context, id int) error
}

// PostgreSQLUserRepository реализация для PostgreSQL
type PostgreSQLUserRepository struct {
    db *sql.DB
}

func NewPostgreSQLUserRepository(db *sql.DB) *PostgreSQLUserRepository {
    return &PostgreSQLUserRepository{db: db}
}

func (r *PostgreSQLUserRepository) GetUser(ctx context.Context, id int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    row := r.db.QueryRowContext(ctx, query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    
    return &user, nil
}

func (r *PostgreSQLUserRepository) CreateUser(ctx context.Context, user *User) error {
    query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
}

func (r *PostgreSQLUserRepository) UpdateUser(ctx context.Context, user *User) error {
    query := "UPDATE users SET name = $1, email = $2 WHERE id = $3"
    _, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.ID)
    return err
}

func (r *PostgreSQLUserRepository) DeleteUser(ctx context.Context, id int) error {
    query := "DELETE FROM users WHERE id = $1"
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}

// InMemoryUserRepository реализация для тестов
type InMemoryUserRepository struct {
    users map[int]*User
    nextID int
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
    return &InMemoryUserRepository{
        users: make(map[int]*User),
        nextID: 1,
    }
}

func (r *InMemoryUserRepository) GetUser(ctx context.Context, id int) (*User, error) {
    user, exists := r.users[id]
    if !exists {
        return nil, nil
    }
    return user, nil
}

func (r *InMemoryUserRepository) CreateUser(ctx context.Context, user *User) error {
    user.ID = r.nextID
    r.users[user.ID] = user
    r.nextID++
    return nil
}

func (r *InMemoryUserRepository) UpdateUser(ctx context.Context, user *User) error {
    if _, exists := r.users[user.ID]; !exists {
        return sql.ErrNoRows
    }
    r.users[user.ID] = user
    return nil
}

func (r *InMemoryUserRepository) DeleteUser(ctx context.Context, id int) error {
    if _, exists := r.users[id]; !exists {
        return sql.ErrNoRows
    }
    delete(r.users, id)
    return nil
}
```

### 2. Сервисы с внедрением зависимостей

```go
// di/services.go
package di

import (
    "context"
    "errors"
    "regexp"
)

// UserService сервис для работы с пользователями
type UserService struct {
    repo UserRepository
    emailValidator *regexp.Regexp
}

func NewUserService(repo UserRepository) *UserService {
    return &UserService{
        repo: repo,
        emailValidator: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
    }
}

func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
    if id <= 0 {
        return nil, errors.New("invalid user id")
    }
    
    return s.repo.GetUser(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*User, error) {
    if name == "" {
        return nil, errors.New("name is required")
    }
    
    if !s.emailValidator.MatchString(email) {
        return nil, errors.New("invalid email format")
    }
    
    user := &User{
        Name:  name,
        Email: email,
    }
    
    err := s.repo.CreateUser(ctx, user)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id int, name, email string) (*User, error) {
    if id <= 0 {
        return nil, errors.New("invalid user id")
    }
    
    if name == "" {
        return nil, errors.New("name is required")
    }
    
    if !s.emailValidator.MatchString(email) {
        return nil, errors.New("invalid email format")
    }
    
    user := &User{
        ID:    id,
        Name:  name,
        Email: email,
    }
    
    err := s.repo.UpdateUser(ctx, user)
    if err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
    if id <= 0 {
        return errors.New("invalid user id")
    }
    
    return s.repo.DeleteUser(ctx, id)
}

// NotificationService сервис уведомлений
type NotificationService interface {
    SendWelcomeEmail(ctx context.Context, user *User) error
    SendPasswordResetEmail(ctx context.Context, user *User) error
}

// EmailNotificationService реализация уведомлений по email
type EmailNotificationService struct {
    smtpHost string
    smtpPort int
}

func NewEmailNotificationService(smtpHost string, smtpPort int) *EmailNotificationService {
    return &EmailNotificationService{
        smtpHost: smtpHost,
        smtpPort: smtpPort,
    }
}

func (s *EmailNotificationService) SendWelcomeEmail(ctx context.Context, user *User) error {
    // Здесь была бы реальная отправка email
    // log.Printf("Sending welcome email to %s at %s:%d", user.Email, s.smtpHost, s.smtpPort)
    return nil
}

func (s *EmailNotificationService) SendPasswordResetEmail(ctx context.Context, user *User) error {
    // Здесь была бы реальная отправка email
    // log.Printf("Sending password reset email to %s at %s:%d", user.Email, s.smtpHost, s.smtpPort)
    return nil
}

// RegistrationService сервис регистрации пользователей
type RegistrationService struct {
    userService       *UserService
    notificationService NotificationService
}

func NewRegistrationService(userService *UserService, notificationService NotificationService) *RegistrationService {
    return &RegistrationService{
        userService:       userService,
        notificationService: notificationService,
    }
}

func (s *RegistrationService) RegisterUser(ctx context.Context, name, email string) (*User, error) {
    // Создаем пользователя
    user, err := s.userService.CreateUser(ctx, name, email)
    if err != nil {
        return nil, err
    }
    
    // Отправляем приветственное письмо
    err = s.notificationService.SendWelcomeEmail(ctx, user)
    if err != nil {
        // Здесь можно реализовать логику отката или повторной попытки
        return nil, err
    }
    
    return user, nil
}
```

### 3. Контейнер зависимостей

```go
// di/container.go
package di

import (
    "database/sql"
    "sync"
)

// Container контейнер зависимостей
type Container struct {
    mu sync.RWMutex
    
    // Зависимости
    db                  *sql.DB
    userRepository      UserRepository
    userService         *UserService
    notificationService NotificationService
    registrationService *RegistrationService
}

// NewContainer создает новый контейнер
func NewContainer(db *sql.DB) *Container {
    return &Container{
        db: db,
    }
}

// GetUserRepository получает репозиторий пользователей
func (c *Container) GetUserRepository() UserRepository {
    c.mu.RLock()
    if c.userRepository != nil {
        defer c.mu.RUnlock()
        return c.userRepository
    }
    c.mu.RUnlock()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.userRepository == nil {
        c.userRepository = NewPostgreSQLUserRepository(c.db)
    }
    
    return c.userRepository
}

// GetUserService получает сервис пользователей
func (c *Container) GetUserService() *UserService {
    c.mu.RLock()
    if c.userService != nil {
        defer c.mu.RUnlock()
        return c.userService
    }
    c.mu.RUnlock()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.userService == nil {
        c.userService = NewUserService(c.GetUserRepository())
    }
    
    return c.userService
}

// GetNotificationService получает сервис уведомлений
func (c *Container) GetNotificationService() NotificationService {
    c.mu.RLock()
    if c.notificationService != nil {
        defer c.mu.RUnlock()
        return c.notificationService
    }
    c.mu.RUnlock()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.notificationService == nil {
        c.notificationService = NewEmailNotificationService("smtp.example.com", 587)
    }
    
    return c.notificationService
}

// GetRegistrationService получает сервис регистрации
func (c *Container) GetRegistrationService() *RegistrationService {
    c.mu.RLock()
    if c.registrationService != nil {
        defer c.mu.RUnlock()
        return c.registrationService
    }
    c.mu.RUnlock()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.registrationService == nil {
        c.registrationService = NewRegistrationService(
            c.GetUserService(),
            c.GetNotificationService(),
        )
    }
    
    return c.registrationService
}

// Close закрывает контейнер и освобождает ресурсы
func (c *Container) Close() error {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Сбрасываем все зависимости
    c.userRepository = nil
    c.userService = nil
    c.notificationService = nil
    c.registrationService = nil
    
    // Закрываем соединение с БД
    if c.db != nil {
        return c.db.Close()
    }
    
    return nil
}
```

## Расширенные техники DI

### 1. Фабрики зависимостей

```go
// di/factories.go
package di

import (
    "database/sql"
    "fmt"
)

// Factory фабрика для создания зависимостей
type Factory struct {
    config *Config
}

// Config конфигурация приложения
type Config struct {
    Database struct {
        Host     string
        Port     int
        User     string
        Password string
        Name     string
    }
    
    SMTP struct {
        Host string
        Port int
    }
}

// NewFactory создает новую фабрику
func NewFactory(config *Config) *Factory {
    return &Factory{config: config}
}

// CreateDatabaseConnection создает соединение с БД
func (f *Factory) CreateDatabaseConnection() (*sql.DB, error) {
    connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        f.config.Database.Host,
        f.config.Database.Port,
        f.config.Database.User,
        f.config.Database.Password,
        f.config.Database.Name,
    )
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    
    // Проверяем соединение
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, err
    }
    
    return db, nil
}

// CreateUserRepository создает репозиторий пользователей
func (f *Factory) CreateUserRepository(db *sql.DB) UserRepository {
    return NewPostgreSQLUserRepository(db)
}

// CreateUserService создает сервис пользователей
func (f *Factory) CreateUserService(repo UserRepository) *UserService {
    return NewUserService(repo)
}

// CreateNotificationService создает сервис уведомлений
func (f *Factory) CreateNotificationService() NotificationService {
    return NewEmailNotificationService(f.config.SMTP.Host, f.config.SMTP.Port)
}

// CreateRegistrationService создает сервис регистрации
func (f *Factory) CreateRegistrationService(
    userService *UserService,
    notificationService NotificationService,
) *RegistrationService {
    return NewRegistrationService(userService, notificationService)
}

// CreateContainer создает контейнер с автоматическим внедрением
func (f *Factory) CreateContainer() (*Container, error) {
    db, err := f.CreateDatabaseConnection()
    if err != nil {
        return nil, err
    }
    
    container := NewContainer(db)
    return container, nil
}
```

### 2. Жизненный цикл зависимостей

```go
// di/lifecycle.go
package di

import (
    "sync"
)

// LifecycleManager управляет жизненным циклом зависимостей
type LifecycleManager struct {
    container *Container
    once      sync.Once
    started   bool
    mu        sync.RWMutex
}

// NewLifecycleManager создает менеджер жизненного цикла
func NewLifecycleManager(container *Container) *LifecycleManager {
    return &LifecycleManager{
        container: container,
    }
}

// Start запускает приложение
func (lm *LifecycleManager) Start() error {
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    if lm.started {
        return nil
    }
    
    // Инициализируем зависимости в правильном порядке
    lm.container.GetUserRepository()
    lm.container.GetUserService()
    lm.container.GetNotificationService()
    lm.container.GetRegistrationService()
    
    lm.started = true
    return nil
}

// Stop останавливает приложение
func (lm *LifecycleManager) Stop() error {
    lm.mu.Lock()
    defer lm.mu.Unlock()
    
    if !lm.started {
        return nil
    }
    
    // Освобождаем ресурсы
    err := lm.container.Close()
    lm.started = false
    
    return err
}

// GracefulShutdown плавное завершение работы
func (lm *LifecycleManager) GracefulShutdown() error {
    return lm.Stop()
}

// HealthCheck проверяет состояние зависимостей
func (lm *LifecycleManager) HealthCheck() map[string]bool {
    lm.mu.RLock()
    defer lm.mu.RUnlock()
    
    health := make(map[string]bool)
    
    // Проверяем соединение с БД
    if lm.container.db != nil {
        health["database"] = lm.container.db.Ping() == nil
    } else {
        health["database"] = false
    }
    
    // Проверяем другие зависимости
    health["user_repository"] = lm.container.userRepository != nil
    health["user_service"] = lm.container.userService != nil
    health["notification_service"] = lm.container.notificationService != nil
    health["registration_service"] = lm.container.registrationService != nil
    
    return health
}
```

### 3. Конфигурация зависимостей

```go
// di/configuration.go
package di

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
)

// DependencyConfig конфигурация зависимостей
type DependencyConfig struct {
    Database struct {
        Type     string `json:"type"` // postgres, mysql, inmemory
        Host     string `json:"host"`
        Port     int    `json:"port"`
        User     string `json:"user"`
        Password string `json:"password"`
        Name     string `json:"name"`
    } `json:"database"`
    
    Notification struct {
        Type string `json:"type"` // email, sms, push
        SMTP struct {
            Host string `json:"host"`
            Port int    `json:"port"`
        } `json:"smtp"`
    } `json:"notification"`
    
    Features struct {
        EnableCaching  bool `json:"enable_caching"`
        EnableLogging  bool `json:"enable_logging"`
        EnableMetrics  bool `json:"enable_metrics"`
    } `json:"features"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(filename string) (*DependencyConfig, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config DependencyConfig
    err = json.Unmarshal(data, &config)
    if err != nil {
        return nil, err
    }
    
    return &config, nil
}

// ConfigurableContainer контейнер с поддержкой конфигурации
type ConfigurableContainer struct {
    *Container
    config *DependencyConfig
}

// NewConfigurableContainer создает контейнер с конфигурацией
func NewConfigurableContainer(config *DependencyConfig) (*ConfigurableContainer, error) {
    // Создаем базовый контейнер
    var db *sql.DB
    var err error
    
    // В зависимости от конфигурации создаем разные реализации
    switch config.Database.Type {
    case "inmemory":
        // Для тестов используем in-memory реализацию
        container := NewContainer(nil)
        return &ConfigurableContainer{
            Container: container,
            config:    config,
        }, nil
    case "postgres":
        // Создаем реальное соединение с БД
        connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
            config.Database.Host,
            config.Database.Port,
            config.Database.User,
            config.Database.Password,
            config.Database.Name,
        )
        
        db, err = sql.Open("postgres", connStr)
        if err != nil {
            return nil, err
        }
        
        if err := db.Ping(); err != nil {
            db.Close()
            return nil, err
        }
    default:
        return nil, fmt.Errorf("unsupported database type: %s", config.Database.Type)
    }
    
    container := NewContainer(db)
    return &ConfigurableContainer{
        Container: container,
        config:    config,
    }, nil
}

// GetConfigurableUserRepository получает репозиторий в зависимости от конфигурации
func (c *ConfigurableContainer) GetConfigurableUserRepository() UserRepository {
    c.mu.RLock()
    if c.userRepository != nil {
        defer c.mu.RUnlock()
        return c.userRepository
    }
    c.mu.RUnlock()
    
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.userRepository == nil {
        switch c.config.Database.Type {
        case "inmemory":
            c.userRepository = NewInMemoryUserRepository()
        case "postgres":
            c.userRepository = NewPostgreSQLUserRepository(c.db)
        default:
            // По умолчанию используем PostgreSQL
            c.userRepository = NewPostgreSQLUserRepository(c.db)
        }
    }
    
    return c.userRepository
}
```

## Тестирование с DI

### 1. Модульные тесты с моками

```go
// di/services_test.go
package di

import (
    "context"
    "database/sql"
    "testing"
)

// MockUserRepository мок репозитория пользователей
type MockUserRepository struct {
    GetUserFunc    func(ctx context.Context, id int) (*User, error)
    CreateUserFunc func(ctx context.Context, user *User) error
    UpdateUserFunc func(ctx context.Context, user *User) error
    DeleteUserFunc func(ctx context.Context, id int) error
}

func (m *MockUserRepository) GetUser(ctx context.Context, id int) (*User, error) {
    if m.GetUserFunc != nil {
        return m.GetUserFunc(ctx, id)
    }
    return nil, nil
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *User) error {
    if m.CreateUserFunc != nil {
        return m.CreateUserFunc(ctx, user)
    }
    return nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *User) error {
    if m.UpdateUserFunc != nil {
        return m.UpdateUserFunc(ctx, user)
    }
    return nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int) error {
    if m.DeleteUserFunc != nil {
        return m.DeleteUserFunc(ctx, id)
    }
    return nil
}

// MockNotificationService мок сервиса уведомлений
type MockNotificationService struct {
    SendWelcomeEmailFunc       func(ctx context.Context, user *User) error
    SendPasswordResetEmailFunc func(ctx context.Context, user *User) error
}

func (m *MockNotificationService) SendWelcomeEmail(ctx context.Context, user *User) error {
    if m.SendWelcomeEmailFunc != nil {
        return m.SendWelcomeEmailFunc(ctx, user)
    }
    return nil
}

func (m *MockNotificationService) SendPasswordResetEmail(ctx context.Context, user *User) error {
    if m.SendPasswordResetEmailFunc != nil {
        return m.SendPasswordResetEmailFunc(ctx, user)
    }
    return nil
}

func TestUserService_GetUser(t *testing.T) {
    // Создаем мок репозитория
    mockRepo := &MockUserRepository{
        GetUserFunc: func(ctx context.Context, id int) (*User, error) {
            if id == 1 {
                return &User{ID: 1, Name: "John Doe", Email: "john@example.com"}, nil
            }
            return nil, nil
        },
    }
    
    // Создаем сервис с моком
    service := NewUserService(mockRepo)
    
    // Тестируем
    ctx := context.Background()
    user, err := service.GetUser(ctx, 1)
    if err != nil {
        t.Fatalf("GetUser failed: %v", err)
    }
    
    if user == nil {
        t.Fatal("Expected user, got nil")
    }
    
    if user.Name != "John Doe" {
        t.Errorf("Expected name 'John Doe', got '%s'", user.Name)
    }
}

func TestUserService_CreateUser(t *testing.T) {
    // Создаем мок репозитория
    var createdUser *User
    mockRepo := &MockUserRepository{
        CreateUserFunc: func(ctx context.Context, user *User) error {
            user.ID = 1
            createdUser = user
            return nil
        },
    }
    
    // Создаем сервис с моком
    service := NewUserService(mockRepo)
    
    // Тестируем
    ctx := context.Background()
    user, err := service.CreateUser(ctx, "Jane Doe", "jane@example.com")
    if err != nil {
        t.Fatalf("CreateUser failed: %v", err)
    }
    
    if user.ID != 1 {
        t.Errorf("Expected ID 1, got %d", user.ID)
    }
    
    if user.Name != "Jane Doe" {
        t.Errorf("Expected name 'Jane Doe', got '%s'", user.Name)
    }
    
    // Проверяем, что репозиторий получил правильные данные
    if createdUser == nil {
        t.Fatal("User was not created in repository")
    }
    
    if createdUser.Name != "Jane Doe" {
        t.Errorf("Repository received wrong name: expected 'Jane Doe', got '%s'", createdUser.Name)
    }
}

func TestRegistrationService_RegisterUser(t *testing.T) {
    // Создаем моки
    mockUserRepo := &MockUserRepository{
        CreateUserFunc: func(ctx context.Context, user *User) error {
            user.ID = 1
            return nil
        },
    }
    
    var welcomeEmailSent bool
    mockNotification := &MockNotificationService{
        SendWelcomeEmailFunc: func(ctx context.Context, user *User) error {
            welcomeEmailSent = true
            return nil
        },
    }
    
    // Создаем сервисы с моками
    userService := NewUserService(mockUserRepo)
    registrationService := NewRegistrationService(userService, mockNotification)
    
    // Тестируем
    ctx := context.Background()
    user, err := registrationService.RegisterUser(ctx, "New User", "new@example.com")
    if err != nil {
        t.Fatalf("RegisterUser failed: %v", err)
    }
    
    if user.ID != 1 {
        t.Errorf("Expected ID 1, got %d", user.ID)
    }
    
    if !welcomeEmailSent {
        t.Error("Welcome email was not sent")
    }
}
```

### 2. Интеграционные тесты DI

```go
// integration/di_test.go
package integration

import (
    "context"
    "database/sql"
    "testing"
    "time"
    _ "github.com/lib/pq"
    "yourproject/di"
)

func setupTestDB(t *testing.T) *sql.DB {
    // Подключение к тестовой базе данных
    db, err := sql.Open("postgres", "user=test dbname=test_di sslmode=disable")
    if err != nil {
        t.Fatal(err)
    }
    
    // Создаем тестовые таблицы
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            email VARCHAR(100) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        t.Fatal(err)
    }
    
    // Очищаем таблицу перед тестом
    _, err = db.Exec("TRUNCATE users RESTART IDENTITY")
    if err != nil {
        t.Fatal(err)
    }
    
    return db
}

func TestDIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    db := setupTestDB(t)
    defer db.Close()
    
    // Создаем контейнер
    container := di.NewContainer(db)
    defer container.Close()
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Тестируем создание пользователя через DI
    t.Run("CreateUser", func(t *testing.T) {
        userService := container.GetUserService()
        
        user, err := userService.CreateUser(ctx, "Integration Test User", "integration@example.com")
        if err != nil {
            t.Fatalf("CreateUser failed: %v", err)
        }
        
        if user.ID == 0 {
            t.Error("User ID should be set")
        }
        
        if user.Name != "Integration Test User" {
            t.Errorf("Expected name 'Integration Test User', got '%s'", user.Name)
        }
    })
    
    // Тестируем получение пользователя через DI
    t.Run("GetUser", func(t *testing.T) {
        userService := container.GetUserService()
        
        // Сначала создаем пользователя
        user, err := userService.CreateUser(ctx, "Get Test User", "get@example.com")
        if err != nil {
            t.Fatalf("CreateUser failed: %v", err)
        }
        
        // Затем получаем его
        retrievedUser, err := userService.GetUser(ctx, user.ID)
        if err != nil {
            t.Fatalf("GetUser failed: %v", err)
        }
        
        if retrievedUser == nil {
            t.Fatal("User not found")
        }
        
        if retrievedUser.ID != user.ID {
            t.Errorf("Expected ID %d, got %d", user.ID, retrievedUser.ID)
        }
        
        if retrievedUser.Name != "Get Test User" {
            t.Errorf("Expected name 'Get Test User', got '%s'", retrievedUser.Name)
        }
    })
    
    // Тестируем регистрацию через DI
    t.Run("Registration", func(t *testing.T) {
        // Создаем контейнер с моками для уведомлений
        mockNotification := &MockNotificationService{
            SendWelcomeEmailFunc: func(ctx context.Context, user *User) error {
                // В реальном тесте здесь можно проверить параметры
                return nil
            },
        }
        
        userService := container.GetUserService()
        registrationService := di.NewRegistrationService(userService, mockNotification)
        
        user, err := registrationService.RegisterUser(ctx, "Registration Test User", "register@example.com")
        if err != nil {
            t.Fatalf("RegisterUser failed: %v", err)
        }
        
        if user.ID == 0 {
            t.Error("User ID should be set")
        }
    })
}

func TestConfigurableDI(t *testing.T) {
    // Тестируем конфигурацию с in-memory репозиторием
    config := &di.DependencyConfig{}
    config.Database.Type = "inmemory"
    config.Notification.Type = "email"
    config.Notification.SMTP.Host = "localhost"
    config.Notification.SMTP.Port = 1025
    
    container, err := di.NewConfigurableContainer(config)
    if err != nil {
        t.Fatalf("Failed to create configurable container: %v", err)
    }
    defer container.Close()
    
    ctx := context.Background()
    
    // Тестируем работу с in-memory репозиторием
    userService := container.GetUserService()
    
    user, err := userService.CreateUser(ctx, "InMemory User", "inmemory@example.com")
    if err != nil {
        t.Fatalf("CreateUser failed: %v", err)
    }
    
    if user.ID == 0 {
        t.Error("User ID should be set")
    }
    
    // Получаем пользователя
    retrievedUser, err := userService.GetUser(ctx, user.ID)
    if err != nil {
        t.Fatalf("GetUser failed: %v", err)
    }
    
    if retrievedUser == nil {
        t.Fatal("User not found in in-memory repository")
    }
    
    if retrievedUser.Name != "InMemory User" {
        t.Errorf("Expected name 'InMemory User', got '%s'", retrievedUser.Name)
    }
}

func TestLifecycleManagement(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    container := di.NewContainer(db)
    lifecycle := di.NewLifecycleManager(container)
    
    // Тестируем запуск
    err := lifecycle.Start()
    if err != nil {
        t.Fatalf("Lifecycle start failed: %v", err)
    }
    
    // Проверяем состояние здоровья
    health := lifecycle.HealthCheck()
    if !health["database"] {
        t.Error("Database should be healthy")
    }
    
    if !health["user_service"] {
        t.Error("User service should be initialized")
    }
    
    // Тестируем остановку
    err = lifecycle.Stop()
    if err != nil {
        t.Fatalf("Lifecycle stop failed: %v", err)
    }
    
    // Проверяем, что сервисы больше не доступны
    health = lifecycle.HealthCheck()
    if health["database"] {
        t.Error("Database should be closed")
    }
}
```

## Лучшие практики DI

### 1. Интерфейсы и абстракции

```go
// di/best_practices.go
package di

import (
    "context"
)

// ХОРОШО - интерфейсы с конкретными методами
type OrderRepository interface {
    GetOrder(ctx context.Context, id int) (*Order, error)
    CreateOrder(ctx context.Context, order *Order) error
    UpdateOrder(ctx context.Context, order *Order) error
    DeleteOrder(ctx context.Context, id int) error
    ListOrders(ctx context.Context, userID int) ([]*Order, error)
}

// ПЛОХО - слишком общий интерфейс
type Repository interface {
    Get(interface{}) (interface{}, error)
    Create(interface{}) error
    Update(interface{}) error
    Delete(interface{}) error
}

// ХОРОШО - интерфейсы в пакетах потребителей
// order/service.go
type UserRepository interface {
    GetUser(ctx context.Context, id int) (*User, error)
}

type OrderService struct {
    userRepo UserRepository // Интерфейс определен в том же пакете
    orderRepo OrderRepository
}
```

### 2. Композиция вместо наследования

```go
// ХОРОШО - композиция
type PaymentService struct {
    paymentProcessor PaymentProcessor
    logger          Logger
    metrics         Metrics
}

func NewPaymentService(processor PaymentProcessor, logger Logger, metrics Metrics) *PaymentService {
    return &PaymentService{
        paymentProcessor: processor,
        logger:          logger,
        metrics:         metrics,
    }
}

// ПЛОХО - наследование (Go не поддерживает наследование)
type BaseService struct {
    logger Logger
}

type PaymentService struct {
    BaseService // Плохая практика
    // ...
}
```

### 3. Явные зависимости

```go
// ХОРОШО - явные зависимости
func NewOrderService(
    userRepo UserRepository,
    productRepo ProductRepository,
    paymentService PaymentService,
    logger Logger,
) *OrderService {
    return &OrderService{
        userRepo:       userRepo,
        productRepo:    productRepo,
        paymentService: paymentService,
        logger:         logger,
    }
}

// ПЛОХО - скрытые зависимости
func NewOrderService() *OrderService {
    return &OrderService{
        userRepo:    NewUserRepository(),    // Скрытая зависимость
        productRepo: NewProductRepository(), // Скрытая зависимость
        // ...
    }
}
```

## Распространенные ошибки и их решение

### 1. Циклические зависимости

```go
// ПЛОХО - циклическая зависимость
type UserService struct {
    orderService *OrderService // UserService зависит от OrderService
}

type OrderService struct {
    userService *UserService // OrderService зависит от UserService
}

// ХОРОШО - решение через интерфейсы
type UserProvider interface {
    GetUser(ctx context.Context, id int) (*User, error)
}

type OrderProvider interface {
    GetOrder(ctx context.Context, id int) (*Order, error)
}

type UserService struct {
    userProvider   UserProvider   // Интерфейс вместо конкретной реализации
    orderProvider  OrderProvider  // Интерфейс для доступа к заказам
}

type OrderService struct {
    orderProvider  OrderProvider  // Интерфейс вместо конкретной реализации
    userProvider   UserProvider   // Интерфейс для доступа к пользователям
}
```

### 2. Слишком много зависимостей

```go
// ПЛОХО - слишком много зависимостей
func NewUserService(
    repo1 Repository1,
    repo2 Repository2,
    repo3 Repository3,
    repo4 Repository4,
    service1 Service1,
    service2 Service2,
    service3 Service3,
    logger Logger,
    metrics Metrics,
    cache Cache,
    config Config,
    // ... еще 10 зависимостей
) *UserService {
    // ...
}

// ХОРОШО - группировка зависимостей
type UserServiceDependencies struct {
    Repositories struct {
        UserRepo    UserRepository
        OrderRepo   OrderRepository
        ProductRepo ProductRepository
    }
    Services struct {
        NotificationService NotificationService
        PaymentService      PaymentService
    }
    Logger  Logger
    Metrics Metrics
}

func NewUserService(deps UserServiceDependencies) *UserService {
    return &UserService{
        userRepo:            deps.Repositories.UserRepo,
        orderRepo:           deps.Repositories.OrderRepo,
        notificationService: deps.Services.NotificationService,
        logger:              deps.Logger,
        metrics:             deps.Metrics,
    }
}
```

### 3. Неправильное управление жизненным циклом

```go
// ПЛОХО - неправильное управление жизненным циклом
func BadLifecycle() {
    db, _ := sql.Open("postgres", "connection_string")
    userService := NewUserService(NewPostgreSQLUserRepository(db))
    
    // db.Close() никогда не вызывается - утечка ресурсов
    _ = userService
}

// ХОРОШО - правильное управление жизненным циклом
func GoodLifecycle() {
    container, err := NewConfigurableContainer(config)
    if err != nil {
        log.Fatal(err)
    }
    defer container.Close()
    
    userService := container.GetUserService()
    // Работаем с userService...
}
```

## Мониторинг и отладка DI

### 1. Логирование зависимостей

```go
// di/logging.go
package di

import (
    "log"
)

// LoggedContainer контейнер с логированием
type LoggedContainer struct {
    *Container
}

func NewLoggedContainer(db *sql.DB) *LoggedContainer {
    log.Printf("Creating container with database connection")
    container := NewContainer(db)
    return &LoggedContainer{Container: container}
}

func (c *LoggedContainer) GetUserRepository() UserRepository {
    log.Printf("Getting UserRepository")
    return c.Container.GetUserRepository()
}

func (c *LoggedContainer) GetUserService() *UserService {
    log.Printf("Getting UserService")
    return c.Container.GetUserService()
}

func (c *LoggedContainer) GetNotificationService() NotificationService {
    log.Printf("Getting NotificationService")
    return c.Container.GetNotificationService()
}

func (c *LoggedContainer) GetRegistrationService() *RegistrationService {
    log.Printf("Getting RegistrationService")
    return c.Container.GetRegistrationService()
}
```

### 2. Метрики DI

```go
// di/metrics.go
package di

import (
    "sync"
    "time"
)

// DIMetrics метрики DI
type DIMetrics struct {
    dependencyCreations map[string]int64
    creationTimes       map[string]time.Duration
    mu                  sync.RWMutex
}

var metrics = &DIMetrics{
    dependencyCreations: make(map[string]int64),
    creationTimes:       make(map[string]time.Duration),
}

func (m *DIMetrics) RecordCreation(dependency string, duration time.Duration) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.dependencyCreations[dependency]++
    m.creationTimes[dependency] += duration
}

func (m *DIMetrics) GetMetrics() (map[string]int64, map[string]time.Duration) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    creations := make(map[string]int64)
    times := make(map[string]time.Duration)
    
    for k, v := range m.dependencyCreations {
        creations[k] = v
    }
    
    for k, v := range m.creationTimes {
        times[k] = v
    }
    
    return creations, times
}

// InstrumentedContainer контейнер с метриками
type InstrumentedContainer struct {
    *Container
}

func NewInstrumentedContainer(db *sql.DB) *InstrumentedContainer {
    start := time.Now()
    container := NewContainer(db)
    duration := time.Since(start)
    
    metrics.RecordCreation("Container", duration)
    return &InstrumentedContainer{Container: container}
}

func (c *InstrumentedContainer) GetUserRepository() UserRepository {
    start := time.Now()
    repo := c.Container.GetUserRepository()
    duration := time.Since(start)
    
    metrics.RecordCreation("UserRepository", duration)
    return repo
}

func (c *InstrumentedContainer) GetUserService() *UserService {
    start := time.Now()
    service := c.Container.GetUserService()
    duration := time.Since(start)
    
    metrics.RecordCreation("UserService", duration)
    return service
}
```

## См. также

- [Архитектура приложений](../concepts/architecture.md) - принципы проектирования
- [Тестирование](../concepts/testing.md) - как тестировать с DI
- [Интерфейсы](../concepts/interface.md) - основы интерфейсов в Go
- [Практические примеры](../examples/di) - примеры кода