# Работа с памятью в Go - объяснение для чайников

## Что такое стек и куча?

Представьте, что память в программе - это **огромный склад**. У каждого есть свой **рабочий стол** (стек) и доступ к **общему складу** (куча).

### Стек (Stack)

**Стек** - это как ваш **рабочий стол**:
- У каждого есть свой стол
- Места немного, но работать удобно
- Когда функция заканчивает работу, весь "мусор" с нее автоматически убирается
- Очень быстрый доступ к данным

```go
func calculateSum(a, b int) int {
    result := a + b // result живет на стеке
    return result   // result автоматически уничтожается
} // Здесь result исчезает со стека
```

### Куча (Heap)

**Куча** - это как **общий склад**:
- Много места
- Медленнее доступ
- Нужно вручную убирать "мусор" (сборщик мусора в Go делает это автоматически)
- Данные могут жить дольше, чем функция, которая их создала

```go
func createPointer() *int {
    x := 42        // x должен "выжить" после функции
    return &x      // x перемещается в кучу (escape analysis)
} // x остается в куче, так как на него есть ссылка
```

## Указатели

### Что такое указатель?

Указатель - это **адрес в памяти**, где хранится значение. Представьте, что это **номер ячейки в шкафу**.

```go
func main() {
    x := 42        // x хранит значение 42
    p := &x        // p хранит адрес x (указатель на x)
    *p = 21        // меняем значение по адресу p
    fmt.Println(x) // выводит 21
}
```

### Визуализация:

```
Память:
Адрес    Значение
0x1000   42       <- x хранится здесь
0x1008   0x1000   <- p хранит адрес x

p -> 0x1000 -> 42
```

### Когда использовать указатели?

1. **Изменение значений в функциях**:
```go
func increment(x *int) {
    *x++ // Изменяем оригинальное значение
}

func main() {
    value := 10
    increment(&value)
    fmt.Println(value) // Вывод: 11
}
```

2. **Избежание копирования больших структур**:
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

## Escape Analysis (Анализ убегания)

Go компилятор автоматически определяет, где должна храниться переменная - в стеке или в куче.

### Примеры:

```go
// Переменная в стеке
func stackExample() int {
    x := 42
    return x // x не "escape", остается в стеке
}

// Переменная в куче
func heapExample() *int {
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
```

### Как проверить escape analysis?

```bash
go build -gcflags="-m" main.go
```

Вывод может быть таким:
```
./main.go:2:2: moved to heap: x
./main.go:3:9: &x escapes to heap
```

## Сборщик мусора (Garbage Collector)

### Что делает GC?

Сборщик мусора - это **автоматический уборщик**, который удаляет неиспользуемые объекты из кучи.

### Как работает GC в Go?

Go использует **три-color mark-and-sweep** алгоритм:
1. **Белый** - объекты, которые еще не проверены
2. **Серый** - объекты, которые обнаружены, но не полностью проверены
3. **Черный** - объекты, которые полностью проверены и используются

### Производительность GC

```go
// Плохой код - много аллокаций
func badFunction() []string {
    result := []string{}
    for i := 0; i < 1000000; i++ {
        result = append(result, fmt.Sprintf("item%d", i)) // Много аллокаций
    }
    return result
}

// Лучший код - предварительная аллокация
func goodFunction() []string {
    result := make([]string, 1000000) // Одна аллокация
    for i := 0; i < 1000000; i++ {
        result[i] = fmt.Sprintf("item%d", i)
    }
    return result
}
```

## Профилирование памяти

### Как профилировать?

```bash
# Запуск с профилированием памяти
go test -memprofile=mem.prof

# Анализ профиля
go tool pprof mem.prof
```

### Что искать в профиле?

1. **Высокие аллокации** - функции, которые создают много объектов
2. **Утечки памяти** - объекты, которые не удаляются
3. **Долгоживущие объекты** - объекты, которые живут слишком долго

## Практические советы

### 1. Минимизируйте аллокации

```go
// ПЛОХО
func concatenateStrings(strs []string) string {
    result := ""
    for _, s := range strs {
        result += s // Каждая конкатенация создает новую строку
    }
    return result
}

// ЛУЧШЕ
func concatenateStringsBetter(strs []string) string {
    var builder strings.Builder
    for _, s := range strs {
        builder.WriteString(s)
    }
    return builder.String()
}
```

### 2. Используйте пулы объектов

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processRequest(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    
    buf.Reset()
    buf.Write(data)
    // Обработка данных
}
```

### 3. Избегайте утечек горутин

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

## Инструменты для анализа

### 1. pprof

```go
import _ "net/http/pprof"

func main() {
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
    // Теперь можно открыть http://localhost:6060/debug/pprof/
}
```

### 2. runtime/pprof

```go
import "runtime/pprof"

func writeMemProfile(filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()
    
    return pprof.WriteHeapProfile(f)
}
```

### 3. go test с профилированием

```bash
# Профиль памяти
go test -memprofile=mem.prof

# Профиль CPU
go test -cpuprofile=cpu.prof

# Гонки данных
go test -race
```

## См. также

- [Горутины](goroutine.md) - как они используют память
- [Профилирование](../theory/profiling.md) - как находить проблемы с памятью
- [Тестирование производительности](../theory/benchmarking.md) - как измерять использование памяти