# Управление памятью в Go: Полная теория

## Введение в управление памятью

### Что такое управление памятью?

Управление памятью в Go - это **автоматический процесс**, который:
- **Выделяет** память для объектов
- **Отслеживает** использование памяти
- **Освобождает** неиспользуемую память

### Преимущества автоматического управления памятью

1. **Безопасность** - нет утечек памяти из-за забытых free()
2. **Простота** - разработчик не управляет памятью вручную
3. **Производительность** - оптимизированный сборщик мусора

## Стек vs Куча

### Стек (Stack)

**Стек** - это **быстрая память**, связанная с **горутиной**:

```go
func calculateSum(a, b int) int {
    result := a + b // result живет на стеке
    return result   // result автоматически уничтожается
} // Здесь result исчезает со стека
```

#### Характеристики стека:

- **Быстрый доступ** - прямая адресация
- **Автоматическое освобождение** - при выходе из функции
- **Ограниченный размер** - обычно 1-2MB на горутину
- **Локальная видимость** - только в текущей функции

### Куча (Heap)

**Куча** - это **динамическая память**, управляемая **сборщиком мусора**:

```go
func createPointer() *int {
    x := 42        // x должен "выжить" после функции
    return &x      // x перемещается в кучу (escape analysis)
} // x остается в куче, так как на него есть ссылка
```

#### Характеристики кучи:

- **Медленный доступ** - косвенная адресация
- **Ручное освобождение** - сборщик мусора
- **Большой размер** - ограничение только ОС
- **Глобальная видимость** - доступна всем

## Escape Analysis (Анализ убегания)

### Что такое escape analysis?

Escape analysis - это **процесс компилятора**, который определяет, где должна храниться переменная - в стеке или в куче.

### Как работает escape analysis?

Компилятор анализирует:
- **Используются ли адреса переменных?**
- **Возвращаются ли указатели на переменные?**
- **Передаются ли переменные в замыкания?**

### Примеры escape analysis:

```go
// Переменная в стеке
func stackAllocation() int {
    x := 42
    return x // x не "escape", остается в стеке
}

// Переменная в куче
func heapAllocation() *int {
    x := 42
    return &x // x "escape" в кучу, так как возвращается указатель
}

// Переменная в куче из-за замыкания
func closureExample() func() int {
    x := 42
    return func() int {
        return x // x "escape" в кучу, так как используется в замыкании
    }
}

// Переменная в куче из-за интерфейса
func interfaceExample() interface{} {
    x := 42
    return x // x "escape" в кучу, так как возвращается как interface{}
}
```

### Как проверить escape analysis?

```bash
# Компиляция с анализом убегания
go build -gcflags="-m" main.go

# Более подробный анализ
go build -gcflags="-m -m" main.go
```

Пример вывода:
```
./main.go:2:2: moved to heap: x
./main.go:3:9: &x escapes to heap
./main.go:7:2: result does not escape
```

## Указатели

### Что такое указатель?

Указатель - это **адрес в памяти**, где хранится значение:

```go
func main() {
    x := 42        // x хранит значение 42
    p := &x        // p хранит адрес x (указатель на x)
    *p = 21        // меняем значение по адресу p
    fmt.Println(x) // выводит 21
}
```

### Визуализация указателей:

```
Память:
Адрес    Значение
0x1000   42       <- x хранится здесь
0x1008   0x1000   <- p хранит адрес x

p -> 0x1000 -> 42
```

### Когда использовать указатели?

#### 1. Изменение значений в функциях

```go
// Без указателя - значение копируется
func incrementBad(x int) {
    x++ // Изменяем копию
}

// С указателем - изменяем оригинальное значение
func incrementGood(x *int) {
    *x++ // Изменяем оригинальное значение
}

func main() {
    value := 10
    incrementGood(&value)
    fmt.Println(value) // Вывод: 11
}
```

#### 2. Избежание копирования больших структур

```go
type BigStruct struct {
    data [1000000]int
}

// ПЛОХО - копируем всю структуру
func processBad(s BigStruct) {
    // ...
}

// ЛУЧШЕ - передаем указатель
func processGood(s *BigStruct) {
    // ...
}
```

#### 3. Представление необязательных значений

```go
type User struct {
    Name     string
    Email    string
    Phone    *string // Необязательное поле
}

func main() {
    user := User{
        Name:  "Иван",
        Email: "ivan@example.com",
        // Phone: nil - отсутствует
    }
    
    // Проверяем наличие телефона
    if user.Phone != nil {
        fmt.Println("Телефон:", *user.Phone)
    } else {
        fmt.Println("Телефон не указан")
    }
}
```

## Сборщик мусора (Garbage Collector)

### Как работает GC в Go?

Go использует **три-color mark-and-sweep** алгоритм:

1. **Белый** - объекты, которые еще не проверены
2. **Серый** - объекты, которые обнаружены, но не полностью проверены
3. **Черный** - объекты, которые полностью проверены и используются

### Фазы сборки мусора:

1. **Mark Setup** - подготовка к маркировке
2. **Mark** - маркировка достижимых объектов
3. **Mark Termination** - завершение маркировки
4. **Sweep** - освобождение недостижимых объектов

### Производительность GC

#### Параметры GC:

```go
// Установка цели GC
import "runtime/debug"

func main() {
    // Цель GC: 50% времени на GC, 50% на работу программы
    debug.SetGCPercent(100)
    
    // Отключение GC
    debug.SetGCPercent(-1)
    
    // Установка цели GC: 20% времени на GC
    debug.SetGCPercent(20)
}
```

#### Мониторинг GC:

```go
import "runtime"

func printGCStats() {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    fmt.Printf("Alloc = %d KB", stats.Alloc/1024)
    fmt.Printf(", TotalAlloc = %d KB", stats.TotalAlloc/1024)
    fmt.Printf(", Sys = %d KB", stats.Sys/1024)
    fmt.Printf(", NumGC = %d\n", stats.NumGC)
}
```

## Профилирование памяти

### Как профилировать память?

```bash
# Запуск с профилированием памяти
go test -memprofile=mem.prof

# Анализ профиля
go tool pprof mem.prof
```

### Интерактивный анализ:

```bash
# Интерактивный режим
go tool pprof mem.prof

(pprof) top          # Топ функций по аллокациям
(pprof) inuse_space  # Используемая память
(pprof) alloc_space  # Всего аллоцировано
(pprof) web          # Визуализация в браузере
```

### Пример вывода:

```
(pprof) top
Showing nodes accounting for 400MB, 100% of 400MB total
      flat  flat%   sum%        cum   cum%
     200MB 50.00% 50.00%      200MB 50.00%  main.createStrings
     100MB 25.00% 75.00%      100MB 25.00%  runtime.mallocgc
     100MB 25.00% 100.00%      400MB   100%  main.main
```

## Оптимизация использования памяти

### 1. Минимизация аллокаций

```go
// ПЛОХО - много аллокаций
func concatenateStrings(strs []string) string {
    result := ""
    for _, s := range strs {
        result += s // Каждая конкатенация создает новую строку
    }
    return result
}

// ЛУЧШЕ - предварительная аллокация
func concatenateStringsBetter(strs []string) string {
    // Вычисляем общую длину
    totalLen := 0
    for _, s := range strs {
        totalLen += len(s)
    }
    
    // Предварительно аллоцируем буфер
    result := make([]byte, 0, totalLen)
    for _, s := range strs {
        result = append(result, s...)
    }
    
    return string(result)
}

// ЕЩЕ ЛУЧШЕ - strings.Builder
func concatenateStringsBest(strs []string) string {
    var builder strings.Builder
    
    // Предварительно устанавливаем размер
    totalLen := 0
    for _, s := range strs {
        totalLen += len(s)
    }
    builder.Grow(totalLen)
    
    for _, s := range strs {
        builder.WriteString(s)
    }
    
    return builder.String()
}
```

### 2. Использование пулов объектов

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processRequest(data []byte) {
    // Получаем буфер из пула
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        // Очищаем и возвращаем в пул
        buf.Reset()
        bufferPool.Put(buf)
    }()
    
    buf.Write(data)
    // Обработка данных
}
```

### 3. Избежание утечек горутин

```go
// ПЛОХО - горутина может жить вечно
func leakyFunction() {
    ch := make(chan int)
    go func() {
        val := <-ch // Может ждать вечно
        fmt.Println(val)
    }()
    // ch никогда не закрывается!
}

// ЛУЧШЕ - используем context для отмены
func betterFunction(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return // Завершаем горутину при отмене
        }
    }()
}
```

### 4. Оптимизация структур данных

```go
// ПЛОХО - неоптимальное выравнивание
type BadStruct struct {
    b   bool    // 1 byte
    i64 int64   // 8 bytes (выравнивание добавит 7 байт)
    i32 int32   // 4 bytes (выравнивание добавит 4 байта)
}

// ЛУЧШЕ - оптимальное выравнивание
type GoodStruct struct {
    i64 int64   // 8 bytes
    i32 int32   // 4 bytes
    b   bool    // 1 byte (выравнивание добавит 3 байта)
}
```

## Распространенные ошибки

### 1. Утечки памяти

```go
// Утечка через горутины
func memoryLeak() {
    for {
        go func() {
            // Долгая операция без возможности отмены
            time.Sleep(time.Hour)
        }()
    }
}

// Утечка через ссылки
func referenceLeak() {
    data := make([]byte, 1024*1024) // 1MB
    
    // Храним ссылку на маленький срез большого массива
    smallSlice := data[0:10]
    
    // data не может быть собран, пока существует smallSlice
    processSmallSlice(smallSlice)
}
```

### 2. Избыточные аллокации

```go
// ПЛОХО
func badFunction() []string {
    result := []string{}
    for i := 0; i < 1000000; i++ {
        result = append(result, fmt.Sprintf("item%d", i)) // Много аллокаций
    }
    return result
}

// ЛУЧШЕ
func goodFunction() []string {
    result := make([]string, 1000000) // Одна аллокация
    for i := 0; i < 1000000; i++ {
        result[i] = fmt.Sprintf("item%d", i)
    }
    return result
}
```

### 3. Игнорирование escape analysis

```go
// ПЛОХО - ненужное выделение в куче
func badAllocation() *int {
    x := 42
    return &x // x выделяется в куче
}

// ЛУЧШЕ - возврат значения
func goodAllocation() int {
    x := 42
    return x // x остается в стеке
}
```

## Лучшие практики

### 1. Используйте профилирование

```bash
# Регулярное профилирование
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

### 2. Мониторинг в production

```go
func monitorMemory() {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)
    
    // Логируем статистику
    log.Printf("Alloc = %d KB, Sys = %d KB, NumGC = %d",
        stats.Alloc/1024, stats.Sys/1024, stats.NumGC)
}
```

### 3. Используйте инструменты анализа

```bash
# Анализ убегания
go build -gcflags="-m" main.go

# Race detector
go test -race

# Benchmarks
go test -bench=.
```

## Расширенные примеры

### 1. Оптимизация структур

```go
// Оптимизация через теги
type User struct {
    ID       int64  `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Active   bool   `json:"active"`
    Created  int64  `json:"created"`
    Modified int64  `json:"modified"`
}

// Использование unsafe для оптимизации (осторожно!)
import "unsafe"

func getSize(user *User) int {
    return int(unsafe.Sizeof(*user))
}
```

### 2. Пул соединений

```go
type ConnectionPool struct {
    pool chan *Connection
    factory func() (*Connection, error)
}

func NewConnectionPool(size int, factory func() (*Connection, error)) *ConnectionPool {
    pool := make(chan *Connection, size)
    
    for i := 0; i < size; i++ {
        conn, err := factory()
        if err != nil {
            panic(err)
        }
        pool <- conn
    }
    
    return &ConnectionPool{
        pool:    pool,
        factory: factory,
    }
}

func (cp *ConnectionPool) Get() (*Connection, error) {
    select {
    case conn := <-cp.pool:
        return conn, nil
    default:
        return cp.factory()
    }
}

func (cp *ConnectionPool) Put(conn *Connection) {
    select {
    case cp.pool <- conn:
    default:
        // Пул полон, закрываем соединение
        conn.Close()
    }
}
```

## См. также

- [Работа с памятью для чайников](../concepts/memory.md) - базовое объяснение
- [Профилирование](../concepts/profiling.md) - как находить проблемы с памятью
- [Горутины](../concepts/goroutine.md) - как они используют память
- [Тестирование производительности](benchmarking.md) - как измерять использование памяти
- [Практические примеры](../examples/memory) - примеры оптимизации