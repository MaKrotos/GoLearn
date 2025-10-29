# Бенчмаркинг в Go: Полная теория

## Введение в бенчмаркинг

### Что такое бенчмаркинг?

Бенчмаркинг в Go - это **процесс измерения производительности** кода для:
- Определения **скорости выполнения**
- Измерения **использования памяти**
- Подсчета **аллокаций**
- Сравнения **разных реализаций**

### Зачем нужен бенчмаркинг?

1. **Оптимизация** - найти узкие места в коде
2. **Сравнение** - выбрать лучшую реализацию
3. **Регрессионное тестирование** - убедиться, что изменения не ухудшили производительность
4. **Документация** - показать ожидаемую производительность

## Основы бенчмаркингa

### Структура бенчмарков

Бенчмарки пишутся в файлах с суффиксом `_test.go`:

```go
// math_test.go
package math

import "testing"

func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(1, 2)
    }
}
```

### Ключевые особенности

1. **Имя функции** начинается с `Benchmark`
2. **Параметр** `*testing.B` вместо `*testing.T`
3. **Цикл** `for i := 0; i < b.N` - фреймворк управляет `b.N`
4. **Измерения** автоматически собираются Go

### Запуск бенчмарков

```bash
# Запустить все бенчмарки
go test -bench=.

# Запустить определенные бенчмарки
go test -bench=BenchmarkAdd

# Запустить с измерением памяти
go test -bench=. -benchmem

# Запустить с профилированием CPU
go test -bench=. -cpuprofile=cpu.prof

# Запустить с профилированием памяти
go test -bench=. -memprofile=mem.prof

# Запустить с несколькими итерациями
go test -bench=. -count=5

# Запустить с таймаутом
go test -bench=. -timeout=30s
```

## Подробное изучение бенчмарков

### 1. Базовая структура

```go
func BenchmarkExample(b *testing.B) {
    // Подготовка (setup) - выполняется один раз
    data := prepareTestData()
    
    // Измеряемая часть - выполняется b.N раз
    b.ResetTimer() // Сброс таймера после подготовки
    for i := 0; i < b.N; i++ {
        process(data)
    }
}
```

### 2. Управление таймером

```go
func BenchmarkWithSetup(b *testing.B) {
    // Дорогая подготовка
    expensiveSetup := func() []int {
        result := make([]int, 1000000)
        for i := range result {
            result[i] = i
        }
        return result
    }
    
    data := expensiveSetup()
    
    // Сброс таймера - не учитываем время подготовки
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

func BenchmarkWithPause(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Начинаем измерение
        data := generateData()
        
        // Приостанавливаем измерение для подготовки
        b.StopTimer()
        prepared := prepareData(data)
        b.StartTimer()
        
        // Продолжаем измерение
        process(prepared)
    }
}
```

### 3. Измерение аллокаций

```go
func BenchmarkAllocation(b *testing.B) {
    b.ReportAllocs() // Показать количество аллокаций
    
    for i := 0; i < b.N; i++ {
        result := make([]int, 1000)
        processSlice(result)
    }
}
```

## Примеры бенчмарков

### 1. Сравнение алгоритмов

```go
// algorithms/sort.go
package algorithms

import "sort"

func BubbleSort(arr []int) []int {
    result := make([]int, len(arr))
    copy(result, arr)
    
    n := len(result)
    for i := 0; i < n-1; i++ {
        for j := 0; j < n-i-1; j++ {
            if result[j] > result[j+1] {
                result[j], result[j+1] = result[j+1], result[j]
            }
        }
    }
    return result
}

func QuickSort(arr []int) []int {
    if len(arr) < 2 {
        return arr
    }
    
    pivot := arr[0]
    var less, equal, greater []int
    
    for _, value := range arr {
        switch {
        case value < pivot:
            less = append(less, value)
        case value == pivot:
            equal = append(equal, value)
        case value > pivot:
            greater = append(greater, value)
        }
    }
    
    less = QuickSort(less)
    greater = QuickSort(greater)
    
    result := make([]int, 0, len(arr))
    result = append(result, less...)
    result = append(result, equal...)
    result = append(result, greater...)
    
    return result
}
```

```go
// algorithms/sort_test.go
package algorithms

import (
    "math/rand"
    "testing"
    "time"
)

func generateRandomSlice(size int) []int {
    rand.Seed(time.Now().UnixNano())
    slice := make([]int, size)
    for i := range slice {
        slice[i] = rand.Intn(1000)
    }
    return slice
}

func BenchmarkBubbleSort(b *testing.B) {
    data := generateRandomSlice(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Создаем копию для каждого теста
        testData := make([]int, len(data))
        copy(testData, data)
        
        BubbleSort(testData)
    }
}

func BenchmarkQuickSort(b *testing.B) {
    data := generateRandomSlice(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        testData := make([]int, len(data))
        copy(testData, data)
        
        QuickSort(testData)
    }
}

func BenchmarkStdlibSort(b *testing.B) {
    data := generateRandomSlice(1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        testData := make([]int, len(data))
        copy(testData, data)
        
        sort.Ints(testData)
    }
}
```

### 2. Сравнение структур данных

```go
// datastructures/map.go
package datastructures

import "sync"

type SafeMap struct {
    mu   sync.RWMutex
    data map[string]interface{}
}

func NewSafeMap() *SafeMap {
    return &SafeMap{
        data: make(map[string]interface{}),
    }
}

func (sm *SafeMap) Set(key string, value interface{}) {
    sm.mu.Lock()
    sm.data[key] = value
    sm.mu.Unlock()
}

func (sm *SafeMap) Get(key string) (interface{}, bool) {
    sm.mu.RLock()
    value, exists := sm.data[key]
    sm.mu.RUnlock()
    return value, exists
}

type ShardedMap struct {
    shards []*sync.Map
    count  int
}

func NewShardedMap(shardCount int) *ShardedMap {
    shards := make([]*sync.Map, shardCount)
    for i := range shards {
        shards[i] = &sync.Map{}
    }
    return &ShardedMap{
        shards: shards,
        count:  shardCount,
    }
}

func (sm *ShardedMap) getShard(key string) *sync.Map {
    hash := 0
    for _, b := range []byte(key) {
        hash = hash*31 + int(b)
    }
    return sm.shards[hash%sm.count]
}

func (sm *ShardedMap) Set(key string, value interface{}) {
    shard := sm.getShard(key)
    shard.Store(key, value)
}

func (sm *ShardedMap) Get(key string) (interface{}, bool) {
    shard := sm.getShard(key)
    return shard.Load(key)
}
```

```go
// datastructures/map_test.go
package datastructures

import (
    "strconv"
    "sync"
    "testing"
)

func BenchmarkSyncMap_Set(b *testing.B) {
    var sm sync.Map
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            sm.Store(strconv.Itoa(i), i)
            i++
        }
    })
}

func BenchmarkSafeMap_Set(b *testing.B) {
    sm := NewSafeMap()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            sm.Set(strconv.Itoa(i), i)
            i++
        }
    })
}

func BenchmarkShardedMap_Set(b *testing.B) {
    sm := NewShardedMap(16)
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            sm.Set(strconv.Itoa(i), i)
            i++
        }
    })
}

func BenchmarkSyncMap_Get(b *testing.B) {
    var sm sync.Map
    // Предварительно заполняем данными
    for i := 0; i < 1000; i++ {
        sm.Store(strconv.Itoa(i), i)
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            sm.Load(strconv.Itoa(i % 1000))
            i++
        }
    })
}

func BenchmarkSafeMap_Get(b *testing.B) {
    sm := NewSafeMap()
    // Предварительно заполняем данными
    for i := 0; i < 1000; i++ {
        sm.Set(strconv.Itoa(i), i)
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            sm.Get(strconv.Itoa(i % 1000))
            i++
        }
    })
}

func BenchmarkShardedMap_Get(b *testing.B) {
    sm := NewShardedMap(16)
    // Предварительно заполняем данными
    for i := 0; i < 1000; i++ {
        sm.Set(strconv.Itoa(i), i)
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            sm.Get(strconv.Itoa(i % 1000))
            i++
        }
    })
}
```

## Продвинутые техники бенчмаркинга

### 1. Параллельные бенчмарки

```go
func BenchmarkParallel(b *testing.B) {
    var counter int64
    var mutex sync.Mutex
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            mutex.Lock()
            counter++
            mutex.Unlock()
        }
    })
}

func BenchmarkParallelAtomic(b *testing.B) {
    var counter int64
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            atomic.AddInt64(&counter, 1)
        }
    })
}
```

### 2. Бенчмарки с подтестами

```go
func BenchmarkAlgorithms(b *testing.B) {
    sizes := []int{100, 1000, 10000}
    
    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
            data := generateRandomSlice(size)
            
            b.Run("BubbleSort", func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    testData := make([]int, len(data))
                    copy(testData, data)
                    BubbleSort(testData)
                }
            })
            
            b.Run("QuickSort", func(b *testing.B) {
                for i := 0; i < b.N; i++ {
                    testData := make([]int, len(data))
                    copy(testData, data)
                    QuickSort(testData)
                }
            })
        })
    }
}
```

### 3. Измерение конкретных метрик

```go
func BenchmarkCustomMetrics(b *testing.B) {
    // Измеряем количество операций в секунду
    b.Run("OperationsPerSecond", func(b *testing.B) {
        start := time.Now()
        operations := 0
        
        for i := 0; i < b.N; i++ {
            processOperation()
            operations++
        }
        
        duration := time.Since(start)
        opsPerSecond := float64(operations) / duration.Seconds()
        b.ReportMetric(opsPerSecond, "ops/sec")
    })
    
    // Измеряем размер обработанных данных
    b.Run("Throughput", func(b *testing.B) {
        dataSize := 1024 * 1024 // 1MB
        data := make([]byte, dataSize)
        
        start := time.Now()
        totalBytes := 0
        
        for i := 0; i < b.N; i++ {
            processData(data)
            totalBytes += dataSize
        }
        
        duration := time.Since(start)
        throughput := float64(totalBytes) / duration.Seconds() / (1024 * 1024) // MB/s
        b.ReportMetric(throughput, "MB/s")
    })
}
```

## Анализ результатов бенчмарков

### 1. Интерпретация вывода

```
BenchmarkAdd-8                    1000000000    0.25 ns/op
BenchmarkBubbleSort100-8           200000      5800 ns/op
BenchmarkQuickSort100-8           1000000      1200 ns/op
BenchmarkStdlibSort100-8          1000000      1100 ns/op
```

Разбор:
- `BenchmarkAdd-8` - имя бенчмарка с количеством CPU ядер
- `1000000000` - количество итераций (b.N)
- `0.25 ns/op` - время на операцию

### 2. Сравнение с -benchmem

```
BenchmarkStringConcat-8           5000000      320 ns/op     480 B/op     10 allocs/op
BenchmarkStringBuilder-8         50000000      32.0 ns/op    64 B/op      1 allocs/op
```

Разбор:
- `480 B/op` - байт памяти на операцию
- `10 allocs/op` - аллокаций на операцию

### 3. Использование benchstat для сравнения

```bash
# Запустить бенчмарки и сохранить результаты
go test -bench=. -count=5 > old.txt

# Внести изменения в код

# Запустить бенчмарки снова
go test -bench=. -count=5 > new.txt

# Сравнить результаты
benchstat old.txt new.txt
```

Пример вывода benchstat:
```
name              old time/op    new time/op    delta
Add-8              0.25ns ± 1%    0.23ns ± 2%   -8.00%  (p=0.008 n=5+5)
BubbleSort100-8    5.80µs ± 0%    5.20µs ± 1%  -10.34%  (p=0.008 n=5+5)
```

## Профилирование с бенчмарками

### 1. CPU профилирование

```bash
# Запустить бенчмарки с CPU профайлингом
go test -bench=BenchmarkComplex -cpuprofile=cpu.prof

# Анализ профиля
go tool pprof cpu.prof

# В интерактивном режиме pprof:
# top - показать топ функций
# list functionName - показать исходный код с профилированием
# web - визуализация в браузере (требует Graphviz)
```

### 2. Память профилирование

```bash
# Запустить бенчмарки с профайлингом памяти
go test -bench=BenchmarkMemoryIntensive -memprofile=mem.prof

# Анализ профиля памяти
go tool pprof -alloc_space mem.prof

# В интерактивном режиме:
# top - показать топ аллокаций
# list functionName - показать аллокации в функции
```

### 3. Блокировки профилирование

```bash
# Запустить бенчмарки с профайлингом блокировок
go test -bench=BenchmarkConcurrent -blockprofile=block.prof

# Анализ профиля блокировок
go tool pprof block.prof
```

## Лучшие практики бенчмаркинга

### 1. Стабильные условия

```go
func TestMain(m *testing.M) {
    // Отключить планировщик Go для стабильных результатов
    // (только для бенчмарков, не для обычных тестов)
    runtime.GOMAXPROCS(1)
    
    // Запустить тесты
    os.Exit(m.Run())
}
```

### 2. Предварительный прогрев

```go
func BenchmarkWithWarmup(b *testing.B) {
    // Прогрев - запускаем код несколько раз перед измерением
    for i := 0; i < 1000; i++ {
        process(data)
    }
    
    // Начинаем измерение
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        process(data)
    }
}
```

### 3. Избегайте побочных эффектов

```go
// ПЛОХО - побочные эффекты
func BenchmarkDatabase(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Каждый вызов изменяет состояние базы данных
        db.Insert(User{Name: fmt.Sprintf("User%d", i)})
    }
}

// ХОРОШО - изолированные тесты
func BenchmarkDatabase(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Создаем уникальные данные для каждого теста
        user := User{ID: i, Name: fmt.Sprintf("User%d", i)}
        db.Upsert(user) // Используем upsert вместо insert
    }
}
```

### 4. Тестирование реалистичных сценариев

```go
func BenchmarkRealisticScenario(b *testing.B) {
    // Создаем реалистичную нагрузку
    users := generateUsers(1000)
    products := generateProducts(100)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Симулируем реальный сценарий использования
        cart := CreateCart(users[i%len(users)])
        for j := 0; j < 5; j++ {
            cart.AddProduct(products[j])
        }
        order := cart.Checkout()
        processOrder(order)
    }
}
```

## Распространенные ошибки

### 1. Неправильная инициализация

```go
// ПЛОХО - инициализация внутри цикла
func BenchmarkBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Дорогая инициализация
        data := initializeExpensiveData()
        process(data)
    }
}

// ХОРОШО - инициализация вне цикла
func BenchmarkGood(b *testing.B) {
    // Инициализация один раз
    data := initializeExpensiveData()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        process(data)
    }
}
```

### 2. Игнорирование аллокаций

```go
// ПЛОХО - создание новых объектов в цикле
func BenchmarkBadAlloc(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // Создаем новый слайс каждый раз
        data := make([]int, 1000)
        processData(data)
    }
}

// ХОРОШО - переиспользование объектов
func BenchmarkGoodAlloc(b *testing.B) {
    // Создаем один раз
    data := make([]int, 1000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Переиспользуем слайс
        processData(data)
        // Сбрасываем состояние если нужно
        for j := range data {
            data[j] = 0
        }
    }
}
```

### 3. Нестабильные результаты

```go
// ПЛОХО - нестабильные результаты
func BenchmarkUnstable(b *testing.B) {
    // Используем случайные данные
    data := generateRandomData()
    
    for i := 0; i < b.N; i++ {
        process(data)
    }
}

// ХОРОШО - стабильные результаты
func BenchmarkStable(b *testing.B) {
    // Используем детерминированные данные
    data := generateDeterministicData()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        process(data)
    }
}
```

## Интеграция с CI/CD

### 1. Автоматическое тестирование производительности

```yaml
# .github/workflows/benchmark.yml
name: Benchmark
on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.19
    
    - name: Run benchmarks
      run: |
        go test -bench=. -benchmem > benchmark.txt
    
    - name: Store benchmark results
      uses: actions/upload-artifact@v2
      with:
        name: benchmark-results
        path: benchmark.txt
```

### 2. Сравнение производительности

```bash
# Скрипт для сравнения производительности
#!/bin/bash

# Запустить бенчмарки на текущей ветке
go test -bench=. -count=5 > new.txt

# Запустить бенчмарки на main ветке
git checkout main
go test -bench=. -count=5 > old.txt

# Сравнить результаты
benchstat old.txt new.txt

# Вернуться к исходной ветке
git checkout -
```

## См. также

- [Тестирование](../concepts/testing.md) - основы тестирования
- [Профилирование](../concepts/profiling.md) - как использовать профилирование
- [Горутины и каналы](../concepts/goroutine.md) - параллельное программирование
- [Синхронизация](../concepts/mutex.md) - синхронизация доступа к данным
- [Практические примеры](../examples/benchmarking) - примеры кода