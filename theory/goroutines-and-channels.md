# Горутины и каналы: Полная теория

## Введение в конкурентность

### Что такое конкурентность?

Конкурентность - это **способность выполнять несколько задач одновременно**. В Go это реализуется через:
- **Горутины** - легковесные потоки
- **Каналы** - средства коммуникации между горутинами

### Почему Go хорош для конкурентности?

1. **Легковесность** - горутины стартуют с ~2KB стека
2. **Встроенный планировщик** - управляет выполнением горутин
3. **Простота использования** - ключевое слово `go` для создания горутин
4. **Безопасность** - каналы для безопасной коммуникации

## Горутины (Goroutines)

### Архитектура планировщика Go

Планировщик Go использует модель **M:N**:
- **M** рабочих потоков ОС
- **N** горутин
- **GOMAXPROCS** определяет количество рабочих потоков

```
Потоки ОС (M)        Горутины (N)
┌─────────────┐     ┌─────────────┐
│   Thread 1  │────▶│ Goroutine 1 │
│             │     │ Goroutine 2 │
│             │     │ Goroutine 3 │
├─────────────┤     └─────────────┘
│   Thread 2  │────▶┌─────────────┐
│             │     │ Goroutine 4 │
│             │     │ Goroutine 5 │
├─────────────┤     └─────────────┘
│   Thread 3  │────▶┌─────────────┐
│             │     │ Goroutine 6 │
│             │     │ Goroutine 7 │
└─────────────┘     └─────────────┘
```

### Создание горутин

```go
// Базовое создание
go func() {
    fmt.Println("Hello from goroutine")
}()

// С параметрами
go func(name string) {
    fmt.Printf("Hello, %s\n", name)
}("Иван")

// С замыканием
for i := 0; i < 5; i++ {
    go func() {
        fmt.Printf("Number: %d\n", i) // ОШИБКА: все горутины используют одно i
    }()
}

// Правильно
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Printf("Number: %d\n", n)
    }(i)
}
```

### Жизненный цикл горутины

1. **Создание** - `go` statement
2. **Выполнение** - планировщик назначает поток
3. **Блокировка** - ожидание (канал, мьютекс, I/O)
4. **Продолжение** - возобновление выполнения
5. **Завершение** - выход из функции

### Планировщик Go (Goroutine Scheduler)

#### Компоненты планировщика:

1. **G (Goroutine)** - структура горутины
2. **M (Machine)** - рабочий поток ОС
3. **P (Processor)** - процессор планировщика

#### Работа планировщика:

```go
// Установка количества процессоров
runtime.GOMAXPROCS(4) // Использовать 4 потока ОС

// Получение текущего значения
fmt.Println(runtime.GOMAXPROCS(0)) // Выводит: 4
```

## Каналы (Channels)

### Архитектура каналов

Каналы реализуют принцип **CSP (Communicating Sequential Processes)**:
- **Коммуникация** через каналы вместо разделяемой памяти
- **Синхронизация** через операции отправки/получения

### Типы каналов

#### 1. Небуферизированные каналы

```go
// Создание
ch := make(chan int)

// Свойства
// - Нет буфера
// - Синхронная передача
// - Отправитель блокируется до получения
```

Пример работы:
```go
func main() {
    ch := make(chan string)
    
    go func() {
        fmt.Println("Горутина: готовлюсь отправить сообщение")
        ch <- "Привет!" // Блокируется до получения
        fmt.Println("Горутина: сообщение отправлено")
    }()
    
    fmt.Println("Главный поток: жду сообщение")
    msg := <-ch // Разблокирует отправителя
    fmt.Println("Главный поток:", msg)
}
```

#### 2. Буферизированные каналы

```go
// Создание
ch := make(chan int, 3) // Буфер на 3 элемента

// Свойства
// - Есть буфер
// - Асинхронная передача (до заполнения буфера)
// - Отправитель блокируется при переполнении
```

Пример работы:
```go
func main() {
    ch := make(chan int, 2)
    
    ch <- 1 // Не блокируется
    ch <- 2 // Не блокируется
    // ch <- 3 // Заблокировалось бы
    
    fmt.Println(<-ch) // Получаем 1
    ch <- 3           // Теперь не блокируется
}
```

### Операции с каналами

#### 1. Отправка и получение

```go
// Отправка
ch <- value

// Получение
value := <-ch

// Получение с проверкой закрытия
value, ok := <-ch
if !ok {
    // Канал закрыт
}
```

#### 2. Закрытие каналов

```go
close(ch)

// После закрытия:
// - Нельзя отправлять данные
// - Можно получать оставшиеся данные
// - Получение возвращает zero value
```

#### 3. Range по каналу

```go
// Получение всех значений до закрытия
for value := range ch {
    fmt.Println(value)
}
```

### Паттерны использования каналов

#### 1. Worker Pool

```go
func workerPool() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)
    
    // Запускаем воркеров
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }
    
    // Отправляем задачи
    for j := 1; j <= 5; j++ {
        jobs <- j
    }
    close(jobs)
    
    // Получаем результаты
    for a := 1; a <= 5; a++ {
        <-results
    }
}

func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        fmt.Printf("Воркер %d начал задачу %d\n", id, j)
        time.Sleep(time.Second)
        fmt.Printf("Воркер %d завершил задачу %d\n", id, j)
        results <- j * 2
    }
}
```

#### 2. Fan-in / Fan-out

```go
// Fan-out: один канал -> много каналов
func fanOut(input <-chan int, numWorkers int) []<-chan int {
    workers := make([]<-chan int, numWorkers)
    for i := 0; i < numWorkers; i++ {
        workers[i] = worker(input)
    }
    return workers
}

// Fan-in: много каналов -> один канал
func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                out <- val
            }
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```

#### 3. Pipeline

```go
func pipeline() {
    // Создаем pipeline: numbers -> multiply -> add -> результаты
    numbers := generate(1, 2, 3, 4, 5)
    multiplied := multiply(numbers, 2)
    added := add(multiplied, 10)
    
    // Потребляем результаты
    for result := range added {
        fmt.Println(result)
    }
}

func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

func multiply(in <-chan int, multiplier int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * multiplier
        }
    }()
    return out
}

func add(in <-chan int, additive int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n + additive
        }
    }()
    return out
}
```

## Select Statement

### Что такое select?

`select` позволяет **ожидать несколько операций с каналами** одновременно:

```go
select {
case msg1 := <-ch1:
    fmt.Println("Получено из ch1:", msg1)
case msg2 := <-ch2:
    fmt.Println("Получено из ch2:", msg2)
case ch3 <- msg3:
    fmt.Println("Отправлено в ch3")
default:
    fmt.Println("Ни одна операция не готова")
}
```

### Примеры использования select

#### 1. Таймаут

```go
func timeoutExample() {
    ch := make(chan string)
    
    go func() {
        time.Sleep(2 * time.Second)
        ch <- "Результат"
    }()
    
    select {
    case result := <-ch:
        fmt.Println("Получено:", result)
    case <-time.After(1 * time.Second):
        fmt.Println("Таймаут!")
    }
}
```

#### 2. Отмена операций

```go
func cancellableOperation(ctx context.Context) {
    ch := make(chan string)
    
    go func() {
        // Долгая операция
        time.Sleep(5 * time.Second)
        ch <- "Результат"
    }()
    
    select {
    case result := <-ch:
        fmt.Println("Получено:", result)
    case <-ctx.Done():
        fmt.Println("Операция отменена:", ctx.Err())
    }
}
```

## Лучшие практики

### 1. Закрывайте каналы правильно

```go
// Правило: тот, кто отправляет, закрывает
func producer(ch chan<- int) {
    defer close(ch)
    for i := 0; i < 10; i++ {
        ch <- i
    }
}

func consumer(ch <-chan int) {
    for value := range ch {
        fmt.Println(value)
    }
}
```

### 2. Используйте направления каналов

```go
// Только для отправки
func sender(ch chan<- string) {
    ch <- "сообщение"
}

// Только для получения
func receiver(ch <-chan string) {
    msg := <-ch
    fmt.Println(msg)
}
```

### 3. Избегайте утечек горутин

```go
// Плохо - горутина может жить вечно
func leaky() chan int {
    ch := make(chan int)
    go func() {
        val := <-ch // Может ждать вечно
        fmt.Println(val)
    }()
    return ch
}

// Лучше - с контекстом
func better(ctx context.Context) chan int {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return
        }
    }()
    return ch
}
```

## Производительность

### 1. Буферизированные vs небуферизированные каналы

```go
// Benchmark небуферизированных каналов
func BenchmarkUnbufferedChannel(b *testing.B) {
    ch := make(chan int)
    go func() {
        for i := 0; i < b.N; i++ {
            <-ch
        }
    }()
    
    for i := 0; i < b.N; i++ {
        ch <- i
    }
}

// Benchmark буферизированных каналов
func BenchmarkBufferedChannel(b *testing.B) {
    ch := make(chan int, 100)
    
    go func() {
        for i := 0; i < b.N; i++ {
            <-ch
        }
    }()
    
    for i := 0; i < b.N; i++ {
        ch <- i
    }
}
```

### 2. Размер буфера

Оптимальный размер буфера зависит от:
- **Частоты отправки/получения**
- **Времени обработки**
- **Требований к латентности**

## Распространенные ошибки

### 1. Deadlock

```go
// ОШИБКА: deadlock
func main() {
    ch := make(chan int)
    ch <- 1 // Никто не читает - блокировка
}
```

### 2. Использование закрытых каналов

```go
// ОШИБКА: panic
func main() {
    ch := make(chan int)
    close(ch)
    ch <- 1 // panic: send on closed channel
}
```

### 3. Неправильная синхронизация

```go
// ОШИБКА: гонка данных
var counter int

func main() {
    for i := 0; i < 1000; i++ {
        go func() {
            counter++ // Гонка данных!
        }()
    }
    time.Sleep(time.Second)
    fmt.Println(counter)
}
```

## См. также

- [Горутины для чайников](../concepts/goroutine.md) - базовое объяснение
- [Каналы для чайников](../concepts/channel.md) - базовое объяснение
- [Синхронизация](synchronization.md) - мьютексы и другие примитивы
- [Контекст](../concepts/context.md) - управление жизненным циклом
- [Практические примеры](../examples/goroutines-and-channels) - примеры кода