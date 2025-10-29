package main

import (
	"fmt"
	"sync"
	"time"
)

// Counter простой счетчик
type Counter struct {
	mu    sync.Mutex
	value int
}

// Increment увеличивает значение счетчика
func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Value возвращает текущее значение счетчика
func (c *Counter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Пример 1: Гонка данных без синхронизации
func raceConditionExample() {
	fmt.Println("=== Гонка данных без синхронизации ===")
	
	counter := 0
	var wg sync.WaitGroup
	
	// Запускаем 1000 горутин, каждая увеличивает счетчик
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // Гонка данных!
		}()
	}
	
	wg.Wait()
	fmt.Printf("Ожидаемое значение: 1000, Фактическое значение: %d\n", counter)
}

// Пример 2: Использование Mutex
func mutexExample() {
	fmt.Println("\n=== Использование Mutex ===")
	
	counter := &Counter{}
	var wg sync.WaitGroup
	
	// Запускаем 1000 горутин с синхронизацией
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	
	wg.Wait()
	fmt.Printf("Ожидаемое значение: 1000, Фактическое значение: %d\n", counter.Value())
}

// Пример 3: RWMutex для чтения/записи
type RWCounter struct {
	mu    sync.RWMutex
	value int
}

// Increment увеличивает значение (требует эксклюзивной блокировки)
func (c *RWCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Value возвращает значение (разделяемая блокировка)
func (c *RWCounter) Value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Пример использования RWMutex
func rwMutexExample() {
	fmt.Println("\n=== Использование RWMutex ===")
	
	counter := &RWCounter{}
	var wg sync.WaitGroup
	
	// Запускаем горутины для записи
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	
	// Запускаем горутины для чтения
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Множественные чтения могут происходить одновременно
			for j := 0; j < 100; j++ {
				_ = counter.Value()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}
	
	wg.Wait()
	fmt.Printf("Финальное значение счетчика: %d\n", counter.Value())
}

// Пример 4: Once для однократного выполнения
func onceExample() {
	fmt.Println("\n=== Использование Once ===")
	
	var once sync.Once
	var calls int
	
	// Функция, которая будет вызвана только один раз
	initialize := func() {
		calls++
		fmt.Println("Инициализация выполнена")
	}
	
	var wg sync.WaitGroup
	// Запускаем несколько горутин
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(initialize)
		}()
	}
	
	wg.Wait()
	fmt.Printf("Функция initialize была вызвана %d раз(а)\n", calls)
}

// Пример 5: WaitGroup для ожидания завершения
func waitGroupExample() {
	fmt.Println("\n=== Использование WaitGroup ===")
	
	var wg sync.WaitGroup
	results := make(chan string, 5)
	
	// Запускаем несколько горутин
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Имитируем работу
			time.Sleep(time.Duration(id) * time.Second)
			results <- fmt.Sprintf("Горутина %d завершена", id)
		}(i)
	}
	
	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Получаем результаты
	for result := range results {
		fmt.Println(result)
	}
}

// Пример 6: Cond для условного ожидания
func condExample() {
	fmt.Println("\n=== Использование Cond ===")
	
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false
	
	var wg sync.WaitGroup
	
	// Горутина, которая ждет сигнала
	wg.Add(1)
	go func() {
		defer wg.Done()
		mu.Lock()
		for !ready {
			fmt.Println("Горутина ждет...")
			cond.Wait() // Ждем сигнала
		}
		fmt.Println("Горутина получила сигнал!")
		mu.Unlock()
	}()
	
	// Главная горутина, которая отправляет сигнал
	time.Sleep(time.Second)
	mu.Lock()
	ready = true
	fmt.Println("Отправляем сигнал...")
	cond.Signal() // Отправляем сигнал
	mu.Unlock()
	
	wg.Wait()
}

func main() {
	raceConditionExample()
	mutexExample()
	rwMutexExample()
	onceExample()
	waitGroupExample()
	condExample()
}