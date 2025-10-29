# Кэширование в Go: Полная теория

## Введение в кэширование

### Что такое кэширование?

Кэширование - это **механизм временного хранения** часто запрашиваемых данных в быстродоступном хранилище для:
- **Уменьшения** времени доступа к данным
- **Снижения** нагрузки на основную систему хранения
- **Повышения** общей производительности приложения

### Типы кэширования

1. **In-memory кэширование** - данные хранятся в оперативной памяти
2. **Распределенное кэширование** - данные распределены между несколькими узлами
3. **Браузерное кэширование** - кэширование на стороне клиента
4. **CDN кэширование** - кэширование на уровне сети доставки контента

## In-Memory кэширование в Go

### Простой кэш с использованием sync.Map

```go
// caching/simple_cache.go
package caching

import (
    "sync"
    "time"
)

// CacheItem элемент кэша с временем жизни
type CacheItem struct {
    Value      interface{}
    Expiration int64 // Unix timestamp
}

// SimpleCache простой in-memory кэш
type SimpleCache struct {
    items sync.Map
}

// NewSimpleCache создает новый кэш
func NewSimpleCache() *SimpleCache {
    cache := &SimpleCache{}
    // Запускаем очистку просроченных элементов
    go cache.cleanup()
    return cache
}

// Set добавляет элемент в кэш
func (c *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
    expiration := time.Now().Add(ttl).Unix()
    item := CacheItem{
        Value:      value,
        Expiration: expiration,
    }
    c.items.Store(key, item)
}

// Get получает элемент из кэша
func (c *SimpleCache) Get(key string) (interface{}, bool) {
    item, ok := c.items.Load(key)
    if !ok {
        return nil, false
    }
    
    cacheItem := item.(CacheItem)
    // Проверяем, не истекло ли время жизни
    if time.Now().Unix() > cacheItem.Expiration {
        c.items.Delete(key)
        return nil, false
    }
    
    return cacheItem.Value, true
}

// Delete удаляет элемент из кэша
func (c *SimpleCache) Delete(key string) {
    c.items.Delete(key)
}

// cleanup очищает просроченные элементы
func (c *SimpleCache) cleanup() {
    for {
        time.Sleep(time.Minute) // Проверяем каждую минуту
        c.items.Range(func(key, value interface{}) bool {
            cacheItem := value.(CacheItem)
            if time.Now().Unix() > cacheItem.Expiration {
                c.items.Delete(key)
            }
            return true
        })
    }
}
```

### Пример использования простого кэша

```go
// caching/simple_cache_example.go
package main

import (
    "fmt"
    "time"
    "your-project/caching"
)

func main() {
    cache := caching.NewSimpleCache()
    
    // Добавляем данные в кэш
    cache.Set("user:1", "John Doe", time.Minute*5)
    cache.Set("user:2", "Jane Smith", time.Minute*10)
    
    // Получаем данные из кэша
    if user, found := cache.Get("user:1"); found {
        fmt.Printf("Найден пользователь: %s\n", user)
    }
    
    // Ждем истечения времени жизни
    time.Sleep(time.Minute * 6)
    
    // Проверяем, что данные исчезли
    if _, found := cache.Get("user:1"); !found {
        fmt.Println("Пользователь больше не в кэше")
    }
}
```

## Распределенное кэширование с Redis

### Установка и настройка Redis

Для работы с Redis в Go используем популярную библиотеку `go-redis`:

```bash
go mod init your-project
go get github.com/go-redis/redis/v8
```

### Подключение к Redis

```go
// caching/redis_client.go
package caching

import (
    "context"
    "github.com/go-redis/redis/v8"
    "time"
)

var (
    RedisClient *redis.Client
    ctx         = context.Background()
)

// InitRedis инициализирует подключение к Redis
func InitRedis(addr string) {
    RedisClient = redis.NewClient(&redis.Options{
        Addr:     addr, // "localhost:6379"
        Password: "",   // без пароля
        DB:       0,    // использовать базу данных по умолчанию
    })
}

// CloseRedis закрывает подключение к Redis
func CloseRedis() {
    if RedisClient != nil {
        RedisClient.Close()
    }
}
```

### Операции с Redis

```go
// caching/redis_operations.go
package caching

import (
    "context"
    "encoding/json"
    "time"
)

// SetString сохраняет строку в Redis
func SetString(key, value string, ttl time.Duration) error {
    return RedisClient.Set(ctx, key, value, ttl).Err()
}

// GetString получает строку из Redis
func GetString(key string) (string, error) {
    return RedisClient.Get(ctx, key).Result()
}

// SetStruct сохраняет структуру в Redis
func SetStruct(key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return RedisClient.Set(ctx, key, data, ttl).Err()
}

// GetStruct получает структуру из Redis
func GetStruct(key string, dest interface{}) error {
    data, err := RedisClient.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(data), dest)
}

// Delete удаляет ключ из Redis
func Delete(key string) error {
    return RedisClient.Del(ctx, key).Err()
}

// Exists проверяет существование ключа
func Exists(key string) (bool, error) {
    result, err := RedisClient.Exists(ctx, key).Result()
    if err != nil {
        return false, err
    }
    return result > 0, nil
}
```

### Пример использования Redis

```go
// caching/redis_example.go
package main

import (
    "fmt"
    "log"
    "time"
    "your-project/caching"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // Инициализируем Redis
    caching.InitRedis("localhost:6377")
    defer caching.CloseRedis()
    
    // Работаем со строками
    err := caching.SetString("greeting", "Hello, Redis!", time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    
    greeting, err := caching.GetString("greeting")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(greeting)
    
    // Работаем со структурами
    user := User{
        ID:    1,
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    err = caching.SetStruct("user:1", user, time.Hour)
    if err != nil {
        log.Fatal(err)
    }
    
    var retrievedUser User
    err = caching.GetStruct("user:1", &retrievedUser)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Получен пользователь: %+v\n", retrievedUser)
}
```

## Паттерны кэширования

### Cache-Aside Pattern

```go
// caching/cache_aside.go
package caching

import (
    "encoding/json"
    "fmt"
    "time"
)

// UserService сервис для работы с пользователями
type UserService struct {
    // Здесь могла бы быть база данных
}

// User модель пользователя
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// GetUserWithCache получает пользователя с использованием кэша
func (s *UserService) GetUserWithCache(userID int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    
    // Пытаемся получить из кэша
    if data, err := GetString(cacheKey); err == nil {
        var user User
        if json.Unmarshal([]byte(data), &user) == nil {
            fmt.Println("Пользователь найден в кэше")
            return &user, nil
        }
    }
    
    // Если нет в кэше, получаем из источника данных
    user, err := s.getUserFromDatabase(userID) // Реализация зависит от БД
    if err != nil {
        return nil, err
    }
    
    // Сохраняем в кэш
    data, _ := json.Marshal(user)
    SetString(cacheKey, string(data), time.Minute*30)
    
    return user, nil
}

// getUserFromDatabase имитация получения данных из БД
func (s *UserService) getUserFromDatabase(userID int) (*User, error) {
    // Здесь была бы реальная реализация
    return &User{
        ID:    userID,
        Name:  fmt.Sprintf("User %d", userID),
        Email: fmt.Sprintf("user%d@example.com", userID),
    }, nil
}
```

### Write-Through Pattern

```go
// caching/write_through.go
package caching

import (
    "encoding/json"
    "fmt"
)

// UserServiceWithWriteThrough сервис с Write-Through кэшированием
type UserServiceWithWriteThrough struct {
    // Здесь могла бы быть база данных
}

// UpdateUserWithCache обновляет пользователя и кэш одновременно
func (s *UserServiceWithWriteThrough) UpdateUserWithCache(user *User) error {
    // Обновляем в источнике данных
    err := s.updateUserInDatabase(user) // Реализация зависит от БД
    if err != nil {
        return err
    }
    
    // Обновляем кэш
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    data, _ := json.Marshal(user)
    SetString(cacheKey, string(data), time.Minute*30)
    
    return nil
}

// updateUserInDatabase имитация обновления данных в БД
func (s *UserServiceWithWriteThrough) updateUserInDatabase(user *User) error {
    // Здесь была бы реальная реализация
    fmt.Printf("Обновлен пользователь в БД: %+v\n", user)
    return nil
}
```

## Стратегии инвалидации кэша

### По времени (TTL)

```go
// Установка времени жизни при добавлении в кэш
cache.Set("key", "value", time.Minute*5) // 5 минут
```

### По событиям

```go
// caching/event_based_invalidation.go
package caching

import "fmt"

// InvalidateUserCache инвалидирует кэш пользователя
func InvalidateUserCache(userID int) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    Delete(cacheKey)
    fmt.Printf("Кэш пользователя %d инвалидирован\n", userID)
}

// InvalidateRelatedCaches инвалидирует связанные кэши
func InvalidateRelatedCaches(userID int) {
    // Инвалидируем основной кэш пользователя
    InvalidateUserCache(userID)
    
    // Инвалидируем кэши связанных данных
    Delete(fmt.Sprintf("user:%d:posts", userID))
    Delete(fmt.Sprintf("user:%d:followers", userID))
    Delete(fmt.Sprintf("user:%d:following", userID))
}
```

## Мониторинг и метрики кэширования

### Сбор метрик

```go
// caching/metrics.go
package caching

import (
    "sync/atomic"
    "time"
)

// CacheMetrics метрики кэширования
type CacheMetrics struct {
    Hits   uint64 // Попадания в кэш
    Misses uint64 // Промахи
}

var metrics CacheMetrics

// RecordHit записывает попадание в кэш
func RecordHit() {
    atomic.AddUint64(&metrics.Hits, 1)
}

// RecordMiss записывает промах
func RecordMiss() {
    atomic.AddUint64(&metrics.Misses, 1)
}

// GetMetrics получает текущие метрики
func GetMetrics() CacheMetrics {
    return CacheMetrics{
        Hits:   atomic.LoadUint64(&metrics.Hits),
        Misses: atomic.LoadUint64(&metrics.Misses),
    }
}

// GetHitRate вычисляет коэффициент попаданий
func GetHitRate() float64 {
    metrics := GetMetrics()
    total := metrics.Hits + metrics.Misses
    if total == 0 {
        return 0
    }
    return float64(metrics.Hits) / float64(total)
}
```

### Интеграция метрик в кэш

```go
// caching/metrics_integration.go
package caching

// GetWithMetrics получает значение с записью метрик
func (c *SimpleCache) GetWithMetrics(key string) (interface{}, bool) {
    value, found := c.Get(key)
    if found {
        RecordHit()
    } else {
        RecordMiss()
    }
    return value, found
}
```

## Лучшие практики кэширования

### 1. Выбор правильного TTL

```go
const (
    ShortTTL  = time.Minute * 5   // Для часто меняющихся данных
    MediumTTL = time.Hour         // Для умеренно меняющихся данных
    LongTTL   = time.Hour * 24    // Для редко меняющихся данных
)
```

### 2. Предотвращение "thundering herd"

```go
// caching/singleflight.go
package caching

import (
    "golang.org/x/sync/singleflight"
)

var singleFlightGroup singleflight.Group

// GetWithSingleFlight получает данные с предотвращением thundering herd
func GetWithSingleFlight(key string, fetchFunc func() (interface{}, error)) (interface{}, error) {
    result, err, _ := singleFlightGroup.Do(key, fetchFunc)
    return result, err
}
```

### 3. Обработка ошибок кэширования

```go
// caching/error_handling.go
package caching

import "log"

// SafeGet безопасное получение из кэша
func SafeGet(key string) (interface{}, bool) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Ошибка при получении из кэша: %v", r)
        }
    }()
    
    return Get(key)
}

// SafeSet безопасная запись в кэш
func SafeSet(key string, value interface{}, ttl time.Duration) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Ошибка при записи в кэш: %v", r)
        }
    }()
    
    Set(key, value, ttl)
}
```

## Производительность кэширования

### Бенчмарки для кэша

```go
// caching/cache_bench_test.go
package caching

import (
    "testing"
    "time"
)

func BenchmarkSimpleCacheSet(b *testing.B) {
    cache := NewSimpleCache()
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        cache.Set(fmt.Sprintf("key%d", i), "value", time.Minute)
    }
}

func BenchmarkSimpleCacheGet(b *testing.B) {
    cache := NewSimpleCache()
    // Заполняем кэш данными
    for i := 0; i < 1000; i++ {
        cache.Set(fmt.Sprintf("key%d", i), "value", time.Minute)
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        cache.Get(fmt.Sprintf("key%d", i%1000))
    }
}
```

## Заключение

Кэширование - это мощный инструмент для повышения производительности приложений Go. Ключевые моменты:

1. **Выбор правильного типа кэша** - in-memory для локальных данных, Redis для распределенных
2. **Правильная стратегия инвалидации** - TTL, события, комбинированные подходы
3. **Мониторинг метрик** - отслеживание hit rate, latency, ошибок
4. **Обработка ошибок** - кэш не должен ломать основное приложение
5. **Тестирование производительности** - бенчмарки для оптимизации

Следуя этим принципам, вы сможете эффективно использовать кэширование в своих Go-приложениях.