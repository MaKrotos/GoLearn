# Интерфейсы в Go: Полная теория

## Введение в интерфейсы

### Что такое интерфейс?

Интерфейс в Go - это **тип, который определяет набор сигнатур методов**. Любой тип, реализующий все методы интерфейса, **автоматически** реализует этот интерфейс.

### Зачем нужны интерфейсы?

Интерфейсы обеспечивают:
1. **Абстракцию** - скрывают детали реализации
2. **Полиморфизм** - возможность работать с разными типами одинаково
3. **Тестируемость** - возможность подменять реальные реализации на моки
4. **Гибкость** - уменьшают связность компонентов

## Основы интерфейсов

### Определение интерфейса

```go
// Определяем интерфейс
type Writer interface {
    Write([]byte) (int, error)
}

// Тип, реализующий интерфейс
type File struct{}

func (f File) Write(data []byte) (int, error) {
    fmt.Println("Запись в файл:", string(data))
    return len(data), nil
}

// Другой тип, реализующий тот же интерфейс
type NetworkConnection struct{}

func (nc NetworkConnection) Write(data []byte) (int, error) {
    fmt.Println("Отправка по сети:", string(data))
    return len(data), nil
}

// Функция, работающая с любым Writer
func saveData(writer Writer, data string) {
    writer.Write([]byte(data))
}

func main() {
    file := File{}
    network := NetworkConnection{}
    
    // Оба типа могут использоваться как Writer
    saveData(file, "Данные для файла")
    saveData(network, "Данные для сети")
}
```

### Неявная реализация

В Go интерфейсы реализуются **неявно** - нет необходимости явно указывать, что тип реализует интерфейс:

```go
// НЕ НУЖНО писать что-то вроде:
// type File implements Writer struct { ... }

// Достаточно просто реализовать методы:
type File struct{}

func (f File) Write(data []byte) (int, error) {
    // Реализация
    return len(data), nil
}

// File автоматически реализует Writer
```

## Стандартные интерфейсы

### 1. io.Reader

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Примеры реализаций:
```go
// strings.Reader
reader := strings.NewReader("Привет, мир!")
buffer := make([]byte, 5)
n, err := reader.Read(buffer)

// os.File
file, _ := os.Open("example.txt")
defer file.Close()
n, err := file.Read(buffer)

// bytes.Buffer
var buf bytes.Buffer
buf.WriteString("Данные")
n, err := buf.Read(buffer)
```

### 2. io.Writer

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Примеры реализаций:
```go
// os.Stdout
fmt.Fprint(os.Stdout, "Вывод в консоль")

// bytes.Buffer
var buf bytes.Buffer
buf.Write([]byte("Данные"))

// http.ResponseWriter
func handler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("HTTP ответ"))
}
```

### 3. fmt.Stringer

```go
type Stringer interface {
    String() string
}
```

Пример реализации:
```go
type Person struct {
    Name string
    Age  int
}

func (p Person) String() string {
    return fmt.Sprintf("%s (%d лет)", p.Name, p.Age)
}

func main() {
    person := Person{Name: "Иван", Age: 30}
    fmt.Println(person) // Автоматически вызывает String()
}
```

### 4. error

```go
type error interface {
    Error() string
}
```

Создание ошибок:
```go
// Простая ошибка
err := errors.New("что-то пошло не так")

// Ошибка с форматированием
err := fmt.Errorf("неверное значение: %d", value)

// Пользовательская ошибка
type ValidationError struct {
    Field string
    Value string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("поле %s имеет неверное значение: %s", e.Field, e.Value)
}
```

## Композиция интерфейсов

### Встраивание интерфейсов

```go
// Базовые интерфейсы
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

type Closer interface {
    Close() error
}

// Композиция интерфейсов
type ReadWriter interface {
    Reader
    Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}
```

### Пример использования композиции

```go
// os.File реализует ReadWriteCloser
file, err := os.OpenFile("example.txt", os.O_RDWR|os.O_CREATE, 0644)
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Может читать
buffer := make([]byte, 1024)
n, err := file.Read(buffer)

// Может писать
file.Write([]byte("Новые данные"))

// Может закрываться
file.Close()
```

## Интерфейсы и указатели

### Методы с получателями

```go
type Counter struct {
    value int
}

// Метод с получателем значения
func (c Counter) GetValue() int {
    return c.value
}

// Метод с получателем указателя
func (c *Counter) Increment() {
    c.value++
}

// Интерфейс
type Incrementer interface {
    Increment()
}

type Getter interface {
    GetValue() int
}

func main() {
    c1 := Counter{value: 0}        // Значение
    c2 := &Counter{value: 0}       // Указатель
    
    // Оба могут вызывать GetValue()
    var g1 Getter = c1
    var g2 Getter = c2
    
    // Только указатель может быть Incrementer
    // var i1 Incrementer = c1  // ОШИБКА!
    var i2 Incrementer = c2     // OK
    
    // Но можно взять адрес значения
    var i1 Incrementer = &c1    // OK
}
```

## Пустой интерфейс

### interface{}

Пустой интерфейс не имеет методов, поэтому **любой тип** реализует его:

```go
func printAnything(value interface{}) {
    fmt.Printf("Значение: %v, Тип: %T\n", value, value)
}

func main() {
    printAnything(42)
    printAnything("Привет")
    printAnything(true)
    printAnything([]int{1, 2, 3})
}
```

### Type Assertion

Получение конкретного типа из interface{}:

```go
func processValue(value interface{}) {
    // Простое утверждение типа
    str, ok := value.(string)
    if ok {
        fmt.Printf("Строка: %s\n", str)
    }
    
    // Утверждение с паникой (если тип не совпадает)
    str = value.(string) // Паника, если value не string
    
    // Switch по типу
    switch v := value.(type) {
    case string:
        fmt.Printf("Строка: %s (длина: %d)\n", v, len(v))
    case int:
        fmt.Printf("Число: %d (квадрат: %d)\n", v, v*v)
    case bool:
        fmt.Printf("Булево: %t\n", v)
    default:
        fmt.Printf("Неизвестный тип: %v\n", v)
    }
}
```

## Лучшие практики

### 1. Делайте интерфейсы маленькими

```go
// ХОРОШО - один метод
type Reader interface {
    Read([]byte) (int, error)
}

// ПЛОХО - много методов
type ReadWriteCloserSeeker interface {
    Read([]byte) (int, error)
    Write([]byte) (int, error)
    Close() error
    Seek(int64, int) (int64, error)
    // ... еще 20 методов
}
```

### 2. Определяйте интерфейсы там, где они используются

```go
// В пакете, который использует интерфейс
type EmailSender interface {
    SendEmail(to, subject, body string) error
}

type UserService struct {
    emailSender EmailSender
}

func (us *UserService) RegisterUser(email string) error {
    // Регистрация пользователя
    // ...
    
    // Отправка приветственного email
    return us.emailSender.SendEmail(email, "Добро пожаловать!", "Приветствуем!")
}
```

### 3. Используйте интерфейсы для зависимости инъекции

```go
// Зависимость от интерфейса, а не от конкретной реализации
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}

// Легко тестировать с моками
func TestUserService_RegisterUser(t *testing.T) {
    mockRepo := &MockUserRepository{}
    service := NewUserService(mockRepo)
    
    // Тестирование
}
```

## Тестирование с интерфейсами

### Создание моков

```go
// Интерфейс репозитория
type UserRepository interface {
    GetUserByID(id int) (*User, error)
    SaveUser(user *User) error
}

// Реальная реализация
type DBUserRepository struct {
    db *sql.DB
}

func (r *DBUserRepository) GetUserByID(id int) (*User, error) {
    // Реальная реализация с базой данных
}

// Мок для тестирования
type MockUserRepository struct {
    GetUserByIDFunc func(id int) (*User, error)
    SaveUserFunc    func(user *User) error
}

func (m *MockUserRepository) GetUserByID(id int) (*User, error) {
    if m.GetUserByIDFunc != nil {
        return m.GetUserByIDFunc(id)
    }
    return nil, nil
}

func (m *MockUserRepository) SaveUser(user *User) error {
    if m.SaveUserFunc != nil {
        return m.SaveUserFunc(user)
    }
    return nil
}
```

### Использование моков в тестах

```go
func TestUserService_GetUser(t *testing.T) {
    // Создаем мок репозитория
    mockRepo := &MockUserRepository{
        GetUserByIDFunc: func(id int) (*User, error) {
            if id == 1 {
                return &User{ID: 1, Name: "Иван"}, nil
            }
            return nil, errors.New("Пользователь не найден")
        },
    }
    
    // Создаем сервис с моком
    service := NewUserService(mockRepo)
    
    // Тестируем
    user, err := service.GetUser(1)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if user.Name != "Иван" {
        t.Errorf("Expected name 'Иван'; got '%s'", user.Name)
    }
}
```

## Распространенные ошибки

### 1. Использование слишком больших интерфейсов

```go
// ПЛОХО
type Database interface {
    Connect() error
    Disconnect() error
    Insert(data interface{}) error
    Update(id string, data interface{}) error
    Delete(id string) error
    FindByID(id string) (interface{}, error)
    FindAll() ([]interface{}, error)
    // ... еще 20 методов
}

// ЛУЧШЕ
type Inserter interface {
    Insert(data interface{}) error
}

type Finder interface {
    FindByID(id string) (interface{}, error)
}
```

### 2. Экспорт интерфейсов, которые не нужны другим

```go
// ПЛОХО - экспортируем интерфейс, который нужен только внутри
type internalProcessor interface {
    Process() error
}

// ЛУЧШЕ - не экспортируем внутренние интерфейсы
type processor interface {
    Process() error
}
```

### 3. Игнорирование ошибок type assertion

```go
// ПЛОХО - может вызвать панику
str := value.(string)

// ЛУЧШЕ - проверяем ошибку
str, ok := value.(string)
if !ok {
    // Обрабатываем ошибку
}
```

## Производительность интерфейсов

### Накладные расходы

Интерфейсы имеют небольшие накладные расходы:
- **Косвенный вызов метода**
- **Проверка типа во время выполнения**

```go
// Benchmark прямого вызова
func BenchmarkDirectCall(b *testing.B) {
    c := Counter{}
    for i := 0; i < b.N; i++ {
        c.GetValue()
    }
}

// Benchmark через интерфейс
func BenchmarkInterfaceCall(b *testing.B) {
    c := Counter{}
    var g Getter = c
    for i := 0; i < b.N; i++ {
        g.GetValue()
    }
}
```

### Когда использовать интерфейсы

Используйте интерфейсы когда:
- **Нужна абстракция**
- **Планируете тестирование**
- **Разные реализации одного поведения**
- **Уменьшение связности**

Не используйте когда:
- **Только одна реализация**
- **Критична производительность**
- **Простые случаи**

## Расширенные примеры

### 1. Интерфейсы для стратегий

```go
type SortStrategy interface {
    Sort([]int) []int
}

type QuickSort struct{}

func (qs QuickSort) Sort(data []int) []int {
    // Реализация быстрой сортировки
    return data
}

type MergeSort struct{}

func (ms MergeSort) Sort(data []int) []int {
    // Реализация сортировки слиянием
    return data
}

type Sorter struct {
    strategy SortStrategy
}

func (s *Sorter) SetStrategy(strategy SortStrategy) {
    s.strategy = strategy
}

func (s *Sorter) Sort(data []int) []int {
    return s.strategy.Sort(data)
}
```

### 2. Интерфейсы для плагинов

```go
type Plugin interface {
    Name() string
    Process(data []byte) ([]byte, error)
}

// Плагин шифрования
type EncryptionPlugin struct{}

func (ep EncryptionPlugin) Name() string {
    return "encryption"
}

func (ep EncryptionPlugin) Process(data []byte) ([]byte, error) {
    // Реализация шифрования
    return data, nil
}

// Плагин сжатия
type CompressionPlugin struct{}

func (cp CompressionPlugin) Name() string {
    return "compression"
}

func (cp CompressionPlugin) Process(data []byte) ([]byte, error) {
    // Реализация сжатия
    return data, nil
}

// Использование плагинов
func processWithPlugins(data []byte, plugins []Plugin) ([]byte, error) {
    result := data
    for _, plugin := range plugins {
        var err error
        result, err = plugin.Process(result)
        if err != nil {
            return nil, err
        }
    }
    return result, nil
}
```

## См. также

- [Интерфейсы для чайников](../concepts/interface.md) - базовое объяснение
- [Тестирование](../concepts/testing.md) - как использовать интерфейсы для моков
- [Архитектура](../concepts/architecture.md) - применение интерфейсов в архитектуре
- [HTTP серверы](../concepts/http-server.md) - стандартные интерфейсы в веб-разработке
- [Практические примеры](../examples/interfaces) - примеры кода