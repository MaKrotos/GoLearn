# Системное проектирование - объяснение для чайников

## Что такое системное проектирование?

Представьте, что вы архитектор, который проектирует **город**. Вам нужно подумать о:
- **Дорогах** - как люди будут перемещаться?
- **Зданиях** - где живут, где работают, где отдыхают?
- **Коммуникациях** - вода, электричество, интернет?
- **Безопасности** - как защитить город?

Системное проектирование в IT - это **проектирование сложных систем**, которые:
- Обрабатывают **много пользователей**
- Работают **надежно**
- Масштабируются **под нагрузку**

## Базовые принципы

### 1. CAP теорема

Вы можете выбрать только **два** из трех:
- **Consistency** (Согласованность) - все видят одни и те же данные
- **Availability** (Доступность) - система всегда отвечает
- **Partition tolerance** (Устойчивость к разделению) - система работает при сетевых проблемах

```
        ┌─────────────┐
        │  Согласованность  │
        └──────┬──────┘
               │
┌─────────────┐│┌─────────────┐
│ Доступность │└│ Устойчивость │
└─────────────┘ └─────────────┘
```

### 2. PACELC теорема

Расширение CAP:
- **Если** сетевое разделение (P), **то** выбираем A или C
- **Иначе** (E), выбираем **Latency** (L) или **Consistency** (C)

## Компоненты распределенных систем

### 1. Load Balancer (Балансировщик нагрузки)

**Назначение**: Распределяет запросы между серверами

```
Клиенты → [Load Balancer] → Сервер 1
                    ├────→ Сервер 2
                    └────→ Сервер 3
```

**Типы**:
- **Round Robin** - по очереди
- **Least Connections** - наименьшее количество соединений
- **IP Hash** - по IP адресу клиента

### 2. Reverse Proxy

**Назначение**: Промежуточный сервер между клиентом и backend

**Преимущества**:
- **Кэширование**
- **SSL termination**
- **Компрессия**
- **Безопасность**

### 3. CDN (Content Delivery Network)

**Назначение**: Доставка контента с ближайшего сервера

```
Пользователь в Москве → CDN сервер в Москве → Быстрая загрузка
Пользователь в Владивостоке → CDN сервер во Владивостоке → Быстрая загрузка
```

## Хранение данных

### 1. Репликация vs Шардинг

#### Репликация (Replication)
```
Master DB ──┐
            ├── Replica 1
            ├── Replica 2
            └── Replica 3
```

**Плюсы**:
- Высокая доступность
- Читать можно с любого replica

**Минусы**:
- Запись только в master
- Возможна рассинхронизация

#### Шардинг (Sharding)
```
User ID 1-1000  → Shard 1
User ID 1001-2000 → Shard 2
User ID 2001-3000 → Shard 3
```

**Плюсы**:
- Горизонтальное масштабирование
- Меньше нагрузки на каждый shard

**Минусы**:
- Сложность JOIN запросов
- Сложность перебалансировки

### 2. Типы баз данных

#### SQL (Реляционные)
- **PostgreSQL**, MySQL, Oracle
- **ACID** транзакции
- **JOIN** таблиц
- **Схема** данных

#### NoSQL
- **Документные**: MongoDB, CouchDB
- **Ключ-значение**: Redis, DynamoDB
- **Графовые**: Neo4j
- **Колоночные**: Cassandra, HBase

## Кэширование

### Уровни кэширования

```
Клиент → [Browser Cache] → [CDN] → [Application Cache] → [Database Cache] → База данных
```

### Redis/Memcached

**Использование**:
```go
// Получение данных с кэшированием
func GetUserProfile(userID string) (*User, error) {
    // Пытаемся получить из кэша
    cached, err := redisClient.Get(context.Background(), "user:"+userID).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }
    
    // Если нет в кэше, получаем из БД
    user, err := db.GetUserByID(userID)
    if err != nil {
        return nil, err
    }
    
    // Сохраняем в кэш на 1 час
    data, _ := json.Marshal(user)
    redisClient.Set(context.Background(), "user:"+userID, data, time.Hour)
    
    return user, nil
}
```

## Микросервисы

### Что такое микросервисы?

Вместо одного большого приложения - **много маленьких**:

```
Монолит:
[Все в одном приложении]

Микросервисы:
[Auth Service] [User Service] [Order Service] [Payment Service]
```

### Паттерны микросервисов

#### API Gateway
```
Клиенты → [API Gateway] → [Auth Service]
                    ├────→ [User Service]
                    ├────→ [Order Service]
                    └────→ [Payment Service]
```

#### Service Discovery
```
[Service Registry]
       ↑
[Auth Service] ──┐
[User Service]   ├── Регистрация
[Order Service]  │
                 ↓
           [API Gateway] → Обнаружение сервисов
```

## Асинхронная обработка

### Message Queue

```
Producer → [Message Queue] → Consumer
              ↑
         [Долговременное хранение]
```

**Преимущества**:
- **Развязка** producer и consumer
- **Буферизация** при пиковой нагрузке
- **Надежность** - сообщения не теряются

### Пример с RabbitMQ:

```go
// Producer
func sendEmail(email Email) error {
    body, _ := json.Marshal(email)
    return channel.Publish(
        "",           // exchange
        "emails",     // routing key
        false,        // mandatory
        false,        // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        body,
        })
}

// Consumer
func processEmails() {
    msgs, _ := channel.Consume(
        "emails", // queue
        "",       // consumer
        true,     // auto-ack
        false,    // exclusive
        false,    // no-local
        false,    // no-wait
        nil,      // args
    )
    
    for msg := range msgs {
        var email Email
        json.Unmarshal(msg.Body, &email)
        sendEmail(email)
    }
}
```

## Мониторинг и логирование

### Метрики

**Типы метрик**:
- **Счетчики** (Counters) - количество запросов
- **Гистограммы** (Histograms) - время ответа
- **Гауги** (Gauges) - текущее состояние (CPU, память)

### Логирование

**Уровни логирования**:
- **DEBUG** - для разработчиков
- **INFO** - важные события
- **WARN** - предупреждения
- **ERROR** - ошибки
- **FATAL** - критические ошибки

### Distributed Tracing

Отслеживание запроса через все сервисы:

```
Запрос: [Frontend] → [API Gateway] → [User Service] → [DB]
Trace ID: abc123
         ├─ Span 1
         ├─ Span 2
         └─ Span 3
```

## Практический пример: Система сокращения URL

### Требования

**Функциональные**:
- Создание короткого URL для длинного
- Переадресация по короткому URL

**Нефункциональные**:
- Низкая задержка (< 100ms)
- Высокая доступность (99.9%)
- Масштабируемость

### Оценка нагрузки

- **100M URL в день** = ~1150 RPS
- **Хранилище**: 100M * 100 bytes = 10GB данных в день

### Высокоуровневый дизайн

```
[Клиент] → [Load Balancer] → [API Gateway] → [URL Service] → [База данных]
                                              │
                                              ↓
                                        [Кэш (Redis)]
```

### Детали реализации

#### Генерация коротких ID

```go
// Base62 кодирование
func generateShortID(id int64) string {
    const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
    var shortID strings.Builder
    
    for id > 0 {
        shortID.WriteByte(base62[id%62])
        id /= 62
    }
    
    return shortID.String()
}
```

#### Кэширование

```go
func getLongURL(shortID string) (string, error) {
    // Проверяем кэш
    if url, err := redisClient.Get(context.Background(), shortID).Result(); err == nil {
        return url, nil
    }
    
    // Получаем из БД
    url, err := db.GetLongURL(shortID)
    if err != nil {
        return "", err
    }
    
    // Сохраняем в кэш
    redisClient.Set(context.Background(), shortID, url, 24*time.Hour)
    
    return url, nil
}
```

### Масштабирование

#### Горизонтальное масштабирование
- **Несколько инстансов** URL Service
- **Шардинг** базы данных
- **Кластер** Redis

#### Географическое распределение
- **CDN** для статических ресурсов
- **Региональные** дата-центры
- **DNS** routing

## Лучшие практики

### 1. Проектирование на будущее

```go
// Плохо - жестко закодированные значения
const MaxURLLength = 2048

// Лучше - конфигурируемо
var MaxURLLength = getEnvInt("MAX_URL_LENGTH", 2048)
```

### 2. Graceful degradation

```go
func getLongURL(shortID string) (string, error) {
    // Пытаемся получить из кэша
    if url, err := redisClient.Get(context.Background(), shortID).Result(); err == nil {
        return url, nil
    }
    
    // Если кэш недоступен, берем из БД
    url, err := db.GetLongURL(shortID)
    if err != nil {
        return "", err
    }
    
    // Асинхронно обновляем кэш
    go func() {
        redisClient.Set(context.Background(), shortID, url, 24*time.Hour)
    }()
    
    return url, nil
}
```

### 3. Circuit Breaker

```go
type CircuitBreaker struct {
    state    string // CLOSED, OPEN, HALF_OPEN
    failures int
    lastFail time.Time
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == "OPEN" && time.Since(cb.lastFail) < time.Minute {
        return errors.New("circuit breaker open")
    }
    
    err := fn()
    if err != nil {
        cb.failures++
        cb.lastFail = time.Now()
        if cb.failures > 5 {
            cb.state = "OPEN"
        }
        return err
    }
    
    cb.state = "CLOSED"
    cb.failures = 0
    return nil
}
```

## Распространенные ошибки

### 1. Переинжиниринг

Не пытайтесь сразу сделать "идеальную" систему. Начните с **MVP** и **итеративно** улучшайте.

### 2. Игнорирование мониторинга

```go
// Плохо - нет метрик
func handler(w http.ResponseWriter, r *http.Request) {
    processRequest(r)
    w.WriteHeader(http.StatusOK)
}

// Лучше - с метриками
func handler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    err := processRequest(r)
    
    duration := time.Since(start)
    requestDuration.Observe(duration.Seconds())
    
    if err != nil {
        requestErrors.Inc()
        http.Error(w, "Internal Error", http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
}
```

### 3. Недостаточное тестирование отказоустойчивости

Регулярно проводите **chaos engineering**:
- Отключайте сервисы
- Вводите сетевые задержки
- Тестируйте переполнение очередей

## См. также

- [Архитектура приложений](architecture.md) - как проектировать отдельные сервисы
- [HTTP серверы](http-server.md) - реализация API
- [Базы данных](database.md) - выбор и использование СУБД
- [Кэширование](../theory/caching.md) - подробнее о кэшах
- [Микросервисы](../theory/microservices.md) - паттерны проектирования