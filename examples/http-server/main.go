package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// User модель пользователя
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// In-memory storage для пользователей
var users = map[int]User{
	1: {ID: 1, Name: "Иван Иванов", Email: "ivan@example.com"},
	2: {ID: 2, Name: "Мария Петрова", Email: "maria@example.com"},
}

var nextID = 3

// Пример 1: Базовый HTTP сервер
func basicHTTPServer() {
	fmt.Println("=== Базовый HTTP сервер ===")
	
	// Обработчик для главной страницы
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Привет, мир! Текущее время: %s\n", time.Now().Format(time.RFC3339))
	})
	
	// Обработчик для API
	http.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"message": "Привет от API",
			"time":    time.Now().Format(time.RFC3339),
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	fmt.Println("Сервер запущен на :8080")
	// Запуск сервера (закомментирован для примера)
	// log.Fatal(http.ListenAndServe(":8080", nil))
}

// Пример 2: REST API для пользователей
func userAPI() {
	fmt.Println("\n=== REST API для пользователей ===")
	
	// Получить всех пользователей
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Возвращаем всех пользователей
			userList := make([]User, 0, len(users))
			for _, user := range users {
				userList = append(userList, user)
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(userList)
			
		case http.MethodPost:
			// Создаем нового пользователя
			var user User
			if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
				http.Error(w, "Неверный JSON", http.StatusBadRequest)
				return
			}
			
			user.ID = nextID
			nextID++
			users[user.ID] = user
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
			
		default:
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		}
	})
	
	// Получить/обновить/удалить конкретного пользователя
	http.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем ID из URL
		idStr := r.URL.Path[len("/api/users/"):]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Неверный ID", http.StatusBadRequest)
			return
		}
		
		switch r.Method {
		case http.MethodGet:
			// Получаем пользователя
			user, exists := users[id]
			if !exists {
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			
		case http.MethodPut:
			// Обновляем пользователя
			user, exists := users[id]
			if !exists {
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
				return
			}
			
			var updatedUser User
			if err := json.NewDecoder(r.Body).Decode(&updatedUser); err != nil {
				http.Error(w, "Неверный JSON", http.StatusBadRequest)
				return
			}
			
			updatedUser.ID = id // Сохраняем оригинальный ID
			users[id] = updatedUser
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedUser)
			
		case http.MethodDelete:
			// Удаляем пользователя
			if _, exists := users[id]; !exists {
				http.Error(w, "Пользователь не найден", http.StatusNotFound)
				return
			}
			
			delete(users, id)
			w.WriteHeader(http.StatusNoContent)
			
		default:
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		}
	})
}

// Middleware для логирования
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Запрос: %s %s", r.Method, r.URL.Path)
		
		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
		
		log.Printf("Завершено за %v", time.Since(start))
	})
}

// Middleware для CORS
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		// Обрабатываем preflight запросы
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Пример 3: Middleware
func middlewareExample() {
	fmt.Println("\n=== Middleware ===")
	
	// Создаем маршрутизатор
	mux := http.NewServeMux()
	
	// Добавляем маршруты
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Оборачиваем маршрутизатор в middleware
	handler := corsMiddleware(loggingMiddleware(mux))
	
	// Создаем сервер
	server := &http.Server{
		Addr:    ":8081",
		Handler: handler,
	}
	
	fmt.Println("Сервер с middleware запущен на :8081")
	// Запуск сервера (закомментирован для примера)
	// log.Fatal(server.ListenAndServe())
}

// Пример 4: Graceful shutdown
func gracefulShutdown() {
	fmt.Println("\n=== Graceful shutdown ===")
	
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"status": "healthy",
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}
	
	// Запуск сервера в отдельной горутине
	go func() {
		fmt.Println("Сервер запущен на :8082")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()
	
	// Здесь мог бы быть код для ожидания сигнала завершения
	// и корректной остановки сервера
	fmt.Println("Для остановки сервера используйте Ctrl+C")
}

// Пример 5: Работа с формами
func formHandling() {
	fmt.Println("\n=== Работа с формами ===")
	
	// Страница с формой
	http.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Отображаем форму
			html := `
<!DOCTYPE html>
<html>
<head>
    <title>Форма пользователя</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>Добавить пользователя</h1>
    <form method="POST" action="/form">
        <label for="name">Имя:</label>
        <input type="text" id="name" name="name" required><br><br>
        
        <label for="email">Email:</label>
        <input type="email" id="email" name="email" required><br><br>
        
        <input type="submit" value="Добавить">
    </form>
</body>
</html>`
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, html)
		} else if r.Method == http.MethodPost {
			// Обрабатываем форму
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Ошибка разбора формы", http.StatusBadRequest)
				return
			}
			
			name := r.FormValue("name")
			email := r.FormValue("email")
			
			// Создаем пользователя
			user := User{
				ID:    nextID,
				Name:  name,
				Email: email,
			}
			nextID++
			users[user.ID] = user
			
			// Перенаправляем на список пользователей
			http.Redirect(w, r, "/api/users", http.StatusSeeOther)
		}
	})
}

// Пример 6: Загрузка файлов
func fileUpload() {
	fmt.Println("\n=== Загрузка файлов ===")
	
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Форма для загрузки файла
			html := `
<!DOCTYPE html>
<html>
<head>
    <title>Загрузка файла</title>
    <meta charset="UTF-8">
</head>
<body>
    <h1>Загрузить файл</h1>
    <form method="POST" enctype="multipart/form-data">
        <input type="file" name="file" required><br><br>
        <input type="submit" value="Загрузить">
    </form>
</body>
</html>`
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, html)
		} else if r.Method == http.MethodPost {
			// Обрабатываем загрузку файла
			file, handler, err := r.FormFile("file")
			if err != nil {
				http.Error(w, "Ошибка получения файла", http.StatusBadRequest)
				return
			}
			defer file.Close()
			
			fmt.Fprintf(w, "Файл загружен успешно!\n")
			fmt.Fprintf(w, "Имя файла: %s\n", handler.Filename)
			fmt.Fprintf(w, "Размер: %d байт\n", handler.Size)
			fmt.Fprintf(w, "Content-Type: %s\n", handler.Header.Get("Content-Type"))
		}
	})
}

// Пример 7: JSON API с валидацией
func jsonAPIWithValidation() {
	fmt.Println("\n=== JSON API с валидацией ===")
	
	http.HandleFunc("/api/users/validated", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Только POST разрешен", http.StatusMethodNotAllowed)
			return
		}
		
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Неверный JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		// Валидация данных
		if user.Name == "" {
			http.Error(w, "Имя обязательно", http.StatusBadRequest)
			return
		}
		
		if user.Email == "" {
			http.Error(w, "Email обязателен", http.StatusBadRequest)
			return
		}
		
		// Проверка формата email (упрощенная)
		if len(user.Email) < 5 || !contains(user.Email, "@") {
			http.Error(w, "Неверный формат email", http.StatusBadRequest)
			return
		}
		
		// Проверка уникальности email
		for _, existingUser := range users {
			if existingUser.Email == user.Email {
				http.Error(w, "Email уже существует", http.StatusBadRequest)
				return
			}
		}
		
		// Создаем пользователя
		user.ID = nextID
		nextID++
		users[user.ID] = user
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	})
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Пример 8: Обработка статических файлов
func staticFiles() {
	fmt.Println("\n=== Обработка статических файлов ===")
	
	// Обслуживание статических файлов
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	
	// Главная страница, ссылающаяся на статические файлы
	http.HandleFunc("/static-page", func(w http.ResponseWriter, r *http.Request) {
		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Статические файлы</title>
    <meta charset="UTF-8">
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <h1>Страница со статическими файлами</h1>
    <img src="/static/image.png" alt="Пример изображения">
    <script src="/static/script.js"></script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	})
}

func main() {
	basicHTTPServer()
	userAPI()
	middlewareExample()
	gracefulShutdown()
	formHandling()
	fileUpload()
	jsonAPIWithValidation()
	staticFiles()
	
	fmt.Println("\n=== Все примеры HTTP серверов ===")
	fmt.Println("Для запуска конкретного примера раскомментируйте соответствующий код в функции main")
}