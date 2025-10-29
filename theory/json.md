# Работа с JSON в Go: Полная теория

## Введение в JSON в Go

### Что такое JSON?

JSON (JavaScript Object Notation) - это **текстовый формат** обмена данными, который:
- **Легко читается** людьми
- **Легко парсится** машинами
- **Является стандартом** для API и конфигураций
- **Поддерживает** вложенность и массивы

### Зачем нужна работа с JSON в Go?

1. **API разработка** - большинство REST API используют JSON
2. **Конфигурации** - JSON часто используется для настроек
3. **Хранение данных** - многие системы хранят данные в JSON
4. **Интеграции** - обмен данными с другими системами
5. **Микросервисы** - межсервисное взаимодействие

## Основы работы с JSON в Go

### Пакет encoding/json

Go предоставляет встроенный пакет `encoding/json` для работы с JSON:

```go
import (
    "encoding/json"
    "fmt"
)

// Базовый пример маршалинга
func basicMarshal() {
    data := map[string]interface{}{
        "name": "John",
        "age":  30,
        "city": "New York",
    }
    
    jsonBytes, err := json.Marshal(data)
    if err != nil {
        fmt.Printf("Error marshaling: %v\n", err)
        return
    }
    
    fmt.Printf("JSON: %s\n", jsonBytes)
    // Вывод: JSON: {"age":30,"city":"New York","name":"John"}
}
```

### Структуры и теги

```go
// Определение структуры с тегами JSON
type Person struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Email   string `json:"email,omitempty"`
    Address string `json:"-"`              // Игнорировать это поле
    City    string `json:"city,omitempty"` // Опустить, если пусто
}

// Пример маршалинга структуры
func structMarshal() {
    person := Person{
        Name:    "Alice",
        Age:     25,
        Email:   "alice@example.com",
        Address: "Secret", // Это поле не попадет в JSON
        City:    "Boston",
    }
    
    jsonBytes, err := json.Marshal(person)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Person JSON: %s\n", jsonBytes)
    // Вывод: Person JSON: {"name":"Alice","age":25,"email":"alice@example.com","city":"Boston"}
}
```

## Практическая реализация работы с JSON

### 1. Маршалинг (Go в JSON)

```go
// json/marshal.go
package json

import (
    "encoding/json"
    "fmt"
    "time"
)

// User структура пользователя
type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email,omitempty"`
    Age       int       `json:"age"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
    Tags      []string  `json:"tags,omitempty"`
    Metadata  Metadata  `json:"metadata,omitempty"`
}

// Metadata дополнительные данные
type Metadata struct {
    LastLogin *time.Time `json:"last_login,omitempty"`
    Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// MarshalUser преобразует пользователя в JSON
func MarshalUser(user User) ([]byte, error) {
    return json.Marshal(user)
}

// MarshalUserIndent преобразует пользователя в форматированный JSON
func MarshalUserIndent(user User) ([]byte, error) {
    return json.MarshalIndent(user, "", "  ")
}

// MarshalCustomUser пользовательский маршалинг
func MarshalCustomUser(user User) ([]byte, error) {
    // Создаем кастомное представление
    custom := map[string]interface{}{
        "user_id": user.ID,
        "full_name": user.Name,
        "contact": map[string]string{
            "email": user.Email,
        },
        "profile": map[string]interface{}{
            "age": user.Age,
            "active": user.IsActive,
        },
        "joined": user.CreatedAt.Format("2006-01-02"),
    }
    
    return json.Marshal(custom)
}

// Пример использования
func ExampleMarshal() {
    now := time.Now()
    user := User{
        ID:       1,
        Name:     "John Doe",
        Email:    "john@example.com",
        Age:      30,
        IsActive: true,
        CreatedAt: now,
        Tags:     []string{"developer", "golang"},
        Metadata: Metadata{
            LastLogin: &now,
            Preferences: map[string]interface{}{
                "theme": "dark",
                "notifications": true,
            },
        },
    }
    
    // Обычный JSON
    jsonBytes, err := MarshalUser(user)
    if err != nil {
        fmt.Printf("Marshal error: %v\n", err)
        return
    }
    fmt.Printf("JSON: %s\n", jsonBytes)
    
    // Форматированный JSON
    jsonIndent, err := MarshalUserIndent(user)
    if err != nil {
        fmt.Printf("Marshal indent error: %v\n", err)
        return
    }
    fmt.Printf("Formatted JSON:\n%s\n", jsonIndent)
    
    // Кастомный JSON
    customJSON, err := MarshalCustomUser(user)
    if err != nil {
        fmt.Printf("Custom marshal error: %v\n", err)
        return
    }
    fmt.Printf("Custom JSON:\n%s\n", customJSON)
}
```

### 2. Анмаршалинг (JSON в Go)

```go
// json/unmarshal.go
package json

import (
    "encoding/json"
    "fmt"
    "time"
)

// UnmarshalUser преобразует JSON в пользователя
func UnmarshalUser(jsonData []byte) (*User, error) {
    var user User
    err := json.Unmarshal(jsonData, &user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// UnmarshalUserWithValidation преобразует JSON с валидацией
func UnmarshalUserWithValidation(jsonData []byte) (*User, error) {
    var user User
    err := json.Unmarshal(jsonData, &user)
    if err != nil {
        return nil, err
    }
    
    // Валидация данных
    if user.Name == "" {
        return nil, fmt.Errorf("name is required")
    }
    
    if user.Age < 0 || user.Age > 150 {
        return nil, fmt.Errorf("invalid age: %d", user.Age)
    }
    
    if user.Email != "" && !isValidEmail(user.Email) {
        return nil, fmt.Errorf("invalid email format")
    }
    
    return &user, nil
}

// isValidEmail простая проверка email
func isValidEmail(email string) bool {
    return len(email) > 0 && 
           len(email) < 255 && 
           email[0] != '@' && 
           email[len(email)-1] != '@'
}

// UnmarshalFlexibleUser гибкое преобразование JSON
func UnmarshalFlexibleUser(jsonData []byte) (*User, error) {
    // Используем map[string]interface{} для гибкого парсинга
    var rawData map[string]interface{}
    err := json.Unmarshal(jsonData, &rawData)
    if err != nil {
        return nil, err
    }
    
    user := &User{}
    
    // Обрабатываем поля по отдельности
    if id, ok := rawData["id"].(float64); ok {
        user.ID = int(id)
    }
    
    if name, ok := rawData["name"].(string); ok {
        user.Name = name
    }
    
    if email, ok := rawData["email"].(string); ok {
        user.Email = email
    }
    
    if age, ok := rawData["age"].(float64); ok {
        user.Age = int(age)
    }
    
    if isActive, ok := rawData["is_active"].(bool); ok {
        user.IsActive = isActive
    }
    
    // Обработка времени
    if createdAtStr, ok := rawData["created_at"].(string); ok {
        if createdAt, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
            user.CreatedAt = createdAt
        }
    }
    
    // Обработка массивов
    if tags, ok := rawData["tags"].([]interface{}); ok {
        user.Tags = make([]string, len(tags))
        for i, tag := range tags {
            if tagStr, ok := tag.(string); ok {
                user.Tags[i] = tagStr
            }
        }
    }
    
    return user, nil
}

// Пример использования
func ExampleUnmarshal() {
    jsonData := []byte(`{
        "id": 1,
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30,
        "is_active": true,
        "created_at": "2023-01-01T00:00:00Z",
        "tags": ["developer", "golang"]
    }`)
    
    // Обычный анмаршалинг
    user, err := UnmarshalUser(jsonData)
    if err != nil {
        fmt.Printf("Unmarshal error: %v\n", err)
        return
    }
    fmt.Printf("User: %+v\n", user)
    
    // Анмаршалинг с валидацией
    validUser, err := UnmarshalUserWithValidation(jsonData)
    if err != nil {
        fmt.Printf("Validation error: %v\n", err)
        return
    }
    fmt.Printf("Valid user: %+v\n", validUser)
    
    // Гибкий анмаршалинг
    flexibleUser, err := UnmarshalFlexibleUser(jsonData)
    if err != nil {
        fmt.Printf("Flexible unmarshal error: %v\n", err)
        return
    }
    fmt.Printf("Flexible user: %+v\n", flexibleUser)
}
```

### 3. Работа с потоками JSON

```go
// json/stream.go
package json

import (
    "encoding/json"
    "fmt"
    "io"
    "strings"
)

// StreamMarshal потоковый маршалинг
func StreamMarshal(w io.Writer, data interface{}) error {
    encoder := json.NewEncoder(w)
    encoder.SetIndent("", "  ") // Форматирование
    return encoder.Encode(data)
}

// StreamUnmarshal потоковый анмаршалинг
func StreamUnmarshal(r io.Reader, v interface{}) error {
    decoder := json.NewDecoder(r)
    return decoder.Decode(v)
}

// ProcessJSONArray обработка массива JSON объектов
func ProcessJSONArray(jsonArray string) error {
    reader := strings.NewReader(jsonArray)
    decoder := json.NewDecoder(reader)
    
    // Читаем начало массива
    token, err := decoder.Token()
    if err != nil {
        return err
    }
    
    if delim, ok := token.(json.Delim); !ok || delim != '[' {
        return fmt.Errorf("expected array start")
    }
    
    // Обрабатываем каждый элемент массива
    for decoder.More() {
        var user User
        err := decoder.Decode(&user)
        if err != nil {
            return err
        }
        
        // Обрабатываем пользователя
        fmt.Printf("Processing user: %+v\n", user)
    }
    
    // Читаем конец массива
    token, err = decoder.Token()
    if err != nil {
        return err
    }
    
    if delim, ok := token.(json.Delim); !ok || delim != ']' {
        return fmt.Errorf("expected array end")
    }
    
    return nil
}

// GenerateJSONArray генерация массива JSON
func GenerateJSONArray(users []User, w io.Writer) error {
    encoder := json.NewEncoder(w)
    encoder.SetIndent("", "  ")
    
    // Начинаем массив
    if _, err := w.Write([]byte("[\n")); err != nil {
        return err
    }
    
    // Записываем пользователей
    for i, user := range users {
        if i > 0 {
            if _, err := w.Write([]byte(",\n")); err != nil {
                return err
            }
        }
        
        if err := encoder.Encode(user); err != nil {
            return err
        }
    }
    
    // Заканчиваем массив
    if _, err := w.Write([]byte("\n]")); err != nil {
        return err
    }
    
    return nil
}

// Пример использования потоков
func ExampleStream() {
    // Потоковый маршалинг
    user := User{
        ID:   1,
        Name: "Stream User",
        Age:  25,
    }
    
    fmt.Println("Stream marshal:")
    err := StreamMarshal(&strings.Builder{}, user)
    if err != nil {
        fmt.Printf("Stream marshal error: %v\n", err)
    }
    
    // Обработка массива
    jsonArray := `[
        {"id": 1, "name": "User 1", "age": 25},
        {"id": 2, "name": "User 2", "age": 30},
        {"id": 3, "name": "User 3", "age": 35}
    ]`
    
    fmt.Println("Processing array:")
    err = ProcessJSONArray(jsonArray)
    if err != nil {
        fmt.Printf("Array processing error: %v\n", err)
    }
}
```

## Расширенные техники работы с JSON

### 1. Кастомный маршалинг и анмаршалинг

```go
// json/custom.go
package json

import (
    "encoding/json"
    "fmt"
    "strconv"
    "time"
)

// CustomTime кастомный тип времени
type CustomTime struct {
    time.Time
}

// MarshalJSON кастомный маршалинг времени
func (ct CustomTime) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf(`"%s"`, ct.Time.Format("2006-01-02 15:04:05"))), nil
}

// UnmarshalJSON кастомный анмаршалинг времени
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    // Убираем кавычки
    str := string(data[1 : len(data)-1])
    
    // Парсим время
    t, err := time.Parse("2006-01-02 15:04:05", str)
    if err != nil {
        return err
    }
    
    ct.Time = t
    return nil
}

// UserWithCustomTime пользователь с кастомным временем
type UserWithCustomTime struct {
    ID        int        `json:"id"`
    Name      string     `json:"name"`
    CreatedAt CustomTime `json:"created_at"`
}

// SensitiveString строка, которая скрывается при маршалинге
type SensitiveString string

// MarshalJSON скрывает чувствительные данные
func (s SensitiveString) MarshalJSON() ([]byte, error) {
    if len(s) == 0 {
        return []byte(`""`), nil
    }
    
    // Скрываем большую часть строки
    masked := string(s)
    if len(masked) > 4 {
        masked = masked[:2] + "***" + masked[len(masked)-2:]
    } else {
        masked = "***"
    }
    
    return json.Marshal(masked)
}

// UserWithSensitiveData пользователь с чувствительными данными
type UserWithSensitiveData struct {
    ID       int             `json:"id"`
    Name     string          `json:"name"`
    Password SensitiveString `json:"password"`
    SSN      SensitiveString `json:"ssn"`
}

// Пример использования кастомного маршалинга
func ExampleCustomMarshal() {
    // Кастомное время
    userTime := UserWithCustomTime{
        ID:        1,
        Name:      "Custom Time User",
        CreatedAt: CustomTime{time.Now()},
    }
    
    jsonBytes, err := json.Marshal(userTime)
    if err != nil {
        fmt.Printf("Custom marshal error: %v\n", err)
        return
    }
    
    fmt.Printf("Custom time JSON: %s\n", jsonBytes)
    
    // Чувствительные данные
    userSensitive := UserWithSensitiveData{
        ID:       1,
        Name:     "Sensitive User",
        Password: SensitiveString("supersecretpassword"),
        SSN:      SensitiveString("123-45-6789"),
    }
    
    jsonBytes, err = json.Marshal(userSensitive)
    if err != nil {
        fmt.Printf("Sensitive marshal error: %v\n", err)
        return
    }
    
    fmt.Printf("Sensitive data JSON: %s\n", jsonBytes)
}
```

### 2. Работа с неизвестными структурами

```go
// json/dynamic.go
package json

import (
    "encoding/json"
    "fmt"
)

// DynamicJSON динамическая структура JSON
type DynamicJSON map[string]interface{}

// Get получает значение по пути
func (d DynamicJSON) Get(path string) interface{} {
    // Простая реализация для одного уровня
    return d[path]
}

// Set устанавливает значение по пути
func (d DynamicJSON) Set(path string, value interface{}) {
    d[path] = value
}

// GetString получает строковое значение
func (d DynamicJSON) GetString(path string) (string, bool) {
    if val, ok := d[path].(string); ok {
        return val, true
    }
    return "", false
}

// GetInt получает целочисленное значение
func (d DynamicJSON) GetInt(path string) (int, bool) {
    switch val := d[path].(type) {
    case float64:
        return int(val), true
    case int:
        return val, true
    case int64:
        return int(val), true
    default:
        return 0, false
    }
}

// Merge объединяет две структуры JSON
func (d DynamicJSON) Merge(other DynamicJSON) DynamicJSON {
    result := make(DynamicJSON)
    
    // Копируем оригинальные значения
    for k, v := range d {
        result[k] = v
    }
    
    // Добавляем или перезаписываем значения из other
    for k, v := range other {
        result[k] = v
    }
    
    return result
}

// Flatten выравнивает вложенную структуру JSON
func (d DynamicJSON) Flatten() map[string]interface{} {
    result := make(map[string]interface{})
    flattenRecursive(d, "", result)
    return result
}

func flattenRecursive(data interface{}, prefix string, result map[string]interface{}) {
    switch val := data.(type) {
    case map[string]interface{}:
        for k, v := range val {
            newKey := k
            if prefix != "" {
                newKey = prefix + "." + k
            }
            flattenRecursive(v, newKey, result)
        }
    case []interface{}:
        for i, v := range val {
            newKey := fmt.Sprintf("%s[%d]", prefix, i)
            flattenRecursive(v, newKey, result)
        }
    default:
        if prefix != "" {
            result[prefix] = val
        }
    }
}

// Пример использования динамического JSON
func ExampleDynamicJSON() {
    // Создаем динамическую структуру
    dynamic := DynamicJSON{
        "name": "John",
        "age":  30,
        "address": map[string]interface{}{
            "street": "123 Main St",
            "city":   "New York",
        },
        "hobbies": []interface{}{"reading", "coding"},
    }
    
    // Получаем значения
    if name, ok := dynamic.GetString("name"); ok {
        fmt.Printf("Name: %s\n", name)
    }
    
    if age, ok := dynamic.GetInt("age"); ok {
        fmt.Printf("Age: %d\n", age)
    }
    
    // Устанавливаем новые значения
    dynamic.Set("email", "john@example.com")
    dynamic.Set("age", 31) // Обновляем возраст
    
    // Преобразуем в JSON
    jsonBytes, err := json.Marshal(dynamic)
    if err != nil {
        fmt.Printf("Marshal error: %v\n", err)
        return
    }
    
    fmt.Printf("Dynamic JSON: %s\n", jsonBytes)
    
    // Выравниваем структуру
    flattened := dynamic.Flatten()
    fmt.Printf("Flattened: %+v\n", flattened)
}
```

### 3. Валидация JSON

```go
// json/validation.go
package json

import (
    "encoding/json"
    "fmt"
    "regexp"
    "strings"
)

// ValidationError ошибка валидации
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("field '%s': %s", e.Field, e.Message)
}

// ValidationErrors список ошибок валидации
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
    if len(ve) == 0 {
        return ""
    }
    
    messages := make([]string, len(ve))
    for i, err := range ve {
        messages[i] = err.Error()
    }
    
    return strings.Join(messages, "; ")
}

// JSONValidator валидатор JSON
type JSONValidator struct {
    requiredFields []string
    fieldTypes     map[string]string
    fieldPatterns  map[string]*regexp.Regexp
    maxLengths     map[string]int
}

// NewJSONValidator создает новый валидатор
func NewJSONValidator() *JSONValidator {
    return &JSONValidator{
        requiredFields: make([]string, 0),
        fieldTypes:     make(map[string]string),
        fieldPatterns:  make(map[string]*regexp.Regexp),
        maxLengths:     make(map[string]int),
    }
}

// Required добавляет обязательное поле
func (v *JSONValidator) Required(fields ...string) *JSONValidator {
    v.requiredFields = append(v.requiredFields, fields...)
    return v
}

// Type устанавливает тип поля
func (v *JSONValidator) Type(field, fieldType string) *JSONValidator {
    v.fieldTypes[field] = fieldType
    return v
}

// Pattern устанавливает паттерн для поля
func (v *JSONValidator) Pattern(field, pattern string) *JSONValidator {
    if re, err := regexp.Compile(pattern); err == nil {
        v.fieldPatterns[field] = re
    }
    return v
}

// MaxLength устанавливает максимальную длину для поля
func (v *JSONValidator) MaxLength(field string, length int) *JSONValidator {
    v.maxLengths[field] = length
    return v
}

// Validate валидирует JSON данные
func (v *JSONValidator) Validate(data []byte) error {
    var jsonData map[string]interface{}
    if err := json.Unmarshal(data, &jsonData); err != nil {
        return fmt.Errorf("invalid JSON format: %w", err)
    }
    
    var errors ValidationErrors
    
    // Проверяем обязательные поля
    for _, field := range v.requiredFields {
        if _, exists := jsonData[field]; !exists {
            errors = append(errors, ValidationError{
                Field:   field,
                Message: "field is required",
            })
        }
    }
    
    // Проверяем типы и другие ограничения
    for field, value := range jsonData {
        // Проверяем тип
        if expectedType, ok := v.fieldTypes[field]; ok {
            if !v.checkType(value, expectedType) {
                errors = append(errors, ValidationError{
                    Field:   field,
                    Message: fmt.Sprintf("expected type %s", expectedType),
                })
            }
        }
        
        // Проверяем паттерн
        if re, ok := v.fieldPatterns[field]; ok {
            if str, ok := value.(string); ok {
                if !re.MatchString(str) {
                    errors = append(errors, ValidationError{
                        Field:   field,
                        Message: "value does not match pattern",
                    })
                }
            }
        }
        
        // Проверяем максимальную длину
        if maxLength, ok := v.maxLengths[field]; ok {
            if str, ok := value.(string); ok {
                if len(str) > maxLength {
                    errors = append(errors, ValidationError{
                        Field:   field,
                        Message: fmt.Sprintf("value exceeds maximum length of %d", maxLength),
                    })
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return errors
    }
    
    return nil
}

// checkType проверяет тип значения
func (v *JSONValidator) checkType(value interface{}, expectedType string) bool {
    switch expectedType {
    case "string":
        _, ok := value.(string)
        return ok
    case "number":
        switch value.(type) {
        case float64, float32, int, int32, int64:
            return true
        default:
            return false
        }
    case "boolean":
        _, ok := value.(bool)
        return ok
    case "array":
        _, ok := value.([]interface{})
        return ok
    case "object":
        _, ok := value.(map[string]interface{})
        return ok
    default:
        return true // Неизвестный тип - считаем валидным
    }
}

// Пример использования валидатора
func ExampleValidation() {
    // Создаем валидатор
    validator := NewJSONValidator().
        Required("name", "email").
        Type("name", "string").
        Type("age", "number").
        Pattern("email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`).
        MaxLength("name", 50)
    
    // Валидные данные
    validJSON := []byte(`{
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30
    }`)
    
    if err := validator.Validate(validJSON); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    } else {
        fmt.Println("Valid JSON")
    }
    
    // Невалидные данные
    invalidJSON := []byte(`{
        "name": "John Doe",
        "email": "invalid-email",
        "age": "thirty"
    }`)
    
    if err := validator.Validate(invalidJSON); err != nil {
        fmt.Printf("Validation error: %v\n", err)
    } else {
        fmt.Println("Valid JSON")
    }
}
```

## Тестирование работы с JSON

### 1. Модульные тесты JSON

```go
// json/marshal_test.go
package json

import (
    "encoding/json"
    "reflect"
    "testing"
    "time"
)

func TestMarshalUser(t *testing.T) {
    now := time.Now()
    user := User{
        ID:        1,
        Name:      "Test User",
        Email:     "test@example.com",
        Age:       25,
        IsActive:  true,
        CreatedAt: now,
        Tags:      []string{"tag1", "tag2"},
    }
    
    jsonBytes, err := MarshalUser(user)
    if err != nil {
        t.Fatalf("MarshalUser failed: %v", err)
    }
    
    // Проверяем, что это валидный JSON
    var parsed map[string]interface{}
    if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
        t.Fatalf("Invalid JSON: %v", err)
    }
    
    // Проверяем поля
    if id, ok := parsed["id"].(float64); !ok || int(id) != user.ID {
        t.Errorf("Expected id %d, got %v", user.ID, id)
    }
    
    if name, ok := parsed["name"].(string); !ok || name != user.Name {
        t.Errorf("Expected name %s, got %s", user.Name, name)
    }
    
    if email, ok := parsed["email"].(string); !ok || email != user.Email {
        t.Errorf("Expected email %s, got %s", user.Email, email)
    }
}

func TestUnmarshalUser(t *testing.T) {
    jsonData := []byte(`{
        "id": 1,
        "name": "Test User",
        "email": "test@example.com",
        "age": 25,
        "is_active": true,
        "created_at": "2023-01-01T00:00:00Z",
        "tags": ["tag1", "tag2"]
    }`)
    
    user, err := UnmarshalUser(jsonData)
    if err != nil {
        t.Fatalf("UnmarshalUser failed: %v", err)
    }
    
    expectedUser := &User{
        ID:       1,
        Name:     "Test User",
        Email:    "test@example.com",
        Age:      25,
        IsActive: true,
        Tags:     []string{"tag1", "tag2"},
    }
    
    // Для сравнения времени создаем его отдельно
    expectedUser.CreatedAt, _ = time.Parse(time.RFC3339, "2023-01-01T00:00:00Z")
    
    if user.ID != expectedUser.ID {
        t.Errorf("Expected ID %d, got %d", expectedUser.ID, user.ID)
    }
    
    if user.Name != expectedUser.Name {
        t.Errorf("Expected Name %s, got %s", expectedUser.Name, user.Name)
    }
    
    if !reflect.DeepEqual(user.Tags, expectedUser.Tags) {
        t.Errorf("Expected Tags %v, got %v", expectedUser.Tags, user.Tags)
    }
}

func TestCustomTimeMarshal(t *testing.T) {
    now := time.Now()
    customTime := CustomTime{now}
    
    jsonBytes, err := json.Marshal(customTime)
    if err != nil {
        t.Fatalf("CustomTime marshal failed: %v", err)
    }
    
    expected := now.Format("2006-01-02 15:04:05")
    expectedJSON := `"` + expected + `"`
    
    if string(jsonBytes) != expectedJSON {
        t.Errorf("Expected %s, got %s", expectedJSON, string(jsonBytes))
    }
}

func TestSensitiveStringMarshal(t *testing.T) {
    sensitive := SensitiveString("supersecretpassword")
    
    jsonBytes, err := json.Marshal(sensitive)
    if err != nil {
        t.Fatalf("SensitiveString marshal failed: %v", err)
    }
    
    // Проверяем, что строка замаскирована
    result := string(jsonBytes)
    if result == `"supersecretpassword"` {
        t.Error("Password should be masked")
    }
    
    if len(result) < 5 {
        t.Error("Masked password should not be empty")
    }
}
```

### 2. Интеграционные тесты JSON

```go
// integration/json_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"
    "yourproject/json"
)

func TestJSONAPIIntegration(t *testing.T) {
    // Создаем тестовый HTTP обработчик
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "POST":
            // Читаем тело запроса
            var user json.User
            if err := json.StreamUnmarshal(r.Body, &user); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            
            // Валидация
            if user.Name == "" {
                http.Error(w, "name is required", http.StatusBadRequest)
                return
            }
            
            // Устанавливаем ID и время создания
            user.ID = 123
            user.CreatedAt = time.Now()
            
            // Отправляем ответ
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusCreated)
            json.StreamMarshal(w, user)
            
        case "GET":
            // Возвращаем список пользователей
            users := []json.User{
                {
                    ID:        1,
                    Name:      "User 1",
                    Age:       25,
                    IsActive:  true,
                    CreatedAt: time.Now(),
                },
                {
                    ID:        2,
                    Name:      "User 2",
                    Age:       30,
                    IsActive:  false,
                    CreatedAt: time.Now(),
                },
            }
            
            w.Header().Set("Content-Type", "application/json")
            json.GenerateJSONArray(users, w)
        }
    })
    
    // Тест POST запроса
    t.Run("CreateUser", func(t *testing.T) {
        userJSON := `{
            "name": "Test User",
            "email": "test@example.com",
            "age": 25,
            "is_active": true
        }`
        
        req := httptest.NewRequest("POST", "/users", strings.NewReader(userJSON))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
        
        if w.Code != http.StatusCreated {
            t.Errorf("Expected status 201, got %d", w.Code)
        }
        
        if w.Header().Get("Content-Type") != "application/json" {
            t.Error("Expected Content-Type application/json")
        }
        
        // Проверяем, что ответ содержит ID
        var responseUser json.User
        if err := json.Unmarshal(w.Body.Bytes(), &responseUser); err != nil {
            t.Fatalf("Failed to unmarshal response: %v", err)
        }
        
        if responseUser.ID == 0 {
            t.Error("Expected user ID to be set")
        }
        
        if responseUser.Name != "Test User" {
            t.Errorf("Expected name 'Test User', got '%s'", responseUser.Name)
        }
    })
    
    // Тест GET запроса
    t.Run("GetUsers", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/users", nil)
        w := httptest.NewRecorder()
        
        handler.ServeHTTP(w, req)
        
        if w.Code != http.StatusOK {
            t.Errorf("Expected status 200, got %d", w.Code)
        }
        
        // Проверяем, что ответ - массив
        if !bytes.HasPrefix(w.Body.Bytes(), []byte("[")) {
            t.Error("Expected array response")
        }
        
        if !bytes.HasSuffix(w.Body.Bytes(), []byte("]")) {
            t.Error("Expected array response")
        }
        
        // Проверяем, что можно распарсить как массив
        var users []json.User
        if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
            t.Fatalf("Failed to unmarshal users array: %v", err)
        }
        
        if len(users) != 2 {
            t.Errorf("Expected 2 users, got %d", len(users))
        }
    })
    
    // Тест невалидного JSON
    t.Run("InvalidJSON", func(t *testing.T) {
        invalidJSON := `{
            "name": "Test User",
            "email": "test@example.com",
            "age": 25,
            "is_active": true,
            // Invalid JSON - trailing comma
        }`
        
        req := httptest.NewRequest("POST", "/users", strings.NewReader(invalidJSON))
        req.Header.Set("Content-Type", "application/json")
        
        w := httptest.NewRecorder()
        handler.ServeHTTP(w, req)
        
        if w.Code != http.StatusBadRequest {
            t.Errorf("Expected status 400, got %d", w.Code)
        }
    })
}

func TestJSONValidationIntegration(t *testing.T) {
    validator := json.NewJSONValidator().
        Required("name", "email").
        Type("name", "string").
        Type("age", "number").
        Pattern("email", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    
    // Валидные данные
    validData := []byte(`{
        "name": "John Doe",
        "email": "john@example.com",
        "age": 30
    }`)
    
    if err := validator.Validate(validData); err != nil {
        t.Errorf("Valid data should not produce validation errors: %v", err)
    }
    
    // Невалидные данные - отсутствует обязательное поле
    missingFieldData := []byte(`{
        "name": "John Doe",
        "age": 30
    }`)
    
    if err := validator.Validate(missingFieldData); err == nil {
        t.Error("Missing required field should produce validation error")
    }
    
    // Невалидные данные - неправильный тип
    wrongTypeData := []byte(`{
        "name": "John Doe",
        "email": "john@example.com",
        "age": "thirty"
    }`)
    
    if err := validator.Validate(wrongTypeData); err == nil {
        t.Error("Wrong type should produce validation error")
    }
    
    // Невалидные данные - неправильный формат email
    invalidEmailData := []byte(`{
        "name": "John Doe",
        "email": "invalid-email",
        "age": 30
    }`)
    
    if err := validator.Validate(invalidEmailData); err == nil {
        t.Error("Invalid email should produce validation error")
    }
}
```

## Лучшие практики работы с JSON

### 1. Эффективное использование тегов

```go
// json/best_practices.go
package json

import (
    "encoding/json"
    "time"
)

// BestPracticeUser демонстрирует лучшие практики тегов
type BestPracticeUser struct {
    // Обязательные поля без omitempty
    ID   int    `json:"id"`
    Name string `json:"name"`
    
    // Опциональные поля с omitempty
    Email    string `json:"email,omitempty"`
    Phone    string `json:"phone,omitempty"`
    Bio      string `json:"bio,omitempty"`
    
    // Числовые поля с правильными типами
    Age      int     `json:"age"`
    Height   float64 `json:"height,omitempty"`
    Weight   float64 `json:"weight,omitempty"`
    
    // Булевы поля
    IsActive bool `json:"is_active"`
    IsAdmin  bool `json:"is_admin,omitempty"` // Только если true
    
    // Временные поля
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt *time.Time `json:"updated_at,omitempty"` // Указатель для опциональности
    
    // Массивы и слайсы
    Tags     []string          `json:"tags,omitempty"`
    Roles    []string          `json:"roles,omitempty"`
    Metadata map[string]string `json:"metadata,omitempty"`
    
    // Игнорируемые поля
    InternalID string `json:"-"` // Полностью игнорируется
    TempField  string `json:"temp,omitempty"` // Используется только если не пусто
}

// EmbeddedUser пользователь с вложенными структурами
type EmbeddedUser struct {
    BestPracticeUser
    Profile Profile `json:"profile"`
}

type Profile struct {
    AvatarURL string `json:"avatar_url,omitempty"`
    Theme     string `json:"theme,omitempty"`
    Language  string `json:"language,omitempty"`
}

// CustomMarshalUser пользователь с кастомным маршалингом
type CustomMarshalUser struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"-"`
    CreatedAtStr string  `json:"created_at"` // Кастомное поле
}

// MarshalJSON кастомный маршалинг
func (u CustomMarshalUser) MarshalJSON() ([]byte, error) {
    // Создаем копию с кастомным форматом времени
    type Alias CustomMarshalUser
    return json.Marshal(&struct {
        *Alias
        CreatedAt string `json:"created_at"`
    }{
        Alias: (*Alias)(&u),
        CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
    })
}

// UnmarshalJSON кастомный анмаршалинг
func (u *CustomMarshalUser) UnmarshalJSON(data []byte) error {
    type Alias CustomMarshalUser
    aux := &struct {
        *Alias
        CreatedAt string `json:"created_at"`
    }{
        Alias: (*Alias)(u),
    }
    
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    
    if aux.CreatedAt != "" {
        t, err := time.Parse("2006-01-02 15:04:05", aux.CreatedAt)
        if err != nil {
            return err
        }
        u.CreatedAt = t
    }
    
    return nil
}
```

### 2. Обработка ошибок

```go
// json/error_handling.go
package json

import (
    "encoding/json"
    "fmt"
)

// JSONError кастомная ошибка JSON
type JSONError struct {
    Op  string
    Err error
    Data []byte
}

func (e *JSONError) Error() string {
    return fmt.Sprintf("JSON error during %s: %v", e.Op, e.Err)
}

func (e *JSONError) Unwrap() error {
    return e.Err
}

// SafeMarshal безопасный маршалинг с логированием
func SafeMarshal(v interface{}) ([]byte, error) {
    data, err := json.Marshal(v)
    if err != nil {
        return nil, &JSONError{
            Op:  "marshal",
            Err: err,
            Data: nil,
        }
    }
    return data, nil
}

// SafeUnmarshal безопасный анмаршалинг с логированием
func SafeUnmarshal(data []byte, v interface{}) error {
    err := json.Unmarshal(data, v)
    if err != nil {
        return &JSONError{
            Op:  "unmarshal",
            Err: err,
            Data: data,
        }
    }
    return nil
}

// FallbackUnmarshal анмаршалинг с fallback стратегией
func FallbackUnmarshal(data []byte, v interface{}) error {
    // Пытаемся стандартный анмаршалинг
    if err := json.Unmarshal(data, v); err == nil {
        return nil
    }
    
    // Если не удалось, пробуем гибкий подход
    var rawData map[string]interface{}
    if err := json.Unmarshal(data, &rawData); err != nil {
        return &JSONError{
            Op:  "fallback unmarshal",
            Err: err,
            Data: data,
        }
    }
    
    // Здесь можно реализовать логику маппинга rawData в v
    // В упрощенном виде просто возвращаем ошибку
    return &JSONError{
        Op:  "fallback mapping",
        Err: fmt.Errorf("unable to map data to target type"),
        Data: data,
    }
}
```

### 3. Производительность

```go
// json/performance.go
package json

import (
    "bytes"
    "encoding/json"
    "sync"
)

// JSONPool пул буферов для JSON операций
type JSONPool struct {
    bufferPool sync.Pool
    encoderPool sync.Pool
    decoderPool sync.Pool
}

// NewJSONPool создает новый пул
func NewJSONPool() *JSONPool {
    return &JSONPool{
        bufferPool: sync.Pool{
            New: func() interface{} {
                return &bytes.Buffer{}
            },
        },
        encoderPool: sync.Pool{
            New: func() interface{} {
                buf := &bytes.Buffer{}
                return json.NewEncoder(buf)
            },
        },
        decoderPool: sync.Pool{
            New: func() interface{} {
                return json.NewDecoder(bytes.NewReader(nil))
            },
        },
    }
}

// MarshalWithPool маршалинг с использованием пула
func (p *JSONPool) MarshalWithPool(v interface{}) ([]byte, error) {
    buf := p.bufferPool.Get().(*bytes.Buffer)
    defer p.bufferPool.Put(buf)
    buf.Reset()
    
    encoder := json.NewEncoder(buf)
    if err := encoder.Encode(v); err != nil {
        return nil, err
    }
    
    // Убираем последний символ новой строки, добавленный Encode
    result := buf.Bytes()
    if len(result) > 0 && result[len(result)-1] == '\n' {
        result = result[:len(result)-1]
    }
    
    // Создаем копию, так как буфер будет возвращен в пул
    return append([]byte(nil), result...), nil
}

// UnmarshalWithPool анмаршалинг с использованием пула
func (p *JSONPool) UnmarshalWithPool(data []byte, v interface{}) error {
    reader := bytes.NewReader(data)
    decoder := json.NewDecoder(reader)
    return decoder.Decode(v)
}

// BatchProcessor пакетная обработка JSON
type BatchProcessor struct {
    pool *JSONPool
    batchSize int
}

// NewBatchProcessor создает новый процессор
func NewBatchProcessor(batchSize int) *BatchProcessor {
    return &BatchProcessor{
        pool: NewJSONPool(),
        batchSize: batchSize,
    }
}

// ProcessBatch обрабатывает пакет JSON объектов
func (bp *BatchProcessor) ProcessBatch(jsonArray []byte, processor func(interface{}) error) error {
    var items []interface{}
    if err := json.Unmarshal(jsonArray, &items); err != nil {
        return err
    }
    
    // Обрабатываем пакетами
    for i := 0; i < len(items); i += bp.batchSize {
        end := i + bp.batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        for _, item := range batch {
            if err := processor(item); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

## Распространенные ошибки и их решение

### 1. Проблемы с типами

```go
// ПЛОХО - неправильная работа с типами
func BadTypeHandling() {
    var data map[string]interface{}
    jsonData := []byte(`{"age": 25, "score": 95.5}`)
    
    json.Unmarshal(jsonData, &data)
    
    // Ошибка - age это float64, не int!
    age := data["age"].(int) // panic: interface conversion
    
    // Ошибка - score это float64, не string!
    score := data["score"].(string) // panic: interface conversion
}

// ХОРОШО - правильная работа с типами
func GoodTypeHandling() {
    var data map[string]interface{}
    jsonData := []byte(`{"age": 25, "score": 95.5}`)
    
    json.Unmarshal(jsonData, &data)
    
    // Правильно - проверка типа
    var age int
    if ageFloat, ok := data["age"].(float64); ok {
        age = int(ageFloat)
    }
    
    // Правильно - использование switch
    switch score := data["score"].(type) {
    case float64:
        fmt.Printf("Score: %.1f\n", score)
    case string:
        fmt.Printf("Score: %s\n", score)
    default:
        fmt.Println("Unknown score type")
    }
}
```

### 2. Утечки памяти

```go
// ПЛОХО - потенциальная утечка памяти
func BadMemoryUsage() {
    for i := 0; i < 1000000; i++ {
        user := &User{
            ID:   i,
            Name: fmt.Sprintf("User %d", i),
        }
        
        jsonBytes, _ := json.Marshal(user)
        // jsonBytes может быть большим и не освобождаться сразу
        processJSON(jsonBytes)
    }
}

// ХОРОШО - использование пулов
func GoodMemoryUsage() {
    pool := NewJSONPool()
    
    for i := 0; i < 1000000; i++ {
        user := &User{
            ID:   i,
            Name: fmt.Sprintf("User %d", i),
        }
        
        jsonBytes, _ := pool.MarshalWithPool(user)
        processJSON(jsonBytes)
        // Буфер будет возвращен в пул
    }
}
```

### 3. Игнорирование ошибок

```go
// ПЛОХО - игнорирование ошибок
func BadErrorHandling() {
    var user User
    jsonData := []byte(`{"invalid": "json"`)
    
    json.Unmarshal(jsonData, &user) // Игнорируем ошибку
    // Продолжаем работу с потенциально невалидными данными
}

// ХОРОШО - обработка ошибок
func GoodErrorHandling() {
    var user User
    jsonData := []byte(`{"invalid": "json"`)
    
    if err := json.Unmarshal(jsonData, &user); err != nil {
        log.Printf("Failed to unmarshal JSON: %v", err)
        // Обрабатываем ошибку соответствующим образом
        return
    }
    
    // Продолжаем работу только с валидными данными
}
```

## Мониторинг и отладка JSON

### 1. Логирование JSON операций

```go
// json/logging.go
package json

import (
    "encoding/json"
    "log"
    "time"
)

// LoggedMarshal маршалинг с логированием
func LoggedMarshal(v interface{}) ([]byte, error) {
    start := time.Now()
    
    data, err := json.Marshal(v)
    
    duration := time.Since(start)
    log.Printf("JSON Marshal: duration=%v, error=%v, size=%d bytes", 
        duration, err, len(data))
    
    return data, err
}

// LoggedUnmarshal анмаршалинг с логированием
func LoggedUnmarshal(data []byte, v interface{}) error {
    start := time.Now()
    
    err := json.Unmarshal(data, v)
    
    duration := time.Since(start)
    log.Printf("JSON Unmarshal: duration=%v, error=%v, size=%d bytes", 
        duration, err, len(data))
    
    return err
}
```

### 2. Метрики производительности

```go
// json/metrics.go
package json

import (
    "sync"
    "time"
)

// JSONMetrics метрики JSON операций
type JSONMetrics struct {
    marshalCount     int64
    unmarshalCount   int64
    marshalTime      time.Duration
    unmarshalTime    time.Duration
    marshalErrors    int64
    unmarshalErrors  int64
    mu               sync.RWMutex
}

var metrics = &JSONMetrics{}

// RecordMarshalMetrics записывает метрики маршалинга
func RecordMarshalMetrics(duration time.Duration, err error) {
    metrics.mu.Lock()
    defer metrics.mu.Unlock()
    
    metrics.marshalCount++
    metrics.marshalTime += duration
    
    if err != nil {
        metrics.marshalErrors++
    }
}

// RecordUnmarshalMetrics записывает метрики анмаршалинга
func RecordUnmarshalMetrics(duration time.Duration, err error) {
    metrics.mu.Lock()
    defer metrics.mu.Unlock()
    
    metrics.unmarshalCount++
    metrics.unmarshalTime += duration
    
    if err != nil {
        metrics.unmarshalErrors++
    }
}

// GetMetrics получает текущие метрики
func GetMetrics() (int64, int64, time.Duration, time.Duration, int64, int64) {
    metrics.mu.RLock()
    defer metrics.mu.RUnlock()
    
    return metrics.marshalCount, metrics.unmarshalCount,
           metrics.marshalTime, metrics.unmarshalTime,
           metrics.marshalErrors, metrics.unmarshalErrors
}

// InstrumentedMarshal инструментированный маршалинг
func InstrumentedMarshal(v interface{}) ([]byte, error) {
    start := time.Now()
    
    data, err := json.Marshal(v)
    
    RecordMarshalMetrics(time.Since(start), err)
    return data, err
}

// InstrumentedUnmarshal инструментированный анмаршалинг
func InstrumentedUnmarshal(data []byte, v interface{}) error {
    start := time.Now()
    
    err := json.Unmarshal(data, v)
    
    RecordUnmarshalMetrics(time.Since(start), err)
    return err
}
```

## См. также

- [HTTP серверы](../concepts/http-server.md) - работа с JSON в веб-приложениях
- [Тестирование](../concepts/testing.md) - как тестировать JSON операции
- [Профилирование](../concepts/profiling.md) - как измерять производительность
- [Практические примеры](../examples/json) - примеры кода