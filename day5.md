# День 5: Производительность и отладка (12 часов)

## Инструменты

### pprof — CPU и memory profiling

pprof - это инструмент для профилирования Go приложений, который помогает находить узкие места в производительности.

#### Включение pprof в приложении
```go
package main

import (
    "net/http"
    _ "net/http/pprof" // Импортируем для включения pprof
)

func main() {
    // Запускаем HTTP сервер для pprof на отдельном порту
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
    
    // Основное приложение
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
    // Имитация нагрузки
    processData()
    w.Write([]byte("Hello, World!"))
}

func processData() {
    // Имитация обработки данных
    data := make([]int, 1000000)
    for i := range data {
        data[i] = i * i
    }
}
```

#### Использование pprof
```bash
# Запуск приложения
go run main.go

# Сбор CPU профиля
go tool pprof http://localhost:6060/debug/pprof/profile

# Сбор memory профиля
go tool pprof http://localhost:6060/debug/pprof/heap

# Сбор профиля блокировок
go tool pprof http://localhost:6060/debug/pprof/block
```

#### Анализ профиля в интерактивном режиме
```bash
# После запуска pprof
(pprof) top          # Показать топ функций
(pprof) web          # Визуализация в браузере
(pprof) list main    # Показать исходный код функции
(pprof) quit
```

#### Пример оптимизации на основе профиля
```go
// До оптимизации
func processLargeData() []int {
    result := []int{}
    for i := 0; i < 1000000; i++ {
        result = append(result, i*i) // Много аллокаций
    }
    return result
}

// После оптимизации
func processLargeDataOptimized() []int {
    result := make([]int, 1000000) // Предварительная аллокация
    for i := 0; i < 1000000; i++ {
        result[i] = i * i
    }
    return result
}
```

### go test -bench . — бенчмарки

Бенчмарки позволяют измерять производительность кода и сравнивать разные реализации.

#### Написание бенчмарков
```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}

func Multiply(a, b int) int {
    return a * b
}

// Рекурсивное вычисление Фибоначчи (медленно)
func Fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return Fibonacci(n-1) + Fibonacci(n-2)
}

// Итеративное вычисление Фибоначчи (быстро)
func FibonacciIterative(n int) int {
    if n <= 1 {
        return n
    }
    
    a, b := 0, 1
    for i := 2; i <= n; i++ {
        a, b = b, a+b
    }
    return b
}
```

```go
// math_test.go
package math

import (
    "testing"
)

func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(1, 2)
    }
}

func BenchmarkMultiply(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Multiply(1, 2)
    }
}

func BenchmarkFibonacci(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Fibonacci(20)
    }
}

func BenchmarkFibonacciIterative(b *testing.B) {
    for i := 0; i < b.N; i++ {
        FibonacciIterative(20)
    }
}

// Бенчмарк с различными входными данными
func BenchmarkFibonacciIterativeN(b *testing.B) {
    tests := []int{10, 20, 30, 40}
    for _, n := range tests {
        b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
            for i := 0; i < b.N; i++ {
                FibonacciIterative(n)
            }
        })
    }
}
```

#### Запуск бенчмарков
```bash
# Запуск всех бенчмарков
go test -bench=.

# Запуск с детальной информацией
go test -bench=. -benchmem

# Запуск определенного бенчмарка
go test -bench=BenchmarkFibonacci

# Запуск с профилированием CPU
go test -bench=. -cpuprofile=cpu.prof

# Запуск с профилированием памяти
go test -bench=. -memprofile=mem.prof
```

#### Анализ результатов бенчмарков
```
BenchmarkAdd-8                    1000000000    0.25 ns/op
BenchmarkMultiply-8               1000000000    0.26 ns/op
BenchmarkFibonacci-8              30000         40200 ns/op
BenchmarkFibonacciIterative-8     20000000      65.2 ns/op
```

### go test -race — детектор гонок

Детектор гонок помогает находить data race - ситуации, когда несколько горутин одновременно обращаются к одной переменной, и хотя бы одна из них выполняет запись.

#### Пример кода с data race
```go
package main

import (
    "fmt"
    "sync"
    "time"
)

var counter int

func main() {
    var wg sync.WaitGroup
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++ // Data race! Несколько горутин одновременно изменяют counter
        }()
    }
    
    wg.Wait()
    fmt.Println("Счетчик:", counter)
}
```

#### Запуск с детектором гонок
```bash
# Запуск приложения с детектором гонок
go run -race main.go

# Запуск тестов с детектором гонок
go test -race .
```

#### Исправление data race
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

func main() {
    var wg sync.WaitGroup
    
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mutex.Lock()
            counter++
            mutex.Unlock()
        }()
    }
    
    wg.Wait()
    fmt.Println("Счетчик:", counter)
}
```

## Что говорить на собеседовании

### Про профилирование
"Для профилирования использую pprof, смотрю на аллокации и время выполнения"

#### Что анализировать:
1. **CPU профиль** - где тратится процессорное время
2. **Memory профиль** - где происходят аллокации памяти
3. **Block профиль** - где происходят блокировки
4. **Mutex профиль** - где происходят конфликты мьютексов

#### Пример вывода pprof:
```
(pprof) top
Showing nodes accounting for 400ms, 100% of 400ms total
      flat  flat%   sum%        cum   cum%
     200ms 50.00% 50.00%      200ms 50.00%  main.processData
     100ms 25.00% 75.00%      100ms 25.00%  runtime.mallocgc
     100ms 25.00% 100.00%      400ms   100%  main.handler
```

### Про детектор гонок
"Всегда запускаю с -race для обнаружения data race"

#### Когда использовать:
1. При разработке многопоточного кода
2. Перед релизом в production
3. При тестировании конкурентных алгоритмов

#### Что обнаруживает:
1. **Data race** - одновременный доступ к переменной без синхронизации
2. **Неправильное использование мьютексов**
3. **Гонки в каналах**

## Практические задания

1. Создайте приложение с нагрузкой и проанализируйте его с помощью pprof.
2. Напишите бенчмарки для разных реализаций одного алгоритма и сравните их производительность.
3. Найдите и исправьте data race в предоставленном коде.
4. Оптимизируйте код на основе результатов профилирования.
5. Настройте CI/CD pipeline для автоматического запуска тестов с детектором гонок.