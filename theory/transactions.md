# Транзакции в Go: Полная теория

## Введение в транзакции

### Что такое транзакции?

Транзакции в базах данных - это **последовательность операций**, которые:
- **Выполняются как единое целое** (атомарность)
- **Сохраняют согласованность** данных
- **Изолированы** от других транзакций
- **Долговечны** (после коммита изменения сохраняются)

### ACID свойства

1. **Atomicity** (Атомарность) - все операции выполняются или ни одна
2. **Consistency** (Согласованность) - данные остаются в согласованном состоянии
3. **Isolation** (Изоляция) - транзакции не влияют друг на друга
4. **Durability** (Долговечность) - изменения сохраняются после коммита

## Основы транзакций в Go

### database/sql и транзакции

Go стандартная библиотека `database/sql` предоставляет **встроенную поддержку** транзакций:

```go
// Базовое использование транзакций
func BasicTransaction(db *sql.DB) error {
    // Начинаем транзакцию
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    
    // Откладываем откат на случай ошибки
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    // Выполняем операции в транзакции
    _, err = tx.Exec("INSERT INTO users (name) VALUES (?)", "John")
    if err != nil {
        return err
    }
    
    _, err = tx.Exec("UPDATE accounts SET balance = balance - 100 WHERE user_id = ?", 1)
    if err != nil {
        return err
    }
    
    // Коммитим транзакцию
    return tx.Commit()
}
```

### Уровни изоляции

```go
// Разные уровни изоляции транзакций
const (
    LevelDefault         = iota // Уровень по умолчанию
    LevelReadUncommitted        // Чтение незафиксированных данных
    LevelReadCommitted          // Чтение зафиксированных данных
    LevelWriteCommitted         // Запись зафиксированных данных
    LevelRepeatableRead         // Повторяемое чтение
    LevelSnapshot               // Снимок
    LevelSerializable           // Сериализуемость
)

func TransactionWithIsolation(db *sql.DB) error {
    // Создаем транзакцию с определенным уровнем изоляции
    opts := &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
        ReadOnly:  false,
    }
    
    tx, err := db.BeginTx(context.Background(), opts)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Работа с транзакцией...
    
    return tx.Commit()
}
```

## Практическая реализация транзакций

### 1. Базовый шаблон транзакций

```go
// database/transaction.go
package database

import (
    "context"
    "database/sql"
)

// TransactionFunc определяет функцию, которая выполняется в транзакции
type TransactionFunc func(tx *sql.Tx) error

// WithTransaction выполняет функцию в транзакции с автоматическим коммитом/откатом
func WithTransaction(db *sql.DB, fn TransactionFunc) error {
    return WithTransactionContext(context.Background(), db, fn)
}

// WithTransactionContext выполняет функцию в транзакции с контекстом
func WithTransactionContext(ctx context.Context, db *sql.DB, fn TransactionFunc) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    
    // Откладываем откат
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p) // Перебрасываем panic
        } else if err != nil {
            tx.Rollback() // Откатываем при ошибке
        } else {
            err = tx.Commit() // Коммитим при успехе
        }
    }()
    
    err = fn(tx)
    return err
}
```

### 2. Репозиторий с поддержкой транзакций

```go
// repository/user.go
package repository

import (
    "context"
    "database/sql"
)

type User struct {
    ID    int
    Name  string
    Email string
}

type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

// CreateUser создает пользователя вне транзакции
func (r *UserRepository) CreateUser(ctx context.Context, user *User) error {
    query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
    return r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&user.ID)
}

// CreateUserTx создает пользователя в транзакции
func (r *UserRepository) CreateUserTx(ctx context.Context, tx *sql.Tx, user *User) error {
    query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
    
    var scanner interface{ Scan(...interface{}) error }
    if tx != nil {
        scanner = tx.QueryRowContext(ctx, query, user.Name, user.Email)
    } else {
        scanner = r.db.QueryRowContext(ctx, query, user.Name, user.Email)
    }
    
    return scanner.Scan(&user.ID)
}

// GetUserByID получает пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
    query := "SELECT id, name, email FROM users WHERE id = $1"
    
    var user User
    err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    
    return &user, nil
}
```

```go
// repository/account.go
package repository

import (
    "context"
    "database/sql"
)

type Account struct {
    ID      int
    UserID  int
    Balance float64
}

type AccountRepository struct {
    db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
    return &AccountRepository{db: db}
}

// CreateAccount создает счет вне транзакции
func (r *AccountRepository) CreateAccount(ctx context.Context, account *Account) error {
    query := "INSERT INTO accounts (user_id, balance) VALUES ($1, $2) RETURNING id"
    return r.db.QueryRowContext(ctx, query, account.UserID, account.Balance).Scan(&account.ID)
}

// CreateAccountTx создает счет в транзакции
func (r *AccountRepository) CreateAccountTx(ctx context.Context, tx *sql.Tx, account *Account) error {
    query := "INSERT INTO accounts (user_id, balance) VALUES ($1, $2) RETURNING id"
    
    var scanner interface{ Scan(...interface{}) error }
    if tx != nil {
        scanner = tx.QueryRowContext(ctx, query, account.UserID, account.Balance)
    } else {
        scanner = r.db.QueryRowContext(ctx, query, account.UserID, account.Balance)
    }
    
    return scanner.Scan(&account.ID)
}

// UpdateBalance обновляет баланс счета
func (r *AccountRepository) UpdateBalance(ctx context.Context, accountID int, amount float64) error {
    query := "UPDATE accounts SET balance = balance + $1 WHERE id = $2"
    _, err := r.db.ExecContext(ctx, query, amount, accountID)
    return err
}

// UpdateBalanceTx обновляет баланс счета в транзакции
func (r *AccountRepository) UpdateBalanceTx(ctx context.Context, tx *sql.Tx, accountID int, amount float64) error {
    query := "UPDATE accounts SET balance = balance + $1 WHERE id = $2"
    
    var execer interface{ ExecContext(context.Context, string, ...interface{}) (sql.Result, error) }
    if tx != nil {
        execer = tx
    } else {
        execer = r.db
    }
    
    _, err := execer.ExecContext(ctx, query, amount, accountID)
    return err
}
```

### 3. Сервисный слой с транзакциями

```go
// service/user.go
package service

import (
    "context"
    "database/sql"
    "yourproject/database"
    "yourproject/repository"
)

type UserService struct {
    db       *sql.DB
    userRepo *repository.UserRepository
    accountRepo *repository.AccountRepository
}

func NewUserService(db *sql.DB) *UserService {
    return &UserService{
        db:          db,
        userRepo:    repository.NewUserRepository(db),
        accountRepo: repository.NewAccountRepository(db),
    }
}

// RegisterUser регистрирует пользователя с созданием счета
func (s *UserService) RegisterUser(ctx context.Context, name, email string, initialBalance float64) error {
    return database.WithTransactionContext(ctx, s.db, func(tx *sql.Tx) error {
        // Создаем пользователя
        user := &repository.User{Name: name, Email: email}
        err := s.userRepo.CreateUserTx(ctx, tx, user)
        if err != nil {
            return err
        }
        
        // Создаем счет для пользователя
        account := &repository.Account{UserID: user.ID, Balance: initialBalance}
        err = s.accountRepo.CreateAccountTx(ctx, tx, account)
        if err != nil {
            return err
        }
        
        return nil
    })
}

// TransferMoney переводит деньги между счетами
func (s *UserService) TransferMoney(ctx context.Context, fromAccountID, toAccountID int, amount float64) error {
    if amount <= 0 {
        return errors.New("сумма должна быть положительной")
    }
    
    return database.WithTransactionContext(ctx, s.db, func(tx *sql.Tx) error {
        // Проверяем баланс отправителя
        var balance float64
        err := tx.QueryRowContext(ctx, 
            "SELECT balance FROM accounts WHERE id = $1", fromAccountID).Scan(&balance)
        if err != nil {
            return err
        }
        
        if balance < amount {
            return errors.New("недостаточно средств")
        }
        
        // Списываем деньги со счета отправителя
        _, err = tx.ExecContext(ctx, 
            "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromAccountID)
        if err != nil {
            return err
        }
        
        // Зачисляем деньги на счет получателя
        _, err = tx.ExecContext(ctx, 
            "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toAccountID)
        if err != nil {
            return err
        }
        
        // Записываем транзакцию в историю
        _, err = tx.ExecContext(ctx, 
            "INSERT INTO transactions (from_account_id, to_account_id, amount) VALUES ($1, $2, $3)",
            fromAccountID, toAccountID, amount)
        if err != nil {
            return err
        }
        
        return nil
    })
}
```

## Расширенные техники транзакций

### 1. Вложенные транзакции (Savepoints)

```go
// database/savepoint.go
package database

import (
    "context"
    "database/sql"
    "fmt"
)

type Savepoint struct {
    tx    *sql.Tx
    name  string
    level int
}

func NewSavepoint(tx *sql.Tx, name string) *Savepoint {
    return &Savepoint{
        tx:   tx,
        name: name,
    }
}

func (s *Savepoint) Create(ctx context.Context) error {
    _, err := s.tx.ExecContext(ctx, fmt.Sprintf("SAVEPOINT %s", s.name))
    return err
}

func (s *Savepoint) Rollback(ctx context.Context) error {
    _, err := s.tx.ExecContext(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", s.name))
    return err
}

func (s *Savepoint) Release(ctx context.Context) error {
    _, err := s.tx.ExecContext(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", s.name))
    return err
}

// Пример использования savepoint
func ComplexTransactionWithSavepoints(db *sql.DB) error {
    return WithTransaction(db, func(tx *sql.Tx) error {
        ctx := context.Background()
        
        // Основная операция
        _, err := tx.ExecContext(ctx, "INSERT INTO users (name) VALUES ($1)", "User1")
        if err != nil {
            return err
        }
        
        // Создаем savepoint
        sp := NewSavepoint(tx, "sp1")
        err = sp.Create(ctx)
        if err != nil {
            return err
        }
        
        // Операция, которая может завершиться ошибкой
        _, err = tx.ExecContext(ctx, "INSERT INTO risky_table (data) VALUES ($1)", "risky_data")
        if err != nil {
            // Откатываем к savepoint, но продолжаем транзакцию
            sp.Rollback(ctx)
            // Выполняем альтернативную операцию
            _, err = tx.ExecContext(ctx, "INSERT INTO fallback_table (data) VALUES ($1)", "fallback_data")
            if err != nil {
                return err
            }
        } else {
            // Освобождаем savepoint
            sp.Release(ctx)
        }
        
        // Еще одна операция
        _, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance + 100 WHERE user_id = 1")
        if err != nil {
            return err
        }
        
        return nil
    })
}
```

### 2. Распределенные транзакции

```go
// database/distributed_tx.go
package database

import (
    "context"
    "database/sql"
    "errors"
    "sync"
)

type DistributedTransaction struct {
    connections []*sql.DB
    transactions []*sql.Tx
    committed   bool
    mu          sync.Mutex
}

func NewDistributedTransaction(connections ...*sql.DB) *DistributedTransaction {
    return &DistributedTransaction{
        connections: connections,
        transactions: make([]*sql.Tx, len(connections)),
    }
}

func (dt *DistributedTransaction) Begin(ctx context.Context) error {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    // Начинаем транзакции во всех базах данных
    for i, conn := range dt.connections {
        tx, err := conn.BeginTx(ctx, nil)
        if err != nil {
            // Откатываем уже начатые транзакции
            dt.rollbackStarted()
            return err
        }
        dt.transactions[i] = tx
    }
    
    return nil
}

func (dt *DistributedTransaction) rollbackStarted() {
    for i, tx := range dt.transactions {
        if tx != nil {
            tx.Rollback()
            dt.transactions[i] = nil
        }
    }
}

func (dt *DistributedTransaction) Exec(ctx context.Context, dbIndex int, query string, args ...interface{}) (sql.Result, error) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if dbIndex >= len(dt.transactions) || dt.transactions[dbIndex] == nil {
        return nil, errors.New("invalid database index or transaction not started")
    }
    
    return dt.transactions[dbIndex].ExecContext(ctx, query, args...)
}

func (dt *DistributedTransaction) Query(ctx context.Context, dbIndex int, query string, args ...interface{}) (*sql.Rows, error) {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if dbIndex >= len(dt.transactions) || dt.transactions[dbIndex] == nil {
        return nil, errors.New("invalid database index or transaction not started")
    }
    
    return dt.transactions[dbIndex].QueryContext(ctx, query, args...)
}

func (dt *DistributedTransaction) QueryRow(ctx context.Context, dbIndex int, query string, args ...interface{}) *sql.Row {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if dbIndex >= len(dt.transactions) || dt.transactions[dbIndex] == nil {
        // Возвращаем пустую строку в случае ошибки
        return &sql.Row{}
    }
    
    return dt.transactions[dbIndex].QueryRowContext(ctx, query, args...)
}

func (dt *DistributedTransaction) Commit(ctx context.Context) error {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if dt.committed {
        return errors.New("transaction already committed")
    }
    
    // Коммитим все транзакции
    for i, tx := range dt.transactions {
        if tx != nil {
            err := tx.Commit()
            if err != nil {
                // Если одна транзакция не коммитится, пытаемся откатить остальные
                dt.rollbackRemaining(i)
                return err
            }
            dt.transactions[i] = nil
        }
    }
    
    dt.committed = true
    return nil
}

func (dt *DistributedTransaction) rollbackRemaining(fromIndex int) {
    for i := fromIndex; i < len(dt.transactions); i++ {
        if dt.transactions[i] != nil {
            dt.transactions[i].Rollback()
            dt.transactions[i] = nil
        }
    }
}

func (dt *DistributedTransaction) Rollback(ctx context.Context) error {
    dt.mu.Lock()
    defer dt.mu.Unlock()
    
    if dt.committed {
        return errors.New("cannot rollback committed transaction")
    }
    
    dt.rollbackStarted()
    return nil
}

// Пример использования распределенной транзакции
func DistributedTransfer(userDB, accountDB *sql.DB, fromUserID, toUserID int, amount float64) error {
    dt := NewDistributedTransaction(userDB, accountDB)
    ctx := context.Background()
    
    err := dt.Begin(ctx)
    if err != nil {
        return err
    }
    
    defer func() {
        if err != nil {
            dt.Rollback(ctx)
        }
    }()
    
    // Проверяем баланс в accountDB
    var balance float64
    err = dt.QueryRow(ctx, 1, "SELECT balance FROM accounts WHERE user_id = $1", fromUserID).Scan(&balance)
    if err != nil {
        return err
    }
    
    if balance < amount {
        return errors.New("insufficient funds")
    }
    
    // Списываем деньги
    _, err = dt.Exec(ctx, 1, "UPDATE accounts SET balance = balance - $1 WHERE user_id = $2", amount, fromUserID)
    if err != nil {
        return err
    }
    
    // Зачисляем деньги
    _, err = dt.Exec(ctx, 1, "UPDATE accounts SET balance = balance + $1 WHERE user_id = $2", amount, toUserID)
    if err != nil {
        return err
    }
    
    // Записываем в историю в userDB
    _, err = dt.Exec(ctx, 0, 
        "INSERT INTO transaction_history (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)",
        fromUserID, toUserID, amount)
    if err != nil {
        return err
    }
    
    return dt.Commit(ctx)
}
```

### 3. Транзакции с таймаутами и контекстами

```go
// database/transaction_with_timeout.go
package database

import (
    "context"
    "database/sql"
    "time"
)

type TransactionConfig struct {
    Timeout         time.Duration
    IsolationLevel  sql.IsolationLevel
    ReadOnly        bool
}

func WithConfiguredTransaction(db *sql.DB, config TransactionConfig, fn TransactionFunc) error {
    // Создаем контекст с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
    defer cancel()
    
    // Настройки транзакции
    opts := &sql.TxOptions{
        Isolation: config.IsolationLevel,
        ReadOnly:  config.ReadOnly,
    }
    
    tx, err := db.BeginTx(ctx, opts)
    if err != nil {
        return err
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        } else if err != nil {
            tx.Rollback()
        } else {
            // Проверяем контекст перед коммитом
            select {
            case <-ctx.Done():
                tx.Rollback()
                err = ctx.Err()
            default:
                err = tx.Commit()
            }
        }
    }()
    
    err = fn(tx)
    return err
}

// Пример использования с таймаутом
func LongRunningTransaction(db *sql.DB) error {
    config := TransactionConfig{
        Timeout:        30 * time.Second,
        IsolationLevel: sql.LevelReadCommitted,
        ReadOnly:       false,
    }
    
    return WithConfiguredTransaction(db, config, func(tx *sql.Tx) error {
        ctx := context.Background()
        
        // Долгая операция
        _, err := tx.ExecContext(ctx, "CALL long_running_procedure()")
        if err != nil {
            return err
        }
        
        // Еще операции...
        return nil
    })
}
```

## Тестирование транзакций

### 1. Модульные тесты транзакций

```go
// repository/user_test.go
package repository

import (
    "context"
    "database/sql"
    "testing"
    _ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatal(err)
    }
    
    // Создаем тестовые таблицы
    _, err = db.Exec(`
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL
        );
        
        CREATE TABLE accounts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            balance REAL NOT NULL DEFAULT 0,
            FOREIGN KEY (user_id) REFERENCES users(id)
        );
    `)
    if err != nil {
        t.Fatal(err)
    }
    
    return db
}

func TestUserRepository_CreateUserTx(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewUserRepository(db)
    
    // Тест без транзакции
    user := &User{Name: "Test User", Email: "test@example.com"}
    err := repo.CreateUser(context.Background(), user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    if user.ID == 0 {
        t.Error("User ID should be set after creation")
    }
    
    // Тест с транзакцией
    tx, err := db.Begin()
    if err != nil {
        t.Fatal(err)
    }
    defer tx.Rollback()
    
    user2 := &User{Name: "Test User 2", Email: "test2@example.com"}
    err = repo.CreateUserTx(context.Background(), tx, user2)
    if err != nil {
        t.Fatalf("Failed to create user in transaction: %v", err)
    }
    
    if user2.ID == 0 {
        t.Error("User ID should be set after creation in transaction")
    }
    
    // Коммитим транзакцию
    err = tx.Commit()
    if err != nil {
        t.Fatalf("Failed to commit transaction: %v", err)
    }
}

func TestUserRepository_TransactionRollback(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    tx, err := db.Begin()
    if err != nil {
        t.Fatal(err)
    }
    
    repo := NewUserRepository(db)
    user := &User{Name: "Test User", Email: "rollback@example.com"}
    err = repo.CreateUserTx(context.Background(), tx, user)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }
    
    // Откатываем транзакцию
    err = tx.Rollback()
    if err != nil {
        t.Fatalf("Failed to rollback transaction: %v", err)
    }
    
    // Проверяем, что пользователь не был создан
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", "rollback@example.com").Scan(&count)
    if err != nil {
        t.Fatalf("Failed to query users: %v", err)
    }
    
    if count != 0 {
        t.Error("User should not exist after rollback")
    }
}
```

### 2. Интеграционные тесты транзакций

```go
// service/user_integration_test.go
package service

import (
    "context"
    "database/sql"
    "testing"
    "time"
    _ "github.com/lib/pq"
)

func setupIntegrationDB(t *testing.T) *sql.DB {
    // Подключение к тестовой базе данных PostgreSQL
    db, err := sql.Open("postgres", "user=test dbname=test_transactions sslmode=disable")
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
        );
        
        CREATE TABLE IF NOT EXISTS accounts (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL REFERENCES users(id),
            balance DECIMAL(10,2) NOT NULL DEFAULT 0,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE TABLE IF NOT EXISTS transactions (
            id SERIAL PRIMARY KEY,
            from_account_id INTEGER NOT NULL REFERENCES accounts(id),
            to_account_id INTEGER NOT NULL REFERENCES accounts(id),
            amount DECIMAL(10,2) NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `)
    if err != nil {
        t.Fatal(err)
    }
    
    // Очищаем таблицы перед тестом
    _, err = db.Exec("TRUNCATE transactions, accounts, users RESTART IDENTITY CASCADE")
    if err != nil {
        t.Fatal(err)
    }
    
    return db
}

func TestUserService_RegisterUser(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    db := setupIntegrationDB(t)
    defer db.Close()
    
    service := NewUserService(db)
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Регистрируем пользователя
    err := service.RegisterUser(ctx, "John Doe", "john@example.com", 1000.0)
    if err != nil {
        t.Fatalf("Failed to register user: %v", err)
    }
    
    // Проверяем, что пользователь создан
    var userID int
    err = db.QueryRow("SELECT id FROM users WHERE email = $1", "john@example.com").Scan(&userID)
    if err != nil {
        t.Fatalf("Failed to find user: %v", err)
    }
    
    if userID == 0 {
        t.Error("User ID should be set")
    }
    
    // Проверяем, что счет создан
    var accountID int
    var balance float64
    err = db.QueryRow("SELECT id, balance FROM accounts WHERE user_id = $1", userID).Scan(&accountID, &balance)
    if err != nil {
        t.Fatalf("Failed to find account: %v", err)
    }
    
    if accountID == 0 {
        t.Error("Account ID should be set")
    }
    
    if balance != 1000.0 {
        t.Errorf("Expected balance 1000.0, got %f", balance)
    }
}

func TestUserService_TransferMoney(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    db := setupIntegrationDB(t)
    defer db.Close()
    
    service := NewUserService(db)
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Регистрируем двух пользователей
    err := service.RegisterUser(ctx, "Alice", "alice@example.com", 1000.0)
    if err != nil {
        t.Fatalf("Failed to register Alice: %v", err)
    }
    
    err = service.RegisterUser(ctx, "Bob", "bob@example.com", 500.0)
    if err != nil {
        t.Fatalf("Failed to register Bob: %v", err)
    }
    
    // Получаем ID счетов
    var aliceAccountID, bobAccountID int
    err = db.QueryRow("SELECT id FROM accounts WHERE user_id = (SELECT id FROM users WHERE email = $1)", 
        "alice@example.com").Scan(&aliceAccountID)
    if err != nil {
        t.Fatalf("Failed to get Alice's account: %v", err)
    }
    
    err = db.QueryRow("SELECT id FROM accounts WHERE user_id = (SELECT id FROM users WHERE email = $1)", 
        "bob@example.com").Scan(&bobAccountID)
    if err != nil {
        t.Fatalf("Failed to get Bob's account: %v", err)
    }
    
    // Переводим деньги
    err = service.TransferMoney(ctx, aliceAccountID, bobAccountID, 300.0)
    if err != nil {
        t.Fatalf("Failed to transfer money: %v", err)
    }
    
    // Проверяем балансы
    var aliceBalance, bobBalance float64
    err = db.QueryRow("SELECT balance FROM accounts WHERE id = $1", aliceAccountID).Scan(&aliceBalance)
    if err != nil {
        t.Fatalf("Failed to get Alice's balance: %v", err)
    }
    
    err = db.QueryRow("SELECT balance FROM accounts WHERE id = $1", bobAccountID).Scan(&bobBalance)
    if err != nil {
        t.Fatalf("Failed to get Bob's balance: %v", err)
    }
    
    if aliceBalance != 700.0 {
        t.Errorf("Expected Alice's balance 700.0, got %f", aliceBalance)
    }
    
    if bobBalance != 800.0 {
        t.Errorf("Expected Bob's balance 800.0, got %f", bobBalance)
    }
    
    // Проверяем историю транзакций
    var transactionCount int
    err = db.QueryRow("SELECT COUNT(*) FROM transactions WHERE from_account_id = $1 AND to_account_id = $2 AND amount = 300.0",
        aliceAccountID, bobAccountID).Scan(&transactionCount)
    if err != nil {
        t.Fatalf("Failed to check transaction history: %v", err)
    }
    
    if transactionCount != 1 {
        t.Errorf("Expected 1 transaction record, got %d", transactionCount)
    }
}

func TestUserService_TransferMoney_InsufficientFunds(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    db := setupIntegrationDB(t)
    defer db.Close()
    
    service := NewUserService(db)
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Регистрируем пользователя с нулевым балансом
    err := service.RegisterUser(ctx, "Charlie", "charlie@example.com", 0.0)
    if err != nil {
        t.Fatalf("Failed to register Charlie: %v", err)
    }
    
    // Регистрируем пользователя с положительным балансом
    err = service.RegisterUser(ctx, "David", "david@example.com", 500.0)
    if err != nil {
        t.Fatalf("Failed to register David: %v", err)
    }
    
    // Получаем ID счетов
    var charlieAccountID, davidAccountID int
    err = db.QueryRow("SELECT id FROM accounts WHERE user_id = (SELECT id FROM users WHERE email = $1)", 
        "charlie@example.com").Scan(&charlieAccountID)
    if err != nil {
        t.Fatalf("Failed to get Charlie's account: %v", err)
    }
    
    err = db.QueryRow("SELECT id FROM accounts WHERE user_id = (SELECT id FROM users WHERE email = $1)", 
        "david@example.com").Scan(&davidAccountID)
    if err != nil {
        t.Fatalf("Failed to get David's account: %v", err)
    }
    
    // Пытаемся перевести больше, чем есть
    err = service.TransferMoney(ctx, charlieAccountID, davidAccountID, 100.0)
    if err == nil {
        t.Fatal("Expected error for insufficient funds, got nil")
    }
    
    if err.Error() != "недостаточно средств" {
        t.Errorf("Expected 'недостаточно средств' error, got '%s'", err.Error())
    }
    
    // Проверяем, что балансы не изменились
    var charlieBalance, davidBalance float64
    err = db.QueryRow("SELECT balance FROM accounts WHERE id = $1", charlieAccountID).Scan(&charlieBalance)
    if err != nil {
        t.Fatalf("Failed to get Charlie's balance: %v", err)
    }
    
    err = db.QueryRow("SELECT balance FROM accounts WHERE id = $1", davidAccountID).Scan(&davidBalance)
    if err != nil {
        t.Fatalf("Failed to get David's balance: %v", err)
    }
    
    if charlieBalance != 0.0 {
        t.Errorf("Expected Charlie's balance 0.0, got %f", charlieBalance)
    }
    
    if davidBalance != 500.0 {
        t.Errorf("Expected David's balance 500.0, got %f", davidBalance)
    }
}
```

## Лучшие практики транзакций

### 1. Обработка ошибок в транзакциях

```go
// Правильная обработка ошибок в транзакциях
func ProperTransactionHandling(db *sql.DB, userID int, amount float64) error {
    tx, err := db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    // Используем defer для автоматического отката
    defer func() {
        if err != nil {
            if rbErr := tx.Rollback(); rbErr != nil {
                // Логируем ошибку отката, но не возвращаем её
                log.Printf("Failed to rollback transaction: %v", rbErr)
            }
        }
    }()
    
    // Выполняем операции
    _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE user_id = $2", amount, userID)
    if err != nil {
        return fmt.Errorf("failed to update account: %w", err)
    }
    
    _, err = tx.Exec("INSERT INTO transactions (user_id, amount) VALUES ($1, $2)", userID, amount)
    if err != nil {
        return fmt.Errorf("failed to insert transaction: %w", err)
    }
    
    // Коммитим транзакцию
    if err = tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    return nil
}
```

### 2. Таймауты и контексты

```go
// Использование контекстов с таймаутами
func TransactionWithTimeout(db *sql.DB, userID int) error {
    // Создаем контекст с таймаутом
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Используем контекст с транзакцией
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    
    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            // Проверяем контекст перед коммитом
            select {
            case <-ctx.Done():
                tx.Rollback()
                err = ctx.Err()
            default:
                err = tx.Commit()
            }
        }
    }()
    
    // Выполняем операции с контекстом
    _, err = tx.ExecContext(ctx, "UPDATE users SET last_activity = NOW() WHERE id = $1", userID)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 3. Логирование транзакций

```go
// Логирование транзакций для отладки
func TransactionWithLogging(db *sql.DB, userID int, amount float64) error {
    start := time.Now()
    log.Printf("Starting transaction for user %d, amount %f", userID, amount)
    
    tx, err := db.Begin()
    if err != nil {
        log.Printf("Failed to begin transaction: %v", err)
        return err
    }
    
    defer func() {
        duration := time.Since(start)
        if err != nil {
            log.Printf("Transaction failed for user %d after %v: %v", userID, duration, err)
            tx.Rollback()
        } else {
            log.Printf("Transaction completed for user %d in %v", userID, duration)
            tx.Commit()
        }
    }()
    
    // Операции с логированием
    log.Printf("Updating account balance for user %d", userID)
    _, err = tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE user_id = $2", amount, userID)
    if err != nil {
        return err
    }
    
    log.Printf("Recording transaction for user %d", userID)
    _, err = tx.Exec("INSERT INTO transactions (user_id, amount) VALUES ($1, $2)", userID, amount)
    if err != nil {
        return err
    }
    
    return nil
}
```

## Распространенные ошибки и их решение

### 1. Забытый откат транзакций

```go
// ПЛОХО - забытый откат
func BadTransactionHandling(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    
    // Если здесь произойдет ошибка, транзакция не будет откачена
    _, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "John")
    if err != nil {
        return err // Забыли tx.Rollback()!
    }
    
    return tx.Commit()
}

// ХОРОШО - правильный откат
func GoodTransactionHandling(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    
    defer func() {
        if err != nil {
            tx.Rollback()
        }
    }()
    
    _, err = tx.Exec("INSERT INTO users (name) VALUES ($1)", "John")
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

### 2. Использование одной транзакции для разных соединений

```go
// ПЛОХО - одна транзакция для разных соединений
func BadDistributedTransaction(db1, db2 *sql.DB) error {
    tx1, err := db1.Begin()
    if err != nil {
        return err
    }
    defer tx1.Rollback()
    
    tx2, err := db2.Begin() // Отдельная транзакция!
    if err != nil {
        return err
    }
    defer tx2.Rollback()
    
    // Эти операции не атомарны!
    _, err = tx1.Exec("UPDATE table1 SET value = 1")
    if err != nil {
        return err
    }
    
    _, err = tx2.Exec("UPDATE table2 SET value = 1")
    if err != nil {
        return err
    }
    
    // Коммитим по отдельности - риск несогласованности
    err = tx1.Commit()
    if err != nil {
        return err
    }
    
    return tx2.Commit()
}

// ХОРОШО - использование распределенных транзакций
func GoodDistributedTransaction(db1, db2 *sql.DB) error {
    // Используем специализированную реализацию распределенных транзакций
    dt := NewDistributedTransaction(db1, db2)
    ctx := context.Background()
    
    err := dt.Begin(ctx)
    if err != nil {
        return err
    }
    
    defer func() {
        if err != nil {
            dt.Rollback(ctx)
        }
    }()
    
    _, err = dt.Exec(ctx, 0, "UPDATE table1 SET value = 1")
    if err != nil {
        return err
    }
    
    _, err = dt.Exec(ctx, 1, "UPDATE table2 SET value = 1")
    if err != nil {
        return err
    }
    
    return dt.Commit(ctx)
}
```

### 3. Долгоживущие транзакции

```go
// ПЛОХО - долгоживущая транзакция
func BadLongTransaction(db *sql.DB) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Транзакция держится открытым соединение всё это время
    time.Sleep(10 * time.Second) // Плохая идея!
    
    _, err = tx.Exec("UPDATE users SET last_seen = NOW()")
    if err != nil {
        return err
    }
    
    return tx.Commit()
}

// ХОРОШО - минимизация времени транзакции
func GoodShortTransaction(db *sql.DB) error {
    // Выполняем долгую операцию вне транзакции
    time.Sleep(10 * time.Second)
    
    // Транзакция только для критической секции
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    _, err = tx.Exec("UPDATE users SET last_seen = NOW()")
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Мониторинг транзакций

### 1. Метрики транзакций

```go
// database/transaction_metrics.go
package database

import (
    "sync"
    "time"
)

type TransactionMetrics struct {
    TotalTransactions int64
    SuccessfulTransactions int64
    FailedTransactions int64
    AverageDuration time.Duration
    MaxDuration time.Duration
    mu sync.RWMutex
}

var metrics = &TransactionMetrics{}

func (tm *TransactionMetrics) RecordTransaction(duration time.Duration, success bool) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    tm.TotalTransactions++
    if success {
        tm.SuccessfulTransactions++
    } else {
        tm.FailedTransactions++
    }
    
    // Обновляем среднюю продолжительность
    if tm.TotalTransactions == 1 {
        tm.AverageDuration = duration
        tm.MaxDuration = duration
    } else {
        tm.AverageDuration = time.Duration(
            (int64(tm.AverageDuration)*(tm.TotalTransactions-1) + int64(duration)) / tm.TotalTransactions)
        
        if duration > tm.MaxDuration {
            tm.MaxDuration = duration
        }
    }
}

func (tm *TransactionMetrics) GetMetrics() (int64, int64, int64, time.Duration, time.Duration) {
    tm.mu.RLock()
    defer tm.mu.RUnlock()
    
    return tm.TotalTransactions, tm.SuccessfulTransactions, tm.FailedTransactions, 
           tm.AverageDuration, tm.MaxDuration
}

// Оберточный метод для транзакций с метриками
func WithTransactionMetrics(db *sql.DB, fn TransactionFunc) error {
    start := time.Now()
    success := false
    
    defer func() {
        duration := time.Since(start)
        metrics.RecordTransaction(duration, success)
    }()
    
    err := WithTransaction(db, fn)
    if err == nil {
        success = true
    }
    
    return err
}
```

### 2. Логирование долгих транзакций

```go
// database/slow_transaction_logger.go
package database

import (
    "context"
    "log"
    "time"
)

func WithSlowTransactionLogging(db *sql.DB, threshold time.Duration, fn TransactionFunc) error {
    start := time.Now()
    
    err := WithTransaction(db, fn)
    
    duration := time.Since(start)
    if duration > threshold {
        log.Printf("SLOW TRANSACTION: took %v (threshold: %v)", duration, threshold)
    }
    
    return err
}

// Использование с контекстом
func WithSlowTransactionLoggingContext(ctx context.Context, db *sql.DB, threshold time.Duration, fn TransactionFunc) error {
    start := time.Now()
    
    err := WithTransactionContext(ctx, db, fn)
    
    duration := time.Since(start)
    select {
    case <-ctx.Done():
        log.Printf("TRANSACTION CANCELLED: took %v before cancellation", duration)
    default:
        if duration > threshold {
            log.Printf("SLOW TRANSACTION: took %v (threshold: %v)", duration, threshold)
        }
    }
    
    return err
}
```

## См. также

- [Работа с базами данных](../concepts/database.md) - основы работы с БД
- [Пул соединений](../concepts/connection-pooling.md) - управление соединениями
- [Тестирование](../concepts/testing.md) - как тестировать код
- [Контекст](../concepts/context.md) - управление жизненным циклом операций
- [Практические примеры](../examples/transactions) - примеры кода