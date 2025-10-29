# Продвинутые концепции Go

## 1. Горутины и планировщик Go

### Архитектура планировщика Go

Планировщик Go реализует M:N планирование, где M горутин планируются на N потоков операционной системы. Это позволяет эффективно использовать ресурсы системы и обеспечивает высокую производительность при конкурентных операциях.

#### Компоненты планировщика:
- **G (Goroutine)**: Представляет собой горутину с ее стеком, состоянием и контекстом выполнения.
- **M (Machine)**: Представляет собой поток операционной системы.
- **P (Processor)**: Логический процессор, который управляет выполнением горутин.

#### Алгоритм планировщика:
1. Горутины помещаются в очереди выполнения.
2. Логические процессоры (P) берут горутины из очередей и выполняют их на потоках ОС (M).
3. При блокировке горутины переключаются между потоками для эффективного использования ресурсов.

#### Вытесняющая и кооперативная многозадачность:
- **Кооперативная**: Горутина сама отдает управление планировщику (при вызове функций, операциях ввода-вывода).
- **Вытесняющая**: Планировщик периодически прерывает выполнение горутин для переключения контекста.

### Практические аспекты работы с горутинами

#### Управление количеством горутин:
```go
// Ограничение количества горутин с помощью worker pool
func workerPool(tasks <-chan Task, numWorkers int) {
    var wg sync.WaitGroup
    
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range tasks {
                processTask(task)
            }
        }()
    }
    
    wg.Wait()
}
```

#### Обработка ошибок в горутинах:
```go
// Использование каналов для передачи ошибок
func processWithErrors(tasks []Task) error {
    errChan := make(chan error, len(tasks))
    var wg sync.WaitGroup
    
    for _, task := range tasks {
        wg.Add(1)
        go func(t Task) {
            defer wg.Done()
            if err := processTask(t); err != nil {
                errChan <- err
            }
        }(task)
    }
    
    wg.Wait()
    close(errChan)
    
    // Возвращаем первую ошибку
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

## 2. Каналы и коммуникация между горутинами

### Типы каналов и их применение

#### Небуферизированные каналы:
```go
// Синхронная коммуникация
ch := make(chan int)
go func() {
    ch <- 42 // Блокируется до тех пор, пока кто-то не прочитает
}()

value := <-ch // Разблокирует отправителя
```

#### Буферизированные каналы:
```go
// Асинхронная коммуникация с буфером
ch := make(chan int, 3)
ch <- 1 // Не блокируется, если буфер не полон
ch <- 2
ch <- 3
// ch <- 4 // Блокируется, если буфер полон
```

### Паттерны работы с каналами

#### Fan-out/Fan-in:
```go
// Fan-out: распределение задач между воркерами
func fanOut(tasks <-chan Task, numWorkers int) <-chan Result {
    results := make(chan Result)
    
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range tasks {
                results <- processTask(task)
            }
        }()
    }
    
    // Закрываем канал результатов после завершения всех воркеров
    go func() {
        wg.Wait()
        close(results)
    }()
    
    return results
}

// Fan-in: объединение результатов из нескольких источников
func fanIn(channels ...<-chan Result) <-chan Result {
    out := make(chan Result)
    var wg sync.WaitGroup
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan Result) {
            defer wg.Done()
            for result := range c {
                out <- result
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

#### Pipeline:
```go
// Создание конвейера обработки данных
func pipeline(numbers []int) <-chan int {
    // Этап 1: генерация чисел
    source := make(chan int)
    go func() {
        defer close(source)
        for _, n := range numbers {
            source <- n
        }
    }()
    
    // Этап 2: возведение в квадрат
    squared := make(chan int)
    go func() {
        defer close(squared)
        for n := range source {
            squared <- n * n
        }
    }()
    
    // Этап 3: фильтрация четных чисел
    filtered := make(chan int)
    go func() {
        defer close(filtered)
        for n := range squared {
            if n%2 == 0 {
                filtered <- n
            }
        }
    }()
    
    return filtered
}
```

### Select и его применение

#### Таймауты:
```go
func withTimeout(operation func() error, timeout time.Duration) error {
    result := make(chan error, 1)
    
    go func() {
        result <- operation()
    }()
    
    select {
    case err := <-result:
        return err
    case <-time.After(timeout):
        return errors.New("operation timed out")
    }
}
```

#### Неблокирующая коммуникация:
```go
func nonBlockingSend(ch chan<- int, value int) bool {
    select {
    case ch <- value:
        return true
    default:
        return false // Канал занят, отправка не удалась
    }
}
```

## 3. Мьютексы и примитивы синхронизации

### Расширенное использование sync.Mutex

#### Правильное использование defer:
```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock() // Гарантирует разблокировку при любом выходе из функции
    c.value++
}

func (c *Counter) Reset() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    // Даже при panic мьютекс будет освобожден
    c.value = 0
}
```

#### RWMutex для оптимизации чтения:
```go
type Config struct {
    mu     sync.RWMutex
    values map[string]string
}

func (c *Config) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    value, exists := c.values[key]
    return value, exists
}

func (c *Config) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.values[key] = value
}
```

### sync.Once для ленивой инициализации

#### Потокобезопасная инициализация:
```go
type Database struct {
    pool *sql.DB
    once sync.Once
}

func (db *Database) Connect() error {
    var err error
    db.once.Do(func() {
        // Этот код выполнится только один раз, даже при конкурентном доступе
        db.pool, err = sql.Open("postgres", "connection_string")
        if err != nil {
            return
        }
        
        // Дополнительная инициализация
        db.pool.SetMaxOpenConns(100)
        db.pool.SetMaxIdleConns(10)
    })
    
    return err
}
```

### sync.WaitGroup для ожидания завершения

#### Расширенное использование:
```go
func processBatch(items []Item) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(items))
    
    // Ограничиваем количество одновременных горутин
    semaphore := make(chan struct{}, 10) // максимум 10 горутин
    
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            
            semaphore <- struct{}{} // Захватываем слот
            defer func() { <-semaphore }() // Освобождаем слот
            
            if err := processItem(i); err != nil {
                errChan <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errChan)
    
    // Проверяем ошибки
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

### sync.Pool для переиспользования объектов

#### Эффективное использование памяти:
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024) // Буфер размером 1KB
    },
}

func processRequest(data []byte) {
    // Получаем буфер из пула
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf) // Возвращаем буфер в пул
    
    // Используем буфер для обработки данных
    // ... обработка данных ...
}
```

## 4. Контексты и управление жизненным циклом операций

### Расширенное использование контекста

#### Создание контекстов с различными параметрами:
```go
// Контекст с таймаутом
func withTimeoutContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    return context.WithTimeout(parent, timeout)
}

// Контекст с дедлайном
func withDeadlineContext(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
    return context.WithDeadline(parent, deadline)
}

// Контекст с отменой
func withCancelContext(parent context.Context) (context.Context, context.CancelFunc) {
    return context.WithCancel(parent)
}

// Контекст со значением
func withValueContext(parent context.Context, key, value interface{}) context.Context {
    return context.WithValue(parent, key, value)
}
```

#### Правильная передача контекста:
```go
// Хорошо: контекст передается как первый параметр
func ProcessRequest(ctx context.Context, request *Request) (*Response, error) {
    // Проверяем отмену контекста
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Передаем контекст дальше
    result, err := database.Query(ctx, request.Query)
    if err != nil {
        return nil, err
    }
    
    return &Response{Data: result}, nil
}
```

#### Использование контекста для отмены операций:
```go
func longRunningOperation(ctx context.Context) error {
    // Создаем канал для сигнала завершения
    done := make(chan error, 1)
    
    go func() {
        // Долгая операция
        err := performLongOperation()
        done <- err
    }()
    
    select {
    case err := <-done:
        return err
    case <-ctx.Done():
        // Контекст отменен, операция должна быть прервана
        return ctx.Err()
    }
}
```

## 5. Интерфейсы и неявная реализация

### Продвинутые аспекты интерфейсов

#### Композиция интерфейсов:
```go
// Базовые интерфейсы
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// Композиция интерфейсов
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

#### Интерфейсы с поведением:
```go
// Интерфейс для объектов, которые могут быть сброшены
type Resetter interface {
    Reset()
}

// Интерфейс для объектов, которые могут сообщить свой размер
typeSizer interface {
    Size() int64
}

// Интерфейс для объектов, которые могут быть проверены на пустоту
type EmptyChecker interface {
    IsEmpty() bool
}
```

#### Type assertion и type switch:
```go
func processValue(v interface{}) {
    switch val := v.(type) {
    case string:
        fmt.Printf("String: %s\n", val)
    case int:
        fmt.Printf("Integer: %d\n", val)
    case fmt.Stringer:
        fmt.Printf("Stringer: %s\n", val.String())
    default:
        fmt.Printf("Unknown type: %T\n", val)
    }
}

// Проверка с возвратом ошибки
func getString(v interface{}) (string, bool) {
    if str, ok := v.(string); ok {
        return str, true
    }
    return "", false
}
```

## 6. Работа с памятью и сборщик мусора

### Оптимизация использования памяти

#### Минимизация аллокаций:
```go
// Плохо: создание новых строк при конкатенации
func badConcat(strs []string) string {
    var result string
    for _, s := range strs {
        result += s // Каждая операция создает новую строку
    }
    return result
}

// Хорошо: использование strings.Builder
func goodConcat(strs []string) string {
    var builder strings.Builder
    for _, s := range strs {
        builder.WriteString(s)
    }
    return builder.String()
}
```

#### Использование sync.Pool:
```go
var userPool = sync.Pool{
    New: func() interface{} {
        return &User{}
    },
}

func processUsers(users []UserData) {
    for _, data := range users {
        // Получаем объект из пула
        user := userPool.Get().(*User)
        
        // Инициализируем объект
        user.Name = data.Name
        user.Email = data.Email
        
        // Используем объект
        saveUser(user)
        
        // Сбрасываем состояние и возвращаем в пул
        user.Reset()
        userPool.Put(user)
    }
}
```

### Профилирование памяти

#### Анализ аллокаций:
```bash
# Запуск бенчмарка с профилированием памяти
go test -bench=. -memprofile=mem.prof

# Анализ профиля
go tool pprof mem.prof

# В интерактивном режиме pprof:
# top - показать топ аллокаций
# list functionName - показать аллокации в функции
# web - визуализация в браузере
```

#### Оптимизация структур данных:
```go
// Плохо: неоптимальное расположение полей
type BadStruct struct {
    b   bool    // 1 байт + 7 байт padding
    i64 int64   // 8 байт
    i32 int32   // 4 байта + 4 байта padding
}

// Хорошо: оптимальное расположение полей
type GoodStruct struct {
    i64 int64   // 8 байт
    i32 int32   // 4 байта
    b   bool    // 1 байт + 3 байта padding
}
```

### Настройка сборщика мусора

#### Параметры GC:
```go
// Управление частотой GC через GOGC переменную окружения
// GOGC=100 - стандартное значение (сборка при удвоении аллокаций)
// GOGC=50 - более частая сборка (при увеличении на 50%)
// GOGC=200 - менее частая сборка (при увеличении на 200%)
// GOGC=off - отключение GC

// Программное управление GC
import "runtime"

func configureGC() {
    // Установка цели GC
    debug.SetGCPercent(50)
    
    // Принудительный запуск GC
    runtime.GC()
    
    // Получение статистики GC
    var stats debug.GCStats
    debug.ReadGCStats(&stats)
    fmt.Printf("Последняя сборка: %v\n", stats.LastGC)
    fmt.Printf("Количество сборок: %d\n", stats.NumGC)
}
```

## Заключение

Продвинутые концепции Go требуют глубокого понимания внутреннего устройства языка и его runtime. Освоение этих концепций позволяет создавать эффективные, масштабируемые и надежные приложения. Ключевые аспекты:

1. **Горутины и планировщик** - понимание конкурентности в Go
2. **Каналы** - правильная коммуникация между горутинами
3. **Синхронизация** - безопасное использование общих ресурсов
4. **Контексты** - управление жизненным циклом операций
5. **Интерфейсы** - гибкое проектирование API
6. **Память** - оптимизация использования ресурсов

Эти концепции тесно связаны между собой и требуют комплексного подхода к изучению и применению на практике.