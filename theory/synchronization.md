# Синхронизация в Go: Полная теория

## Введение в синхронизацию

### Что такое синхронизация?

Синхронизация - это **координация доступа к общим ресурсам** несколькими горутинами. Без синхронизации возникают **гонки данных** (data races).

### Гонки данных

Гонка данных происходит когда:
1. **Две или более горутины** обращаются к одной переменной
2. **Хотя бы одна** выполняет запись
3. **Нет синхронизации**

```go
// Пример гонки данных
var counter int

func main() {
    for i := 0; i < 1000; i++ {
        go func() {
            counter++ // Гонка данных!
        }()
    }
    
    time.Sleep(time.Second)
    fmt.Println(counter) // Результат будет разным каждый раз
}
```

## Примитивы синхронизации

### 1. sync.Mutex (Мьютекс)

#### Основы мьютексов

Мьютекс обеспечивает **взаимное исключение** - только одна горутина может владеть мьютексом в любой момент.

```go
var (
    counter int
    mutex   sync.Mutex
)

func increment() {
    mutex.Lock()
    counter++
    mutex.Unlock()
}

func main() {
    var wg sync.WaitGroup
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            increment()
        }()
    }
    
    wg.Wait()
    fmt.Println("Счетчик:", counter) // Всегда 1000
}
```

#### Лучшие практики с мьютексами

```go
// Используйте defer для автоматического освобождения
func safeOperation() {
    mutex.Lock()
    defer mutex.Unlock()
    
    // Работа с защищенными данными
    // Мьютекс автоматически освободится при выходе
}

// Избегайте длительных блокировок
func betterOperation() {
    // Получаем данные под блокировкой
    mutex.Lock()
    data := protectedData
    mutex.Unlock()
    
    // Долгая операция без блокировки
    time.Sleep(time.Second)
    
    // Сохраняем результат под блокировкой
    mutex.Lock()
    protectedData = newData
    mutex.Unlock()
}
```

### 2. sync.RWMutex (Читатель-писатель мьютекс)

RWMutex позволяет **множественным читателям** или **одному писателю** одновременно.

```go
var (
    data    map[string]string
    rwMutex sync.RWMutex
)

func ReadData(key string) (string, bool) {
    rwMutex.RLock()
    defer rwMutex.RUnlock()
    
    value, exists := data[key]
    return value, exists
}

func WriteData(key, value string) {
    rwMutex.Lock()
    defer rwMutex.Unlock()
    
    data[key] = value
}
```

#### Когда использовать RWMutex?

Используйте RWMutex когда:
- **Чтений больше, чем записей**
- **Долгие операции чтения**
- **Нужна высокая производительность чтения**

### 3. sync.Once

Once гарантирует, что **функция выполнится только один раз**.

```go
var (
    config *Config
    once   sync.Once
)

func GetConfig() *Config {
    once.Do(func() {
        // Эта функция выполнится только один раз
        config = loadConfig()
    })
    return config
}
```

#### Практическое применение Once

```go
type Database struct {
    pool *sql.DB
    once sync.Once
}

func (db *Database) Connect() error {
    var err error
    db.once.Do(func() {
        db.pool, err = sql.Open("postgres", "connection_string")
    })
    return err
}
```

### 4. sync.WaitGroup

WaitGroup ожидает завершения **группы горутин**.

```go
func main() {
    var wg sync.WaitGroup
    
    for i := 0; i < 10; i++ {
        wg.Add(1) // Увеличиваем счетчик
        go func(id int) {
            defer wg.Done() // Уменьшаем счетчик при завершении
            fmt.Printf("Горутина %d завершена\n", id)
        }(i)
    }
    
    wg.Wait() // Ждем, пока счетчик станет 0
    fmt.Println("Все горутины завершены")
}
```

#### Расширенное использование WaitGroup

```go
func processFiles(files []string) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(f string) {
            defer wg.Done()
            if err := processFile(f); err != nil {
                errChan <- err
            }
        }(file)
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

### 5. sync.Cond

Cond реализует **условную переменную** для синхронизации.

```go
var (
    queue     []int
    mutex     sync.Mutex
    condition = sync.NewCond(&mutex)
)

func producer() {
    for i := 0; i < 10; i++ {
        mutex.Lock()
        queue = append(queue, i)
        mutex.Unlock()
        
        condition.Signal() // Уведомляем одного потребителя
        time.Sleep(time.Millisecond * 100)
    }
}

func consumer(id int) {
    for {
        mutex.Lock()
        for len(queue) == 0 {
            condition.Wait() // Ждем, пока очередь не заполнится
        }
        
        item := queue[0]
        queue = queue[1:]
        mutex.Unlock()
        
        fmt.Printf("Потребитель %d получил: %d\n", id, item)
        time.Sleep(time.Millisecond * 200)
    }
}
```

## Атомарные операции

### Что такое atomic операции?

Atomic операции выполняются **неделимо** - они либо полностью завершаются, либо не начинаются.

### Пакет sync/atomic

```go
import "sync/atomic"

var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

func getValue() int64 {
    return atomic.LoadInt64(&counter)
}

func compareAndSwap(old, new int64) bool {
    return atomic.CompareAndSwapInt64(&counter, old, new)
}
```

### Когда использовать atomic?

Используйте atomic когда:
- **Простые операции** с целыми числами
- **Высокая производительность** критична
- **Нет сложной логики** синхронизации

```go
// Atomic vs Mutex для счетчика
var (
    atomicCounter int64
    mutexCounter  int
    mutex         sync.Mutex
)

// Быстрее
func atomicIncrement() {
    atomic.AddInt64(&atomicCounter, 1)
}

// Медленнее
func mutexIncrement() {
    mutex.Lock()
    mutexCounter++
    mutex.Unlock()
}
```

## Сравнение подходов синхронизации

### Mutex vs Channels

| Характеристика | Mutex | Channels |
|----------------|-------|----------|
| Назначение | Защита состояния | Коммуникация |
| Сложность | Ниже | Выше |
| Производительность | Выше для простых случаев | Ниже для простых случаев |
| Использование | Общие переменные | Передача данных |

```go
// Mutex для защиты состояния
type Counter struct {
    value int
    mutex sync.Mutex
}

func (c *Counter) Increment() {
    c.mutex.Lock()
    c.value++
    c.mutex.Unlock()
}

// Channels для коммуникации
func worker(jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        result := process(job)
        results <- result
    }
}
```

### Atomic vs Mutex

| Характеристика | Atomic | Mutex |
|----------------|--------|-------|
| Скорость | Быстрее | Медленнее |
| Гибкость | Ограниченная | Полная |
| Сложность | Простая | Сложнее |

```go
// Для простых счетчиков - atomic
var requests int64

func handleRequest() {
    atomic.AddInt64(&requests, 1)
}

// Для сложных операций - mutex
type BankAccount struct {
    balance int
    mutex   sync.Mutex
}

func (ba *BankAccount) Transfer(amount int, to *BankAccount) error {
    ba.mutex.Lock()
    defer ba.mutex.Unlock()
    
    if ba.balance < amount {
        return errors.New("недостаточно средств")
    }
    
    ba.balance -= amount
    to.mutex.Lock()
    to.balance += amount
    to.mutex.Unlock()
    
    return nil
}
```

## Практические паттерны

### 1. Thread-Safe коллекции

```go
type SafeMap struct {
    data map[string]interface{}
    mutex sync.RWMutex
}

func (sm *SafeMap) Set(key string, value interface{}) {
    sm.mutex.Lock()
    defer sm.mutex.Unlock()
    sm.data[key] = value
}

func (sm *SafeMap) Get(key string) (interface{}, bool) {
    sm.mutex.RLock()
    defer sm.mutex.RUnlock()
    value, exists := sm.data[key]
    return value, exists
}
```

### 2. Singleton с ленивой инициализацией

```go
type Database struct {
    pool *sql.DB
}

var (
    instance *Database
    once     sync.Once
)

func GetInstance() *Database {
    once.Do(func() {
        instance = &Database{
            pool: createConnectionPool(),
        }
    })
    return instance
}
```

### 3. Pool объектов

```go
type Worker struct {
    id int
    // другие поля
}

var workerPool = sync.Pool{
    New: func() interface{} {
        return &Worker{
            id: generateID(),
        }
    },
}

func processTask() {
    worker := workerPool.Get().(*Worker)
    defer workerPool.Put(worker)
    
    // Используем worker
    doWork(worker)
}
```

## Профилирование синхронизации

### Обнаружение гонок данных

```bash
# Запуск с детектором гонок
go run -race main.go

# Тестирование с детектором гонок
go test -race
```

### Профилирование блокировок

```bash
# Включение профилирования блокировок
import _ "net/http/pprof"

func main() {
    // Настройка блок-профилирования
    runtime.SetBlockProfileRate(1)
    
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
    
    // Приложение
}
```

Анализ:
```bash
# Сбор профиля блокировок
go tool pprof http://localhost:6060/debug/pprof/block
```

## Распространенные ошибки

### 1. Забытый Unlock

```go
// ОШИБКА
func badFunction() {
    mutex.Lock()
    if someCondition {
        return // Забыли Unlock!
    }
    mutex.Unlock()
}

// ИСПРАВЛЕНО
func goodFunction() {
    mutex.Lock()
    defer mutex.Unlock()
    
    if someCondition {
        return // defer гарантирует Unlock
    }
}
```

### 2. Взаимная блокировка (Deadlock)

```go
// ОШИБКА - deadlock
var mutex1, mutex2 sync.Mutex

func function1() {
    mutex1.Lock()
    mutex2.Lock() // Может ждать function2
    mutex2.Unlock()
    mutex1.Unlock()
}

func function2() {
    mutex2.Lock()
    mutex1.Lock() // Может ждать function1
    mutex1.Unlock()
    mutex2.Unlock()
}
```

### 3. Копирование синхронизирующих примитивов

```go
// ОШИБКА
func badFunction(mutex sync.Mutex) {
    // Копирование мьютекса
}

// ИСПРАВЛЕНО
func goodFunction(mutex *sync.Mutex) {
    // Передаем указатель
}
```

## Лучшие практики

### 1. Минимизируйте время блокировки

```go
// ПЛОХО
func badOperation() {
    mutex.Lock()
    defer mutex.Unlock()
    
    // Долгая операция под блокировкой
    time.Sleep(time.Second)
}

// ЛУЧШЕ
func goodOperation() {
    // Получаем данные под блокировкой
    mutex.Lock()
    data := protectedData
    mutex.Unlock()
    
    // Долгая операция без блокировки
    processedData := processData(data)
    
    // Сохраняем результат под блокировкой
    mutex.Lock()
    protectedData = processedData
    mutex.Unlock()
}
```

### 2. Используйте defer для освобождения

```go
// Всегда используйте defer
mutex.Lock()
defer mutex.Unlock()

// Работа с защищенными данными
```

### 3. Тестируйте с -race

```bash
# Регулярно тестируйте на гонки
go test -race ./...
```

## См. также

- [Мьютексы для чайников](../concepts/mutex.md) - базовое объяснение
- [Каналы](../concepts/channel.md) - альтернативный способ синхронизации
- [Горутины](../concepts/goroutine.md) - что синхронизировать
- [Профилирование](../concepts/profiling.md) - как находить проблемы синхронизации
- [Практические примеры](../examples/synchronization) - примеры кода