# День 1-2: Железное ядро Go (24 часа)

## Утро

### Горутины и каналы

#### Что такое горутина?
Горутина — это легковесный поток, управляемый рантаймом Go.

Горутины легковеснее потоков, планируются рантаймом Go. Они стартуют с маленьким стеком (~2KB), который может расти по мере необходимости.

#### Как создать горутину?
```go
go func() {
    // код, который будет выполняться в горутине
}()
```

#### Каналы
Каналы используются для коммуникации между горутинами.

##### Небуферизированные каналы
```go
ch := make(chan int) // небуферизированный канал
```

Небуферизированные каналы блокируют отправителя до тех пор, пока кто-то не готов принять значение.

##### Буферизированные каналы
```go
ch := make(chan int, 3) // буферизированный канал с емкостью 3
```

Буферизированные каналы могут хранить определенное количество значений без блокировки отправителя.

#### Пример использования горутин и каналов
```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch := make(chan string)
    
    // Запускаем горутину
    go func() {
        time.Sleep(1 * time.Second)
        ch <- "Привет из горутины!"
    }()
    
    // Получаем значение из канала
    msg := <-ch
    fmt.Println(msg)
}
```

### Мьютексы

#### sync.Mutex
Мьютекс используется для защиты общих ресурсов от одновременного доступа нескольких горутин.

```go
package main

import (
    "fmt"
    "sync"
)

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
    fmt.Println("Счетчик:", counter)
}
```

#### sync.RWMutex
RWMutex позволяет множественным читателям или одному писателю одновременно.

```go
var rwMutex sync.RWMutex

// Для чтения
rwMutex.RLock()
// чтение данных
rwMutex.RUnlock()

// Для записи
rwMutex.Lock()
// запись данных
rwMutex.Unlock()
```

#### Когда использовать мьютекс вместо каналов?
"Mutex для защиты состояния, каналы для коммуникации"

- Используйте мьютексы, когда нужно защитить состояние (переменные, структуры данных)
- Используйте каналы, когда нужно передавать данные между горутинами

### Context

Context используется для отмены операций и передачи метаданных.

#### WithCancel
```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    select {
    case <-time.After(1 * time.Second):
        fmt.Println("Операция завершена")
    case <-ctx.Done():
        fmt.Println("Операция отменена")
    }
}()

time.Sleep(500 * time.Millisecond)
cancel() // Отменяем операцию
time.Sleep(1 * time.Second)
```

#### WithTimeout
```go
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
defer cancel()

select {
case <-time.After(2 * time.Second):
    fmt.Println("Операция заняла больше времени")
case <-ctx.Done():
    fmt.Println("Операция завершена или отменена:", ctx.Err())
}
```

#### WithValue
```go
type key string
const userIDKey key = "userID"

ctx := context.WithValue(context.Background(), userIDKey, "12345")
userID := ctx.Value(userIDKey).(string)
fmt.Println("User ID:", userID)
```

## Вечер

### Интерфейсы

#### Неявная реализация
В Go интерфейсы реализуются неявно. Тип реализует интерфейс, если он имеет все методы этого интерфейса.

```go
type Writer interface {
    Write([]byte) (int, error)
}

type File struct{}

func (f File) Write(data []byte) (int, error) {
    // реализация записи в файл
    return len(data), nil
}

// File автоматически реализует интерфейс Writer
```

#### Стандартные интерфейсы

##### io.Reader
```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

##### io.Writer
```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

##### fmt.Stringer
```go
type Stringer interface {
    String() string
}

type Person struct {
    Name string
    Age  int
}

func (p Person) String() string {
    return fmt.Sprintf("%s (%d лет)", p.Name, p.Age)
}
```

### Работа с памятью

#### Стек vs Куча

##### Стек
- Быстрый доступ к памяти
- Автоматическое освобождение при выходе из функции
- Ограниченный размер (обычно 1-2MB)

##### Куча
- Динамическая память
- Управляется garbage collector
- Медленнее доступ, но больше места

#### Указатели

```go
func main() {
    x := 42          // значение хранится в стеке
    p := &x          // указатель на x
    *p = 21          // изменяем значение через указатель
    fmt.Println(x)   // выводит 21
}
```

#### Escape Analysis

Escape analysis - это процесс, с помощью которого компилятор Go определяет, где должна быть размещена переменная: в стеке или в куче.

```go
func createPointer() *int {
    x := 42
    return &x  // x "escape" в кучу, так как возвращается указатель
}
```

## Что запомнить дословно

1. "Горутина — это легковесный поток, управляемый рантаймом Go"
2. "Mutex для защиты состояния, каналы для коммуникации"
3. "Context для отмены операций и передачи метаданных"

## Практические задания

1. Напишите программу, которая создает 10 горутин, каждая из которых увеличивает общий счетчик на 1. Используйте мьютекс для защиты счетчика.
2. Создайте буферизированный канал и отправьте в него 5 значений. Затем получите их в другой горутине.
3. Реализуйте функцию, которая принимает context и выполняет какую-либо операцию с таймаутом в 2 секунды.
4. Создайте собственный тип, который реализует интерфейс Stringer.
5. Напишите пример, демонстрирующий разницу между стеком и кучей.