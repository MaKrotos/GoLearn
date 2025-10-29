package main

import (
	"fmt"
	"sync"
	"time"
)

// Пример 1: Базовые горутины
func basicGoroutines() {
	fmt.Println("=== Базовые горутины ===")
	
	// Запуск горутины
	go func() {
		fmt.Println("Привет из горутины!")
	}()
	
	// Ждем немного, чтобы горутина успела выполниться
	time.Sleep(time.Second)
}

// Пример 2: Горутины с WaitGroup
func goroutinesWithWaitGroup() {
	fmt.Println("\n=== Горутины с WaitGroup ===")
	
	var wg sync.WaitGroup
	
	// Запускаем 3 горутины
	for i := 1; i <= 3; i++ {
		wg.Add(1) // Увеличиваем счетчик
		go func(id int) {
			defer wg.Done() // Уменьшаем счетчик при завершении
			fmt.Printf("Горутина %d выполняется\n", id)
			time.Sleep(time.Duration(id) * time.Second)
			fmt.Printf("Горутина %d завершена\n", id)
		}(i)
	}
	
	// Ждем завершения всех горутин
	wg.Wait()
	fmt.Println("Все горутины завершены")
}

// Пример 3: Работа с данными в горутинах
func goroutinesWithData() {
	fmt.Println("\n=== Работа с данными в горутинах ===")
	
	// Создаем канал для передачи результатов
	results := make(chan string, 3)
	
	// Запускаем горутины, которые отправляют результаты в канал
	for i := 1; i <= 3; i++ {
		go func(id int) {
			// Имитируем работу
			time.Sleep(time.Duration(id) * time.Second)
			results <- fmt.Sprintf("Результат от горутины %d", id)
		}(i)
	}
	
	// Получаем результаты
	for i := 1; i <= 3; i++ {
		result := <-results
		fmt.Println(result)
	}
	
	close(results)
}

func main() {
	basicGoroutines()
	goroutinesWithWaitGroup()
	goroutinesWithData()
}