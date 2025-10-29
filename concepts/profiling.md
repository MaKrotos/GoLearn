# Профилирование в Go - объяснение для чайников

## Что такое профилирование?

Представьте, что вы тренер футбольной команды. Чтобы улучшить игру, вы **анализируете**:
- Где игроки тратят больше всего времени?
- Где происходят ошибки?
- Какие действия самые медленные?

Профилирование в программировании - это **анализ работы программы** для нахождения:
- **Медленных участков** кода
- **Проблем с памятью**
- **Утечек ресурсов**

## Инструмент pprof

### Что такое pprof?

`pprof` - это **встроенный профайлер** в Go, который помогает:
- Измерять **время выполнения** функций
- Анализировать **использование памяти**
- Находить **блокировки** и **гонки**

### Как включить pprof?

```go
package main

import (
    "net/http"
    _ "net/http/pprof" // Включаем pprof
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
```

### Как использовать pprof?

```bash
# Сбор CPU профиля (30 секунд)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Сбор memory профиля
go tool pprof http://localhost:6060/debug/pprof/heap

# Сбор профиля блокировок
go tool pprof http://localhost:6060/debug/pprof/block

# Сбор профиля мьютексов
go tool pprof http://localhost:6060/debug/pprof/mutex
```

## Типы профилей

### 1. CPU профиль

Показывает, **где тратится процессорное время**.

```bash
# Сбор CPU профиля
go tool pprof http://localhost:6060/debug/pprof/profile

# Интерактивный режим
(pprof) top          # Топ функций по времени
(pprof) web          # Визуализация в браузере
(pprof) list main    # Исходный код с аннотациями
(pprof) quit
```

Пример вывода:
```
(pprof) top
Showing nodes accounting for 400ms, 100% of 400ms total
      flat  flat%   sum%        cum   cum%
     200ms 50.00% 50.00%      200ms 50.00%  main.processData
     100ms 25.00% 75.00%      100ms 25.00%  runtime.mallocgc
     100ms 25.00% 100.00%      400ms   100%  main.handler
```

### 2. Memory профиль

Показывает, **где происходят аллокации памяти**.

```bash
# Сбор memory профиля
go tool pprof http://localhost:6060/debug/pprof/heap

# Интерактивный режим
(pprof) top          # Топ функций по аллокациям
(pprof) inuse_space  # Используемая память
(pprof) alloc_space  # Всего аллоцировано
(pprof) web          # Визуализация
```

### 3. Block профиль

Показывает, **где происходят блокировки** (ожидание каналов, мьютексов).

```bash
# Сбор block профиля
go tool pprof http://localhost:6060/debug/pprof/block
```

### 4. Mutex профиль

Показывает, **где происходят конфликты мьютексов**.

```bash
# Сбор mutex профиля
go tool pprof http://localhost:6060/debug/pprof/mutex
```

## Практический пример

### Проблемный код:

```go
// bad_code.go
package main

import (
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "strings"
    "time"
)

func main() {
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
    
    http.HandleFunc("/bad", badHandler)
    http.ListenAndServe(":8080", nil)
}

func badHandler(w http.ResponseWriter, r *http.Request) {
    // Проблема 1: Много аллокаций
    var result string
    for i := 0; i < 100000; i++ {
        result += fmt.Sprintf("item%d,", i) // Каждая конкатенация создает новую строку
    }
    
    // Проблема 2: Долгая операция без контекста
    time.Sleep(5 * time.Second)
    
    w.Write([]byte(result))
}
```

### Профилирование:

```bash
# Запускаем приложение
go run bad_code.go

# В другом терминале вызываем endpoint
curl http://localhost:8080/bad

# Собираем CPU профиль
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=10
```

### Анализ:

```
(pprof) top
Showing nodes accounting for 800ms, 90% of 890ms total
      flat  flat%   sum%        cum   cum%
     500ms 56.2% 56.2%      500ms 56.2%  runtime.usleep      # time.Sleep
     200ms 22.5% 78.7%      200ms 22.5%  runtime.mallocgc    # Аллокации
     100ms 11.2% 90.0%      100ms 11.2%  main.badHandler     # Основная функция
```

### Исправленный код:

```go
// good_code.go
package main

import (
    "context"
    "fmt"
    "net/http"
    _ "net/http/pprof"
    "strings"
    "time"
)

func main() {
    go func() {
        http.ListenAndServe(":6060", nil)
    }()
    
    http.HandleFunc("/good", goodHandler)
    http.ListenAndServe(":8080", nil)
}

func goodHandler(w http.ResponseWriter, r *http.Request) {
    // Создаем контекст с таймаутом
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    
    // Используем strings.Builder для эффективной конкатенации
    var builder strings.Builder
    for i := 0; i < 100000; i++ {
        builder.WriteString(fmt.Sprintf("item%d,", i))
    }
    
    // Проверяем контекст
    select {
    case <-ctx.Done():
        http.Error(w, "Таймаут", http.StatusRequestTimeout)
        return
    default:
        w.Write([]byte(builder.String()))
    }
}
```

## Бенчмарки и профилирование

### Создание бенчмарка:

```go
// math.go
func Fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return Fibonacci(n-1) + Fibonacci(n-2)
}

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
```

### Профилирование бенчмарка:

```bash
# Запуск бенчмарка с CPU профилем
go test -bench=BenchmarkFibonacci -cpuprofile=cpu.prof

# Анализ профиля
go tool pprof cpu.prof

# Запуск бенчмарка с memory профилем
go test -bench=BenchmarkFibonacci -memprofile=mem.prof

# Анализ memory профиля
go tool pprof mem.prof
```

## Визуализация профилей

### Web интерфейс:

```bash
# Открываем web интерфейс
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile
```

Это откроет браузер с графическим представлением профиля.

### Граф в консоли:

```bash
(pprof) png > profile.png    # Сохранить как PNG
(pprof) svg > profile.svg    # Сохранить как SVG
(pprof) web                  # Открыть в браузере
```

## Лучшие практики

### 1. Регулярное профилирование

```go
// В production приложениях включайте pprof условно
var enableProfiling = os.Getenv("ENABLE_PROFILING") == "true"

func main() {
    if enableProfiling {
        go func() {
            log.Println("pprof server started on :6060")
            http.ListenAndServe(":6060", nil)
        }()
    }
    
    // Основное приложение
}
```

### 2. Используйте теги сборки

```go
// profiling.go
// +build profiling

package main

import _ "net/http/pprof"
```

```go
// main.go
// +build !profiling

package main

// Нет pprof в production сборке
```

Сборка:
```bash
# С профилированием
go build -tags profiling

# Без профилирования
go build
```

### 3. Мониторинг в production

```go
// Добавляем метрики для мониторинга
import "github.com/prometheus/client_golang/prometheus"

var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "request_duration_seconds",
            Help: "Request duration in seconds",
        },
        []string{"handler"},
    )
)

func instrumentedHandler(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next(w, r)
        duration := time.Since(start)
        
        requestDuration.WithLabelValues(r.URL.Path).Observe(duration.Seconds())
    }
}
```

## Распространенные ошибки

### 1. Профилирование в production без защиты

```go
// ПЛОХО - pprof доступен всем
func main() {
    go http.ListenAndServe(":6060", nil) // Открыт для всех!
}

// ЛУЧШЕ - защита доступа
func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    
    // Добавляем аутентификацию
    server := &http.Server{
        Addr:    ":6060",
        Handler: authMiddleware(mux),
    }
    
    go server.ListenAndServe()
}
```

### 2. Игнорирование overhead

Профилирование добавляет нагрузку на приложение. Не забывайте об этом при интерпретации результатов.

## См. также

- [Тестирование производительности](../theory/benchmarking.md) - как писать бенчмарки
- [Работа с памятью](memory.md) - основы управления памятью
- [Горутины](goroutine.md) - как профилировать конкурентный код
- [Мьютексы](mutex.md) - анализ блокировок