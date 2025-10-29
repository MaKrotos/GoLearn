package main

import (
	"fmt"
	"sync"
	"time"
)

// Пример 1: Базовые каналы
func basicChannels() {
	fmt.Println("=== Базовые каналы ===")
	
	// Создаем канал для передачи строк
	ch := make(chan string)
	
	// Горутина, которая отправляет данные в канал
	go func() {
		ch <- "Привет из канала!"
	}()
	
	// Получаем данные из канала
	message := <-ch
	fmt.Println("Получено:", message)
}

// Пример 2: Буферизированные каналы
func bufferedChannels() {
	fmt.Println("\n=== Буферизированные каналы ===")
	
	// Создаем буферизированный канал с емкостью 3
	ch := make(chan int, 3)
	
	// Отправляем данные в канал (не блокируется, пока буфер не заполнен)
	ch <- 1
	ch <- 2
	ch <- 3
	
	// Закрываем канал
	close(ch)
	
	// Получаем все данные из канала
	for value := range ch {
		fmt.Println("Получено:", value)
	}
}

// Пример 3: Направленные каналы
func directionalChannels() {
	fmt.Println("\n=== Направленные каналы ===")
	
	ch := make(chan int)
	
	// Горутина, которая только отправляет данные
	go func(out chan<- int) {
		for i := 1; i <= 3; i++ {
			out <- i * 10
			time.Sleep(time.Second)
		}
		close(out)
	}(ch)
	
	// Горутина, которая только получает данные
	go func(in <-chan int) {
		for value := range in {
			fmt.Println("Получено:", value)
		}
	}(ch)
	
	// Ждем немного, чтобы горутины успели выполниться
	time.Sleep(5 * time.Second)
}

// Пример 4: Select для множественных каналов
func selectExample() {
	fmt.Println("\n=== Select для множественных каналов ===")
	
	ch1 := make(chan string)
	ch2 := make(chan string)
	
	// Горутины, отправляющие данные в разные каналы
	go func() {
		time.Sleep(2 * time.Second)
		ch1 <- "Сообщение из первого канала"
	}()
	
	go func() {
		time.Sleep(1 * time.Second)
		ch2 <- "Сообщение из второго канала"
	}()
	
	// Используем select для ожидания данных из любого канала
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Println("Получено из ch1:", msg1)
		case msg2 := <-ch2:
			fmt.Println("Получено из ch2:", msg2)
		case <-time.After(3 * time.Second):
			fmt.Println("Таймаут!")
		}
	}
}

// Пример 5: Паттерн Worker Pool
func workerPool() {
	fmt.Println("\n=== Паттерн Worker Pool ===")
	
	const numWorkers = 3
	const numJobs = 10
	
	// Каналы для jobs и results
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	
	// Запускаем workers
	var wg sync.WaitGroup
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				fmt.Printf("Worker %d обрабатывает job %d\n", workerID, job)
				// Имитируем работу
				time.Sleep(time.Second)
				results <- job * 2 // Возвращаем результат
			}
		}(i)
	}
	
	// Отправляем jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)
	
	// Закрываем results канал после завершения всех workers
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Получаем results
	for result := range results {
		fmt.Println("Результат:", result)
	}
}

func main() {
	basicChannels()
	bufferedChannels()
	directionalChannels()
	selectExample()
	workerPool()
}