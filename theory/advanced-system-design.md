# Продвинутые концепции системного дизайна

## 1. Распределенные системы - углубленное изучение

### CAP теорема и её применение

#### Понимание компромиссов:
```
CAP теорема утверждает, что в распределенной системе данных можно одновременно обеспечить только два из трех свойств:

1. Consistency (Согласованность) - все узлы видят одни и те же данные в один момент времени
2. Availability (Доступность) - система продолжает отвечать на запросы даже при отказе узлов
3. Partition tolerance (Устойчивость к разделению) - система продолжает работать despite сетевых разделений

В реальных системах Partition tolerance практически всегда обязательна, поэтому приходится выбирать между Consistency и Availability.
```

#### PACELC теорема:
```
PACELC расширяет CAP теорему:

- Если (P)artition, то выбираем между (A)vailability и (C)onsistency
- Иначе (E), когда система работает нормально, выбираем между (L)atency и (C)onsistency

Это помогает понять, что даже в отсутствие сетевых разделений нужно делать компромиссы между производительностью и согласованностью.
```

### Типы распределенных систем

#### Stateless vs Stateful:
```go
// Stateless сервис - не хранит состояние между запросами
type StatelessUserService struct {
    db UserRepository
}

func (s *StatelessUserService) GetUser(id string) (*User, error) {
    // Все данные получаются из внешнего хранилища
    return s.db.FindByID(id)
}

// Stateful сервис - хранит состояние между запросами
type StatefulSessionService struct {
    sessions map[string]*Session // Внутреннее состояние
    mu       sync.RWMutex
}

func (s *StatefulSessionService) CreateSession(userID string) string {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    sessionID := generateSessionID()
    s.sessions[sessionID] = &Session{
        UserID:    userID,
        CreatedAt: time.Now(),
    }
    
    return sessionID
}

func (s *StatefulSessionService) GetSession(sessionID string) (*Session, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    session, exists := s.sessions[sessionID]
    if !exists {
        return nil, errors.New("session not found")
    }
    
    return session, nil
}
```

#### Event-driven архитектура:
```go
// Event событие в системе
type Event struct {
    ID        string
    Type      string
    Payload   interface{}
    Timestamp time.Time
}

// EventBus шина событий
type EventBus struct {
    subscribers map[string][]chan Event
    mu          sync.RWMutex
}

func NewEventBus() *EventBus {
    return &EventBus{
        subscribers: make(map[string][]chan Event),
    }
}

func (eb *EventBus) Subscribe(eventType string, ch chan Event) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
}

func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    
    if subscribers, exists := eb.subscribers[event.Type]; exists {
        for _, ch := range subscribers {
            select {
            case ch <- event:
            default:
                // Не блокируем, если канал полон
                log.Printf("Event channel is full for event type: %s", event.Type)
            }
        }
    }
}

// EventHandler обработчик событий
type EventHandler struct {
    eventType string
    handler   func(Event) error
}

func NewEventHandler(eventType string, handler func(Event) error) *EventHandler {
    return &EventHandler{
        eventType: eventType,
        handler:   handler,
    }
}

func (eh *EventHandler) Handle(event Event) error {
    if event.Type != eh.eventType {
        return fmt.Errorf("unexpected event type: %s", event.Type)
    }
    
    return eh.handler(event)
}
## 2. Масштабирование - углубленное изучение

### Горизонтальное vs Вертикальное масштабирование

#### Горизонтальное масштабирование с шардингом:
```go
// ShardManager менеджер шардов
type ShardManager struct {
    shards []Shard
    mu     sync.RWMutex
}

// Shard интерфейс шарда
type Shard interface {
    ID() string
    Store(key string, value interface{}) error
    Retrieve(key string) (interface{}, error)
    Delete(key string) error
}

// ConsistentHashShardManager менеджер шардов с консистентным хешированием
type ConsistentHashShardManager struct {
    hashRing *consistenthash.Map
    shards   map[string]Shard
    mu       sync.RWMutex
}

func NewConsistentHashShardManager() *ConsistentHashShardManager {
    return &ConsistentHashShardManager{
        hashRing: consistenthash.New(100, nil), // 100 виртуальных узлов
        shards:   make(map[string]Shard),
    }
}

func (sm *ConsistentHashShardManager) AddShard(shard Shard) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    shardID := shard.ID()
    sm.shards[shardID] = shard
    sm.hashRing.Add(shardID)
}

func (sm *ConsistentHashShardManager) GetShard(key string) (Shard, error) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    if len(sm.shards) == 0 {
        return nil, errors.New("no shards available")
    }
    
    shardID := sm.hashRing.Get(key)
    shard, exists := sm.shards[shardID]
    if !exists {
        return nil, fmt.Errorf("shard not found: %s", shardID)
    }
    
    return shard, nil
}

func (sm *ConsistentHashShardManager) Store(key string, value interface{}) error {
    shard, err := sm.GetShard(key)
    if err != nil {
        return err
    }
    
    return shard.Store(key, value)
}

func (sm *ConsistentHashShardManager) Retrieve(key string) (interface{}, error) {
    shard, err := sm.GetShard(key)
    if err != nil {
        return nil, err
    }
    
    return shard.Retrieve(key)
}
```

#### Вертикальное масштабирование с пулингом:
```go
// ConnectionPool пул соединений
type ConnectionPool struct {
    factory   func() (Connection, error)
    pool      chan Connection
    maxConns  int
    mu        sync.Mutex
    created   int
}

// Connection интерфейс соединения
type Connection interface {
    Execute(query string, args ...interface{}) (Result, error)
    Close() error
    IsAlive() bool
}

func NewConnectionPool(factory func() (Connection, error), maxConns int) *ConnectionPool {
    return &ConnectionPool{
        factory:  factory,
        pool:     make(chan Connection, maxConns),
        maxConns: maxConns,
    }
}

func (cp *ConnectionPool) Get() (Connection, error) {
    select {
    case conn := <-cp.pool:
        // Проверяем, что соединение живое
        if conn.IsAlive() {
            return conn, nil
        }
        // Если соединение мертво, закрываем его и создаем новое
        conn.Close()
        cp.mu.Lock()
        cp.created--
        cp.mu.Unlock()
    default:
    }
    
    // Создаем новое соединение, если пул пуст
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    if cp.created >= cp.maxConns {
        return nil, errors.New("connection pool exhausted")
    }
    
    conn, err := cp.factory()
    if err != nil {
        return nil, err
    }
    
    cp.created++
    return conn, nil
}

func (cp *ConnectionPool) Put(conn Connection) {
    select {
    case cp.pool <- conn:
    default:
        // Если пул полон, закрываем соединение
        conn.Close()
        cp.mu.Lock()
        cp.created--
        cp.mu.Unlock()
    }
}

func (cp *ConnectionPool) Close() {
    cp.mu.Lock()
    defer cp.mu.Unlock()
    
    close(cp.pool)
    for conn := range cp.pool {
        conn.Close()
    }
    cp.created = 0
}
```
### Load Balancing алгоритмы

#### Расширенные алгоритмы балансировки:
```go
// LoadBalancer балансировщик нагрузки
type LoadBalancer interface {
    Select(nodes []Node) Node
}

// Node узел в системе
type Node struct {
    ID       string
    Address  string
    Weight   int
    Healthy  bool
    Load     int64 // Текущая нагрузка
}

// RoundRobinLoadBalancer балансировщик по кругу
type RoundRobinLoadBalancer struct {
    current int64
}

func (lb *RoundRobinLoadBalancer) Select(nodes []Node) Node {
    var healthyNodes []Node
    for _, node := range nodes {
        if node.Healthy {
            healthyNodes = append(healthyNodes, node)
        }
    }
    
    if len(healthyNodes) == 0 {
        return Node{} // Нет доступных узлов
    }
    
    index := atomic.AddInt64(&lb.current, 1) % int64(len(healthyNodes))
    return healthyNodes[index]
}

// WeightedRoundRobinLoadBalancer взвешенный балансировщик по кругу
type WeightedRoundRobinLoadBalancer struct {
    current int64
}

func (lb *WeightedRoundRobinLoadBalancer) Select(nodes []Node) Node {
    var healthyNodes []Node
    totalWeight := 0
    
    for _, node := range nodes {
        if node.Healthy {
            healthyNodes = append(healthyNodes, node)
            totalWeight += node.Weight
        }
    }
    
    if len(healthyNodes) == 0 {
        return Node{}
    }
    
    if totalWeight == 0 {
        // Если веса не заданы, используем обычный round-robin
        index := atomic.AddInt64(&lb.current, 1) % int64(len(healthyNodes))
        return healthyNodes[index]
    }
    
    // Выбираем узел на основе весов
    point := atomic.AddInt64(&lb.current, 1) % int64(totalWeight)
    current := int64(0)
    
    for _, node := range healthyNodes {
        current += int64(node.Weight)
        if current > point {
            return node
        }
    }
    
    return healthyNodes[0]
}

// LeastConnectionsLoadBalancer балансировщик по наименьшему количеству соединений
type LeastConnectionsLoadBalancer struct{}

func (lb *LeastConnectionsLoadBalancer) Select(nodes []Node) Node {
    var healthyNodes []Node
    for _, node := range nodes {
        if node.Healthy {
            healthyNodes = append(healthyNodes, node)
        }
    }
    
    if len(healthyNodes) == 0 {
        return Node{}
    }
    
    minNode := healthyNodes[0]
    for _, node := range healthyNodes[1:] {
        if node.Load < minNode.Load {
            minNode = node
        }
    }
    
    return minNode
}
```
## 3. Кэширование - углубленное изучение

### Распределенное кэширование

#### Redis Cluster клиент:
```go
// RedisClusterClient клиент для Redis Cluster
type RedisClusterClient struct {
    clients map[string]*redis.Client // Клиенты для каждого шарда
    slots   [16384]string            // Распределение слотов по узлам
    mu      sync.RWMutex
}

func NewRedisClusterClient(nodes []string) (*RedisClusterClient, error) {
    client := &RedisClusterClient{
        clients: make(map[string]*redis.Client),
        slots:   [16384]string{},
    }
    
    // Инициализируем клиентов для каждого узла
    for _, node := range nodes {
        client.clients[node] = redis.NewClient(&redis.Options{
            Addr: node,
        })
    }
    
    // Инициализируем распределение слотов
    if err := client.initializeSlots(); err != nil {
        return nil, err
    }
    
    return client, nil
}

func (c *RedisClusterClient) initializeSlots() error {
    // Получаем информацию о слотах от одного из узлов
    for _, client := range c.clients {
        result, err := client.ClusterSlots().Result()
        if err != nil {
            continue
        }
        
        // Заполняем таблицу слотов
        for _, slotInfo := range result {
            start, end := slotInfo.Start, slotInfo.End
            masterNode := slotInfo.Nodes[0].Addr
            
            for i := start; i <= end; i++ {
                c.slots[i] = masterNode
            }
        }
        
        return nil
    }
    
    return errors.New("failed to initialize cluster slots")
}

func (c *RedisClusterClient) getKeySlot(key string) int {
    // Вычисляем слот для ключа по хешу
    hash := crc16.Checksum([]byte(key), crc16.MakeTable(crc16.CRC16CCITTFalse))
    return int(hash) % 16384
}

func (c *RedisClusterClient) getClientForKey(key string) (*redis.Client, error) {
    slot := c.getKeySlot(key)
    
    c.mu.RLock()
    nodeAddr, exists := c.slots[slot]
    c.mu.RUnlock()
    
    if !exists {
        return nil, fmt.Errorf("no node found for slot %d", slot)
    }
    
    client, exists := c.clients[nodeAddr]
    if !exists {
        return nil, fmt.Errorf("client not found for node %s", nodeAddr)
    }
    
    return client, nil
}

func (c *RedisClusterClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    client, err := c.getClientForKey(key)
    if err != nil {
        return err
    }
    
    return client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisClusterClient) Get(ctx context.Context, key string) (string, error) {
    client, err := c.getClientForKey(key)
    if err != nil {
        return "", err
    }
    
    return client.Get(ctx, key).Result()
}
```

#### Cache-Aside с инвалидацией:
```go
// CacheWithInvalidation кэш с поддержкой инвалидации
type CacheWithInvalidation struct {
    cache    Cache
    invalidator CacheInvalidator
}

// CacheInvalidator интерфейс для инвалидации кэша
type CacheInvalidator interface {
    InvalidatePattern(pattern string) error
    InvalidateTag(tag string) error
}

// RedisCacheInvalidator реализация инвалидации через Redis
type RedisCacheInvalidator struct {
    client *redis.Client
}

func (ci *RedisCacheInvalidator) InvalidatePattern(pattern string) error {
    // Используем SCAN для поиска ключей по паттерну
    iter := ci.client.Scan(0, pattern, 0).Iterator()
    var keys []string
    
    for iter.Next() {
        keys = append(keys, iter.Val())
    }
    
    if err := iter.Err(); err != nil {
        return err
    }
    
    if len(keys) > 0 {
        return ci.client.Del(context.Background(), keys...).Err()
    }
    
    return nil
}

func (ci *RedisCacheInvalidator) InvalidateTag(tag string) error {
    // Получаем все ключи с тегом
    keys, err := ci.client.SMembers(context.Background(), "tag:"+tag).Result()
    if err != nil {
        return err
    }
    
    if len(keys) > 0 {
        // Удаляем ключи
        if err := ci.client.Del(context.Background(), keys...).Err(); err != nil {
            return err
        }
        
        // Удаляем тег
        return ci.client.Del(context.Background(), "tag:"+tag).Err()
    }
    
    return nil
}

// TaggedCache кэш с поддержкой тегов
type TaggedCache struct {
    cache Cache
    tags  map[string][]string // tag -> keys
    mu    sync.RWMutex
}

func (tc *TaggedCache) Set(key string, value interface{}, tags []string) error {
    if err := tc.cache.Set(key, value); err != nil {
        return err
    }
    
    tc.mu.Lock()
    defer tc.mu.Unlock()
    
    // Добавляем ключ к каждому тегу
    for _, tag := range tags {
        tc.tags[tag] = append(tc.tags[tag], key)
    }
    
    return nil
}

func (tc *TaggedCache) InvalidateTag(tag string) error {
    tc.mu.RLock()
    keys, exists := tc.tags[tag]
    tc.mu.RUnlock()
    
    if !exists {
        return nil
    }
    
    // Удаляем все ключи с тегом
    for _, key := range keys {
        tc.cache.Delete(key)
    }
    
    // Удаляем тег
    tc.mu.Lock()
    delete(tc.tags, tag)
    tc.mu.Unlock()
    
    return nil
}
```
## 4. Мониторинг и логирование - углубленное изучение

### Распределенная трассировка

#### OpenTelemetry интеграция:
```go
// TracedService сервис с трассировкой
type TracedService struct {
    tracer trace.Tracer
}

func NewTracedService() *TracedService {
    // Инициализируем OpenTelemetry
    tp := initTracer()
    otel.SetTracerProvider(tp)
    
    return &TracedService{
        tracer: tp.Tracer("service-name"),
    }
}

func (s *TracedService) ProcessOrder(ctx context.Context, order Order) error {
    // Создаем span для отслеживания операции
    ctx, span := s.tracer.Start(ctx, "ProcessOrder")
    defer span.End()
    
    // Добавляем атрибуты к span
    span.SetAttributes(
        attribute.String("order.id", order.ID),
        attribute.String("order.user_id", order.UserID),
        attribute.Float64("order.amount", order.Amount),
    )
    
    // Обрабатываем заказ
    if err := s.validateOrder(ctx, order); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    if err := s.chargePayment(ctx, order); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }
    
    return nil
}

func (s *TracedService) validateOrder(ctx context.Context, order Order) error {
    _, span := s.tracer.Start(ctx, "validateOrder")
    defer span.End()
    
    // Имитация валидации
    time.Sleep(50 * time.Millisecond)
    
    if order.Amount <= 0 {
        return errors.New("invalid order amount")
    }
    
    return nil
}

func (s *TracedService) chargePayment(ctx context.Context, order Order) error {
    ctx, span := s.tracer.Start(ctx, "chargePayment")
    defer span.End()
    
    // Имитация вызова платежной системы
    time.Sleep(100 * time.Millisecond)
    
    // Добавляем событие в span
    span.AddEvent("Payment charged", trace.WithAttributes(
        attribute.String("payment.id", generatePaymentID()),
    ))
    
    return nil
}
```

### Метрики и алертинг

#### Prometheus интеграция:
```go
// Метрики для мониторинга
var (
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "Duration of HTTP requests",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
    
    httpRequestCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_request_count_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    activeConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_connections",
            Help: "Number of active connections",
        },
    )
)

func init() {
    // Регистрируем метрики
    prometheus.MustRegister(httpRequestDuration)
    prometheus.MustRegister(httpRequestCount)
    prometheus.MustRegister(activeConnections)
}

// Middleware для сбора метрик
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Оборачиваем ResponseWriter для получения кода статуса
        wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        // Увеличиваем счетчик активных соединений
        activeConnections.Inc()
        defer activeConnections.Dec()
        
        // Обрабатываем запрос
        next.ServeHTTP(wrapped, r)
        
        // Засекаем время выполнения
        duration := time.Since(start).Seconds()
        
        // Обновляем метрики
        httpRequestDuration.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.statusCode),
        ).Observe(duration)
        
        httpRequestCount.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(wrapped.statusCode),
        ).Inc()
    })
}

// Обертка для ResponseWriter
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}
```
## 5. Отказоустойчивость и восстановление

### Circuit Breaker паттерн

#### Расширенная реализация:
```go
// CircuitBreakerState состояние circuit breaker
type CircuitBreakerState int

const (
    Closed CircuitBreakerState = iota
    Open
    HalfOpen
)

// CircuitBreaker реализация паттерна
type CircuitBreaker struct {
    name          string
    state         CircuitBreakerState
    failureCount  int
    successCount  int
    lastFailure   time.Time
    timeout       time.Duration
    failureThreshold int
    successThreshold int
    mu            sync.RWMutex
}

func NewCircuitBreaker(name string, timeout time.Duration, failureThreshold, successThreshold int) *CircuitBreaker {
    return &CircuitBreaker{
        name:             name,
        state:            Closed,
        timeout:          timeout,
        failureThreshold: failureThreshold,
        successThreshold: successThreshold,
    }
}

func (cb *CircuitBreaker) Call(ctx context.Context, fn func() error) error {
    cb.mu.Lock()
    defer cb.mu.Unlock()
    
    switch cb.state {
    case Open:
        // Проверяем, можно ли перейти в HalfOpen
        if time.Since(cb.lastFailure) >= cb.timeout {
            cb.state = HalfOpen
            cb.successCount = 0
        } else {
            return fmt.Errorf("circuit breaker %s is open", cb.name)
        }
        
    case HalfOpen:
        // В HalfOpen состоянии разрешаем только один запрос
        fallthrough
        
    case Closed:
        // Выполняем запрос
        err := fn()
        
        if err != nil {
            cb.onFailure()
            return err
        } else {
            cb.onSuccess()
            return nil
        }
    }
    
    return fmt.Errorf("unexpected circuit breaker state")
}

func (cb *CircuitBreaker) onFailure() {
    cb.failureCount++
    cb.lastFailure = time.Now()
    
    if cb.state == HalfOpen || cb.failureCount >= cb.failureThreshold {
        cb.state = Open
        log.Printf("Circuit breaker %s opened", cb.name)
    }
}

func (cb *CircuitBreaker) onSuccess() {
    cb.failureCount = 0
    
    if cb.state == HalfOpen {
        cb.successCount++
        if cb.successCount >= cb.successThreshold {
            cb.state = Closed
            cb.successCount = 0
            log.Printf("Circuit breaker %s closed", cb.name)
        }
    }
}

// Пример использования с HTTP клиентом
type ResilientHTTPClient struct {
    client        *http.Client
    circuitBreaker *CircuitBreaker
}

func NewResilientHTTPClient() *ResilientHTTPClient {
    return &ResilientHTTPClient{
        client: &http.Client{
            Timeout: 5 * time.Second,
        },
        circuitBreaker: NewCircuitBreaker("http-client", 30*time.Second, 5, 3),
    }
}

func (c *ResilientHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
    var resp *http.Response
    var err error
    
    callErr := c.circuitBreaker.Call(ctx, func() error {
        req, reqErr := http.NewRequestWithContext(ctx, "GET", url, nil)
        if reqErr != nil {
            return reqErr
        }
        
        resp, err = c.client.Do(req)
        return err
    })
    
    if callErr != nil {
        return nil, callErr
    }
    
    return resp, nil
}
```

### Retry паттерн с экспоненциальной задержкой

#### Расширенная реализация:
```go
// RetryPolicy политика повторных попыток
type RetryPolicy struct {
    MaxRetries      int
    InitialDelay    time.Duration
    MaxDelay        time.Duration
    Multiplier      float64
    Jitter          bool
    RetryableErrors []error
}

// DefaultRetryPolicy стандартная политика повторов
func DefaultRetryPolicy() RetryPolicy {
    return RetryPolicy{
        MaxRetries:   3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     5 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
    }
}

// ShouldRetry определяет, нужно ли повторять операцию
func (rp RetryPolicy) ShouldRetry(err error, attempt int) bool {
    if attempt >= rp.MaxRetries {
        return false
    }
    
    // Если список ошибок не задан, повторяем для всех ошибок
    if len(rp.RetryableErrors) == 0 {
        return true
    }
    
    // Проверяем, является ли ошибка повторяемой
    for _, retryableErr := range rp.RetryableErrors {
        if errors.Is(err, retryableErr) {
            return true
        }
    }
    
    return false
}

// CalculateDelay вычисляет задержку перед следующей попыткой
func (rp RetryPolicy) CalculateDelay(attempt int) time.Duration {
    delay := time.Duration(float64(rp.InitialDelay) * math.Pow(rp.Multiplier, float64(attempt)))
    
    // Ограничиваем максимальной задержкой
    if delay > rp.MaxDelay {
        delay = rp.MaxDelay
    }
    
    // Добавляем jitter для избежания thundering herd
    if rp.Jitter {
        jitter := time.Duration(rand.Int63n(int64(delay) / 10))
        delay += jitter
    }
    
    return delay
}

// RetryWithPolicy выполняет операцию с заданной политикой повторов
func RetryWithPolicy(ctx context.Context, policy RetryPolicy, fn func() error) error {
    var lastErr error
    
    for attempt := 0; ; attempt++ {
        // Выполняем операцию
        err := fn()
        if err == nil {
            // Успех
            return nil
        }
        
        lastErr = err
        
        // Проверяем, нужно ли повторять
        if !policy.ShouldRetry(err, attempt) {
            return err
        }
        
        // Вычисляем задержку
        delay := policy.CalculateDelay(attempt)
        
        // Ждем с учетом контекста
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            // Продолжаем повторять
        }
    }
}

// Пример использования
func fetchUserData(userID string) (*User, error) {
    policy := RetryPolicy{
        MaxRetries:   3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     1 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
        RetryableErrors: []error{
            context.DeadlineExceeded,
            io.ErrUnexpectedEOF,
        },
    }
    
    var user *User
    
    err := RetryWithPolicy(context.Background(), policy, func() error {
        var err error
        user, err = api.GetUser(userID)
        return err
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to fetch user data after retries: %w", err)
    }
    
    return user, nil
}
```
## 6. Безопасность в распределенных системах

### Аутентификация и авторизация

#### JWT с отзывом токенов:
```go
// JWTManager менеджер JWT токенов
type JWTManager struct {
    secretKey     []byte
    refreshTokenStore RefreshTokenStore
}

// RefreshTokenStore хранилище refresh токенов
type RefreshTokenStore interface {
    Store(userID, tokenID string, expiresAt time.Time) error
    Get(userID, tokenID string) (time.Time, error)
    Delete(userID, tokenID string) error
}

// Claims кастомные claims для JWT
type Claims struct {
    UserID   string `json:"user_id"`
    TokenID  string `json:"token_id"`
    Role     string `json:"role"`
    jwt.StandardClaims
}

// GenerateTokens генерирует access и refresh токены
func (jm *JWTManager) GenerateTokens(userID, role string) (string, string, error) {
    // Генерируем access token
    accessTokenExp := time.Now().Add(15 * time.Minute).Unix()
    accessClaims := Claims{
        UserID: userID,
        Role:   role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: accessTokenExp,
            IssuedAt:  time.Now().Unix(),
        },
    }
    
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString(jm.secretKey)
    if err != nil {
        return "", "", err
    }
    
    // Генерируем refresh token
    refreshTokenExp := time.Now().Add(7 * 24 * time.Hour).Unix()
    tokenID := generateTokenID()
    
    refreshClaims := Claims{
        UserID:  userID,
        TokenID: tokenID,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: refreshTokenExp,
            IssuedAt:  time.Now().Unix(),
        },
    }
    
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString(jm.secretKey)
    if err != nil {
        return "", "", err
    }
    
    // Сохраняем refresh token в хранилище
    err = jm.refreshTokenStore.Store(userID, tokenID, time.Unix(refreshTokenExp, 0))
    if err != nil {
        return "", "", err
    }
    
    return accessTokenString, refreshTokenString, nil
}

// RefreshTokens обновляет токены
func (jm *JWTManager) RefreshTokens(refreshTokenString string) (string, string, error) {
    // Парсим refresh token
    refreshClaims := &Claims{}
    token, err := jwt.ParseWithClaims(refreshTokenString, refreshClaims, func(token *jwt.Token) (interface{}, error) {
        return jm.secretKey, nil
    })
    
    if err != nil || !token.Valid {
        return "", "", errors.New("invalid refresh token")
    }
    
    // Проверяем, существует ли refresh token в хранилище
    expiresAt, err := jm.refreshTokenStore.Get(refreshClaims.UserID, refreshClaims.TokenID)
    if err != nil {
        return "", "", errors.New("refresh token not found")
    }
    
    // Проверяем, не истек ли refresh token
    if time.Now().After(expiresAt) {
        // Удаляем просроченный токен
        jm.refreshTokenStore.Delete(refreshClaims.UserID, refreshClaims.TokenID)
        return "", "", errors.New("refresh token expired")
    }
    
    // Удаляем старый refresh token
    jm.refreshTokenStore.Delete(refreshClaims.UserID, refreshClaims.TokenID)
    
    // Генерируем новые токены
    return jm.GenerateTokens(refreshClaims.UserID, refreshClaims.Role)
}
```

### Шифрование и хэширование

#### Безопасное хранение паролей:
```go
// PasswordHasher хэширует пароли
type PasswordHasher struct {
    cost int // bcrypt cost
}

func NewPasswordHasher(cost int) *PasswordHasher {
    return &PasswordHasher{cost: cost}
}

// HashPassword хэширует пароль
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), ph.cost)
    if err != nil {
        return "", err
    }
    
    return string(hashed), nil
}

// ComparePassword сравнивает пароль с хэшем
func (ph *PasswordHasher) ComparePassword(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// RateLimiter ограничивает количество попыток
type RateLimiter struct {
    store map[string]*AttemptInfo
    mu    sync.RWMutex
    maxAttempts int
    window      time.Duration
}

type AttemptInfo struct {
    attempts int
    lastAttempt time.Time
}

func NewRateLimiter(maxAttempts int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        store:       make(map[string]*AttemptInfo),
        maxAttempts: maxAttempts,
        window:      window,
    }
}

func (rl *RateLimiter) IsAllowed(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    info, exists := rl.store[key]
    if !exists {
        rl.store[key] = &AttemptInfo{
            attempts:    1,
            lastAttempt: time.Now(),
        }
        return true
    }
    
    // Проверяем, прошло ли достаточно времени с последней попытки
    if time.Since(info.lastAttempt) > rl.window {
        // Сбрасываем счетчик
        info.attempts = 1
        info.lastAttempt = time.Now()
        return true
    }
    
    // Проверяем, не превышен ли лимит
    if info.attempts >= rl.maxAttempts {
        return false
    }
    
    // Увеличиваем счетчик
    info.attempts++
    info.lastAttempt = time.Now()
    return true
}

func (rl *RateLimiter) Reset(key string) {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    delete(rl.store, key)
}
```
## 7. Заключение

### Рекомендации по проектированию распределенных систем

1. **Начинайте с простого** - не пытайтесь сразу реализовать сложную распределенную систему. Начните с монолита и постепенно переходите к микросервисам по мере необходимости.

2. **Планируйте отказоустойчивость с самого начала** - учитывайте возможные точки отказа и реализуйте механизмы восстановления.

3. **Используйте правильные инструменты для правильных задач** - выбирайте технологии и паттерны, которые подходят для ваших конкретных требований.

4. **Мониторинг и логирование - неотъемлемая часть** - реализуйте комплексный мониторинг с самого начала разработки.

5. **Тестируйте отказоустойчивость** - регулярно проводите тестирование на устойчивость к сбоям и нагрузочное тестирование.

### Дополнительные ресурсы

- [Designing Data-Intensive Applications](https://dataintensive.netlify.app/) - отличная книга по системному дизайну
- [Google Cloud Architecture Framework](https://cloud.google.com/architecture/framework) - рекомендации по архитектуре от Google
- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/) - подходы к проектированию от Amazon
- [Microsoft Azure Architecture Center](https://docs.microsoft.com/en-us/azure/architecture/) - руководства по архитектуре от Microsoft

### Практические упражнения

1. **Разработайте систему обработки платежей** - реализуйте систему с высокой доступностью, отказоустойчивостью и строгой согласованностью.

2. **Создайте распределенный кэш** - реализуйте кэш с поддержкой репликации, шардинга и инвалидации.

3. **Разработайте систему уведомлений** - создайте масштабируемую систему для отправки уведомлений по различным каналам.

4. **Постройте систему аналитики** - реализуйте систему для сбора, обработки и анализа больших объемов данных.

Эти упражнения помогут закрепить знания и получить практический опыт в проектировании распределенных систем.