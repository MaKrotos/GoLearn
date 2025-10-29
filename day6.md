# День 6: Системное проектирование (8 часов)

## Шаблон для ответов на вопросы системного дизайна

На собеседованиях по системному дизайну важно структурированно подходить к решению задач. Вот шаблон, который поможет вам организовать свои мысли:

### 1. Уточнение требований
Перед началом проектирования уточните требования:
- Какой ожидаемый объем трафика?
- Какие функциональные требования?
- Какие нефункциональные требования (доступность, задержки, согласованность)?
- Какие ограничения по ресурсам?

### 2. Оценка масштаба
Оцените объем данных и нагрузку:
- Количество запросов в секунду (RPS)
- Объем данных для хранения
- Пропускная способность сети

### 3. Высокоуровневый дизайн
Создайте архитектурную схему с основными компонентами.

### 4. Детализация ключевых компонентов
Рассмотрите детали реализации ключевых частей системы.

### 5. Оптимизации и улучшения
Предложите способы улучшения производительности и масштабируемости.

## Примеры ответов

### "Разбил бы на микросервисы: auth, user, order"

При проектировании системы с несколькими доменами логично разделить функциональность на микросервисы:

#### Микросервис аутентификации (auth)
- Отвечает за регистрацию, вход, управление сессиями
- Выдает JWT токены
- Хранит хэши паролей

#### Микросервис пользователей (user)
- Управляет профилями пользователей
- Хранит пользовательские данные
- Обрабатывает обновления профилей

#### Микросервис заказов (order)
- Управляет созданием и обработкой заказов
- Интегрируется с платежными системами
- Отслеживает статус заказов

#### Преимущества микросервисной архитектуры:
- Независимая разработка и развертывание
- Лучшая масштабируемость
- Изоляция сбоев
- Технологическая гибкость

### "Добавил кэш Redis для горячих данных"

Кэширование - ключевой способ улучшения производительности:

#### Уровни кэширования:
1. **Клиентский кэш** - браузер кэширует статические ресурсы
2. **CDN** - кэширование контента на граничных серверах
3. **Серверный кэш** - Redis/Memcached для часто запрашиваемых данных
4. **База данных** - встроенные механизмы кэширования

#### Пример использования Redis:
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

### "Настроил балансировщик нагрузки"

Балансировщик нагрузки распределяет трафик между несколькими серверами:

#### Типы балансировщиков:
1. **Round Robin** - поочередно направляет запросы
2. **Least Connections** - направляет запросы на сервер с наименьшим количеством активных соединений
3. **IP Hash** - направляет запросы на один и тот же сервер на основе IP клиента

#### Пример конфигурации NGINX:
```nginx
upstream backend {
    server backend1.example.com;
    server backend2.example.com;
    server backend3.example.com;
}

server {
    listen 80;
    
    location / {
        proxy_pass http://backend;
    }
}
```

### "Добавил мониторинг через Prometheus"

Мониторинг критически важен для production систем:

#### Основные метрики:
1. **Системные метрики** - CPU, память, диск, сеть
2. **Прикладные метрики** - RPS, время отклика, коды ошибок
3. **Бизнес-метрики** - количество пользователей, конверсии

#### Пример инструментации в Go:
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "endpoint"},
    )
)

func init() {
    prometheus.MustRegister(httpRequestsTotal)
    prometheus.MustRegister(httpRequestDuration)
}

func handler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    
    // Обработка запроса
    // ...
    
    // Запись метрик
    httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
    httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
}
```

### "Использовал бы message queue для асинхронных задач"

Очереди сообщений позволяют обрабатывать задачи асинхронно:

#### Сценарии использования:
1. **Отправка email** - не блокирует основной поток
2. **Обработка изображений** - ресурсоемкие задачи
3. **Интеграция с внешними сервисами** - обработка в фоне

#### Пример с RabbitMQ:
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

## Практические задания

1. Спроектируйте систему сокращения URL (tinyurl)
2. Спроектируйте систему чатов
3. Спроектируйте систему социальной сети
4. Спроектируйте систему обработки платежей
5. Спроектируйте систему рекомендаций

## Полезные ресурсы

1. **Книги**:
   - "Designing Data-Intensive Applications" Мартина Клеппмана
   - "System Design Interview" Алекса Сюя

2. **Онлайн ресурсы**:
   - System Design Primer на GitHub
   - High Scalability блог
   - AWS Architecture Center

3. **Инструменты**:
   - Draw.io для создания диаграмм
   - Prometheus и Grafana для мониторинга
   - Kafka для очередей сообщений