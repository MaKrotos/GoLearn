package main

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Пример 1: Базовое использование контекста
func basicContext() {
	fmt.Println("=== Базовое использование контекста ===")
	
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel() // Всегда отменяем контекст
	
	// Запускаем операцию в отдельной горутине
	done := make(chan bool)
	go func() {
		// Имитируем долгую операцию
		time.Sleep(3 * time.Second)
		done <- true
	}()
	
	// Ждем завершения или отмены контекста
	select {
	case <-done:
		fmt.Println("Операция завершена успешно")
	case <-ctx.Done():
		fmt.Println("Операция отменена:", ctx.Err())
	}
}

// Пример 2: Передача значений через контекст
func contextWithValue() {
	fmt.Println("\n=== Передача значений через контекст ===")
	
	// Создаем контекст с значениями
	type key string
	userIDKey := key("userID")
	requestIDKey := key("requestID")
	
	ctx := context.WithValue(context.Background(), userIDKey, "user123")
	ctx = context.WithValue(ctx, requestIDKey, "req456")
	
	// Функция, которая использует значения из контекста
	processRequest := func(ctx context.Context) {
		userID := ctx.Value(userIDKey).(string)
		requestID := ctx.Value(requestIDKey).(string)
		fmt.Printf("Обработка запроса userID=%s, requestID=%s\n", userID, requestID)
	}
	
	processRequest(ctx)
}

// Пример 3: Контекст в HTTP сервере
func httpServerWithContext() {
	fmt.Println("\n=== Контекст в HTTP сервере ===")
	
	// Создаем сервер с обработчиком
	mux := http.NewServeMux()
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		// Получаем контекст из запроса
		ctx := r.Context()
		
		// Имитируем долгую операцию с проверкой контекста
		select {
		case <-time.After(3 * time.Second):
			fmt.Fprintln(w, "Данные получены")
		case <-ctx.Done():
			// Контекст отменен (клиент отключился)
			http.Error(w, "Запрос отменен", http.StatusRequestTimeout)
			return
		}
	})
	
	// Создаем сервер с таймаутом
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	
	// Запускаем сервер в отдельной горутине
	go func() {
		fmt.Println("Сервер запущен на :8080")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("Ошибка сервера: %v\n", err)
		}
	}()
	
	// Ждем немного, затем останавливаем сервер
	time.Sleep(2 * time.Second)
	
	// Создаем контекст с таймаутом для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Останавливаем сервер
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Ошибка остановки сервера: %v\n", err)
	} else {
		fmt.Println("Сервер остановлен корректно")
	}
}

// Пример 4: Контекст с отменой
func contextWithCancel() {
	fmt.Println("\n=== Контекст с отменой ===")
	
	// Создаем контекст с функцией отмены
	ctx, cancel := context.WithCancel(context.Background())
	
	// Запускаем worker
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Worker остановлен:", ctx.Err())
				done <- true
				return
			default:
				// Выполняем работу
				fmt.Print(".")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()
	
	// Ждем немного, затем отменяем контекст
	time.Sleep(3 * time.Second)
	fmt.Println("\nОтменяем контекст...")
	cancel()
	
	// Ждем завершения worker
	<-done
}

// Пример 5: Контекст с дедлайном
func contextWithDeadline() {
	fmt.Println("\n=== Контекст с дедлайном ===")
	
	// Создаем контекст с дедлайном (через 3 секунды)
	deadline := time.Now().Add(3 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	
	// Запускаем операцию
	done := make(chan string)
	go func() {
		// Имитируем работу
		time.Sleep(5 * time.Second)
		done <- "Работа завершена"
	}()
	
	// Ждем результат или дедлайн
	select {
	case result := <-done:
		fmt.Println("Результат:", result)
	case <-ctx.Done():
		fmt.Println("Дедлайн превышен:", ctx.Err())
	}
}

// Пример 6: Контекст в цепочке вызовов
func contextInChain() {
	fmt.Println("\n=== Контекст в цепочке вызовов ===")
	
	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Добавляем значения в контекст
	type key string
	operationKey := key("operation")
	ctx = context.WithValue(ctx, operationKey, "data-processing")
	
	// Цепочка вызовов
	result := processData(ctx)
	fmt.Println("Результат обработки:", result)
}

func processData(ctx context.Context) string {
	// Получаем значение из контекста
	operation := ctx.Value(key("operation")).(string)
	fmt.Printf("Начинаем операцию: %s\n", operation)
	
	// Вызываем следующую функцию в цепочке
	return fetchAndProcessData(ctx)
}

func fetchAndProcessData(ctx context.Context) string {
	// Проверяем, не отменен ли контекст
	select {
	case <-ctx.Done():
		return "Операция отменена: " + ctx.Err().Error()
	default:
	}
	
	// Имитируем получение данных
	fmt.Println("Получаем данные...")
	time.Sleep(2 * time.Second)
	
	// Проверяем контекст снова
	select {
	case <-ctx.Done():
		return "Операция отменена: " + ctx.Err().Error()
	default:
	}
	
	// Имитируем обработку данных
	fmt.Println("Обрабатываем данные...")
	time.Sleep(2 * time.Second)
	
	return "Данные успешно обработаны"
}

func main() {
	basicContext()
	contextWithValue()
	httpServerWithContext()
	contextWithCancel()
	contextWithDeadline()
	contextInChain()
}