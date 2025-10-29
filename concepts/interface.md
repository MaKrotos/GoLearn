# Интерфейсы (Interfaces) - объяснение для чайников

## Что такое интерфейс?

Представьте, что интерфейс - это **набор правил**, которым должен следовать объект, чтобы считаться определенного типа.

В реальной жизни это похоже на:
- **Паспорт** - чтобы быть гражданином, нужно иметь паспорт с определенными данными
- **Водительские права** - чтобы водить машину, нужно пройти экзамен и получить права
- **Диплом** - чтобы называться специалистом, нужно получить образование

В программировании интерфейс определяет **что объект должен уметь делать**, но не **как** он это делает.

## Техническое определение

Интерфейс - это **тип, который определяет набор сигнатур методов**. Тип реализует интерфейс, **неявно** реализуя все его методы.

## Неявная реализация

В Go интерфейсы реализуются **неявно**. Это означает:
- Вы **не пишете** "этот тип реализует этот интерфейс"
- Тип **автоматически** реализует интерфейс, если имеет все нужные методы

### Пример

```go
// Определяем интерфейс
type Writer interface {
    Write([]byte) (int, error)
}

// Тип File реализует Writer неявно
type File struct{}

func (f File) Write(data []byte) (int, error) {
    fmt.Println("Запись в файл:", string(data))
    return len(data), nil
}

// Тип NetworkConnection тоже реализует Writer
type NetworkConnection struct{}

func (nc NetworkConnection) Write(data []byte) (int, error) {
    fmt.Println("Отправка по сети:", string(data))
    return len(data), nil
}

// Функция, которая работает с любым Writer
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

## Стандартные интерфейсы

### io.Reader

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Что-то, что может **читать** данные:

```go
func readExample(reader io.Reader) {
    buffer := make([]byte, 1024)
    n, err := reader.Read(buffer)
    if err != nil {
        fmt.Println("Ошибка чтения:", err)
        return
    }
    fmt.Printf("Прочитано %d байт: %s\n", n, buffer[:n])
}

func main() {
    // strings.Reader реализует io.Reader
    reader := strings.NewReader("Привет, мир!")
    readExample(reader)
    
    // os.File тоже реализует io.Reader
    file, _ := os.Open("example.txt")
    defer file.Close()
    readExample(file)
}
```

### io.Writer

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Что-то, что может **записывать** данные:

```go
func writeExample(writer io.Writer) {
    data := []byte("Привет, мир!")
    n, err := writer.Write(data)
    if err != nil {
        fmt.Println("Ошибка записи:", err)
        return
    }
    fmt.Printf("Записано %d байт\n", n)
}

func main() {
    // os.Stdout реализует io.Writer
    writeExample(os.Stdout)
    
    // bytes.Buffer тоже реализует io.Writer
    var buffer bytes.Buffer
    writeExample(&buffer)
    fmt.Println("В буфере:", buffer.String())
}
```

### fmt.Stringer

```go
type Stringer interface {
    String() string
}
```

Тип, который может **представить себя в виде строки**:

```go
type Person struct {
    Name string
    Age  int
}

// Реализуем Stringer для красивого вывода
func (p Person) String() string {
    return fmt.Sprintf("%s (%d лет)", p.Name, p.Age)
}

type Car struct {
    Brand string
    Model string
    Year  int
}

// Реализуем Stringer для автомобилей
func (c Car) String() string {
    return fmt.Sprintf("%s %s (%d)", c.Brand, c.Model, c.Year)
}

func main() {
    person := Person{Name: "Иван", Age: 30}
    car := Car{Brand: "Toyota", Model: "Camry", Year: 2020}
    
    // fmt.Println автоматически вызывает String()
    fmt.Println(person) // Вывод: Иван (30 лет)
    fmt.Println(car)    // Вывод: Toyota Camry (2020)
}
```

## Практические примеры использования

### Пример 1: Логирование

```go
type Logger interface {
    Info(message string)
    Error(message string)
}

type ConsoleLogger struct{}

func (cl ConsoleLogger) Info(message string) {
    fmt.Printf("INFO: %s\n", message)
}

func (cl ConsoleLogger) Error(message string) {
    fmt.Printf("ERROR: %s\n", message)
}

type FileLogger struct {
    file *os.File
}

func (fl FileLogger) Info(message string) {
    fmt.Fprintf(fl.file, "INFO: %s\n", message)
}

func (fl FileLogger) Error(message string) {
    fmt.Fprintf(fl.file, "ERROR: %s\n", message)
}

// Функция, которая работает с любым логгером
func doWork(logger Logger) {
    logger.Info("Начинаем работу")
    
    // Делаем что-то...
    if err := someOperation(); err != nil {
        logger.Error(fmt.Sprintf("Ошибка: %v", err))
    }
    
    logger.Info("Работа завершена")
}
```

### Пример 2: Платежные системы

```go
type PaymentProcessor interface {
    ProcessPayment(amount float64) error
    RefundPayment(transactionID string) error
}

type CreditCardProcessor struct{}

func (ccp CreditCardProcessor) ProcessPayment(amount float64) error {
    fmt.Printf("Обработка кредитной карты на сумму %.2f\n", amount)
    return nil
}

func (ccp CreditCardProcessor) RefundPayment(transactionID string) error {
    fmt.Printf("Возврат по кредитной карте, транзакция %s\n", transactionID)
    return nil
}

type PayPalProcessor struct{}

func (ppp PayPalProcessor) ProcessPayment(amount float64) error {
    fmt.Printf("Обработка PayPal на сумму %.2f\n", amount)
    return nil
}

func (ppp PayPalProcessor) RefundPayment(transactionID string) error {
    fmt.Printf("Возврат через PayPal, транзакция %s\n", transactionID)
    return nil
}

// Сервис, который может использовать разные способы оплаты
type OrderService struct {
    paymentProcessor PaymentProcessor
}

func (os OrderService) ProcessOrder(amount float64) error {
    fmt.Printf("Обрабатываем заказ на сумму %.2f\n", amount)
    return os.paymentProcessor.ProcessPayment(amount)
}
```

## Пустой интерфейс interface{}

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

## Type Assertion

Иногда нужно получить конкретный тип из интерфейса:

```go
func processValue(value interface{}) {
    // Проверяем тип
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
}
```

### 2. Используйте интерфейсы для зависимости инъекции

```go
// Зависимость от интерфейса, а не от конкретной реализации
func NewUserService(repo UserRepository) *UserService {
    return &UserService{repo: repo}
}
```

### 3. Определяйте интерфейсы там, где они используются

```go
// В пакете, который использует интерфейс
type EmailSender interface {
    SendEmail(to, subject, body string) error
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

## См. также

- [Тестирование с моками](../theory/testing.md) - как использовать интерфейсы для тестирования
- [Архитектура приложений](../theory/architecture.md) - где интерфейсы играют ключевую роль
- [Dependency Injection](../theory/di.md) - паттерн, активно использующий интерфейсы