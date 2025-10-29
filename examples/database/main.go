package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// User модель пользователя
type User struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

// Database структура для работы с БД
type Database struct {
	db *sql.DB
}

// NewDatabase создает новое подключение к БД
func NewDatabase(dataSourceName string) (*Database, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	
	// Проверяем подключение
	if err := db.Ping(); err != nil {
		return nil, err
	}
	
	return &Database{db: db}, nil
}

// Init создает таблицы
func (d *Database) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err := d.db.Exec(query)
	return err
}

// CreateUser создает нового пользователя
func (d *Database) CreateUser(name, email string) (int64, error) {
	query := `INSERT INTO users (name, email) VALUES (?, ?)`
	result, err := d.db.Exec(query, name, email)
	if err != nil {
		return 0, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	
	return id, nil
}

// GetUserByID получает пользователя по ID
func (d *Database) GetUserByID(id int) (*User, error) {
	query := `SELECT id, name, email, created_at FROM users WHERE id = ?`
	row := d.db.QueryRow(query, id)
	
	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// GetAllUsers получает всех пользователей
func (d *Database) GetAllUsers() ([]User, error) {
	query := `SELECT id, name, email, created_at FROM users`
	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, nil
}

// UpdateUser обновляет пользователя
func (d *Database) UpdateUser(id int, name, email string) error {
	query := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	_, err := d.db.Exec(query, name, email, id)
	return err
}

// DeleteUser удаляет пользователя
func (d *Database) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := d.db.Exec(query, id)
	return err
}

// Close закрывает подключение к БД
func (d *Database) Close() error {
	return d.db.Close()
}

// Пример 1: Базовое подключение и операции
func basicDatabaseOperations() {
	fmt.Println("=== Базовое подключение и операции ===")
	
	// Создаем подключение к SQLite в памяти
	db, err := NewDatabase(":memory:")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	
	// Инициализируем таблицы
	if err := db.Init(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	
	// Создаем пользователей
	id1, err := db.CreateUser("Иван Иванов", "ivan@example.com")
	if err != nil {
		log.Fatal("Ошибка создания пользователя:", err)
	}
	fmt.Printf("Создан пользователь с ID: %d\n", id1)
	
	id2, err := db.CreateUser("Мария Петрова", "maria@example.com")
	if err != nil {
		log.Fatal("Ошибка создания пользователя:", err)
	}
	fmt.Printf("Создан пользователь с ID: %d\n", id2)
	
	// Получаем пользователя по ID
	user, err := db.GetUserByID(int(id1))
	if err != nil {
		log.Fatal("Ошибка получения пользователя:", err)
	}
	fmt.Printf("Получен пользователь: %+v\n", user)
	
	// Получаем всех пользователей
	users, err := db.GetAllUsers()
	if err != nil {
		log.Fatal("Ошибка получения пользователей:", err)
	}
	fmt.Println("Все пользователи:")
	for _, u := range users {
		fmt.Printf("  %+v\n", u)
	}
}

// Пример 2: Транзакции
func transactionsExample() {
	fmt.Println("\n=== Транзакции ===")
	
	db, err := NewDatabase(":memory:")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	
	if err := db.Init(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	
	// Начинаем транзакцию
	tx, err := db.db.Begin()
	if err != nil {
		log.Fatal("Ошибка начала транзакции:", err)
	}
	
	// Выполняем несколько операций в транзакции
	_, err = tx.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, "Пользователь 1", "user1@example.com")
	if err != nil {
		tx.Rollback()
		log.Fatal("Ошибка вставки 1:", err)
	}
	
	_, err = tx.Exec(`INSERT INTO users (name, email) VALUES (?, ?)`, "Пользователь 2", "user2@example.com")
	if err != nil {
		tx.Rollback()
		log.Fatal("Ошибка вставки 2:", err)
	}
	
	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		log.Fatal("Ошибка коммита транзакции:", err)
	}
	
	fmt.Println("Транзакция успешно завершена")
	
	// Проверяем результаты
	users, err := db.GetAllUsers()
	if err != nil {
		log.Fatal("Ошибка получения пользователей:", err)
	}
	fmt.Println("Пользователи после транзакции:")
	for _, u := range users {
		fmt.Printf("  %+v\n", u)
	}
}

// Пример 3: Подготовленные запросы
func preparedStatements() {
	fmt.Println("\n=== Подготовленные запросы ===")
	
	db, err := NewDatabase(":memory:")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	
	if err := db.Init(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	
	// Создаем пользователей
	usersData := []struct {
		name  string
		email string
	}{
		{"Алексей", "alex@example.com"},
		{"Елена", "elena@example.com"},
		{"Дмитрий", "dmitry@example.com"},
	}
	
	// Подготавливаем запрос
	stmt, err := db.db.Prepare(`INSERT INTO users (name, email) VALUES (?, ?)`)
	if err != nil {
		log.Fatal("Ошибка подготовки запроса:", err)
	}
	defer stmt.Close()
	
	// Выполняем запрос несколько раз
	for _, u := range usersData {
		_, err := stmt.Exec(u.name, u.email)
		if err != nil {
			log.Fatal("Ошибка выполнения запроса:", err)
		}
	}
	
	fmt.Println("Пользователи созданы с помощью подготовленного запроса")
	
	// Получаем всех пользователей
	users, err := db.GetAllUsers()
	if err != nil {
		log.Fatal("Ошибка получения пользователей:", err)
	}
	fmt.Println("Все пользователи:")
	for _, u := range users {
		fmt.Printf("  %+v\n", u)
	}
}

// Пример 4: Пулы соединений
func connectionPooling() {
	fmt.Println("\n=== Пулы соединений ===")
	
	// Для демонстрации создадим файловую БД
	db, err := NewDatabase("./test.db")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	
	if err := db.Init(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	
	// Настраиваем пул соединений
	sqlDB := db.db
	sqlDB.SetMaxOpenConns(10) // Максимум 10 открытых соединений
	sqlDB.SetMaxIdleConns(5)  // Максимум 5 соединений в режиме ожидания
	sqlDB.SetConnMaxLifetime(time.Hour) // Максимальное время жизни соединения
	
	fmt.Printf("Максимум открытых соединений: %d\n", sqlDB.Stats().MaxOpenConnections)
	fmt.Printf("Открытые соединения: %d\n", sqlDB.Stats().OpenConnections)
	fmt.Printf("Соединения в режиме ожидания: %d\n", sqlDB.Stats().Idle)
	
	// Создаем несколько пользователей
	for i := 1; i <= 20; i++ {
		_, err := db.CreateUser(fmt.Sprintf("Пользователь %d", i), fmt.Sprintf("user%d@example.com", i))
		if err != nil {
			log.Printf("Ошибка создания пользователя %d: %v", i, err)
		}
	}
	
	fmt.Println("Создано 20 пользователей")
	
	// Проверяем статистику соединений
	stats := sqlDB.Stats()
	fmt.Printf("После операций - Открытые соединения: %d, В режиме ожидания: %d\n", 
		stats.OpenConnections, stats.Idle)
}

// Пример 5: Обработка ошибок БД
func databaseErrorHandling() {
	fmt.Println("\n=== Обработка ошибок БД ===")
	
	db, err := NewDatabase(":memory:")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	
	if err := db.Init(); err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	
	// Создаем пользователя
	id, err := db.CreateUser("Тестовый пользователь", "test@example.com")
	if err != nil {
		log.Printf("Ошибка создания пользователя: %v", err)
		return
	}
	fmt.Printf("Создан пользователь с ID: %d\n", id)
	
	// Пытаемся создать пользователя с тем же email (должна быть ошибка)
	_, err = db.CreateUser("Другой пользователь", "test@example.com")
	if err != nil {
		fmt.Printf("Ожидаемая ошибка уникальности: %v\n", err)
	}
	
	// Пытаемся получить несуществующего пользователя
	_, err = db.GetUserByID(999)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Пользователь не найден (ErrNoRows)")
		} else {
			fmt.Printf("Другая ошибка БД: %v\n", err)
		}
	}
}

// Пример 6: Работа с NULL значениями
func nullValues() {
	fmt.Println("\n=== Работа с NULL значениями ===")
	
	// Создаем таблицу с NULL полями
	db, err := NewDatabase(":memory:")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	
	// Создаем таблицу с NULL полями
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price REAL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	
	_, err = db.db.Exec(query)
	if err != nil {
		log.Fatal("Ошибка создания таблицы:", err)
	}
	
	// Вставляем данные с NULL значениями
	_, err = db.db.Exec(`INSERT INTO products (name, description, price) VALUES (?, ?, ?)`, 
		"Товар 1", nil, 99.99)
	if err != nil {
		log.Fatal("Ошибка вставки:", err)
	}
	
	_, err = db.db.Exec(`INSERT INTO products (name, description, price) VALUES (?, ?, ?)`, 
		"Товар 2", "Описание товара 2", nil)
	if err != nil {
		log.Fatal("Ошибка вставки:", err)
	}
	
	// Получаем данные с NULL значениями
	type Product struct {
		ID          int
		Name        string
		Description sql.NullString
		Price       sql.NullFloat64
		CreatedAt   time.Time
	}
	
	rows, err := db.db.Query(`SELECT id, name, description, price, created_at FROM products`)
	if err != nil {
		log.Fatal("Ошибка запроса:", err)
	}
	defer rows.Close()
	
	fmt.Println("Продукты с возможными NULL значениями:")
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price, &product.CreatedAt)
		if err != nil {
			log.Fatal("Ошибка сканирования:", err)
		}
		
		fmt.Printf("ID: %d, Name: %s", product.ID, product.Name)
		if product.Description.Valid {
			fmt.Printf(", Description: %s", product.Description.String)
		} else {
			fmt.Print(", Description: NULL")
		}
		
		if product.Price.Valid {
			fmt.Printf(", Price: %.2f", product.Price.Float64)
		} else {
			fmt.Print(", Price: NULL")
		}
		fmt.Println()
	}
}

func main() {
	basicDatabaseOperations()
	transactionsExample()
	preparedStatements()
	connectionPooling()
	databaseErrorHandling()
	nullValues()
	
	fmt.Println("\n=== Все примеры работы с БД ===")
	fmt.Println("Для запуска примеров убедитесь, что установлен драйвер: go get github.com/mattn/go-sqlite3")
}