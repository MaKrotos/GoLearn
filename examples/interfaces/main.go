package main

import (
	"fmt"
	"math"
)

// Shape интерфейс для геометрических фигур
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rectangle прямоугольник
type Rectangle struct {
	Width, Height float64
}

// Area реализация метода интерфейса
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter реализация метода интерфейса
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Circle круг
type Circle struct {
	Radius float64
}

// Area реализация метода интерфейса
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Perimeter реализация метода интерфейса
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// Пример 1: Базовое использование интерфейсов
func basicInterfaces() {
	fmt.Println("=== Базовое использование интерфейсов ===")
	
	// Создаем фигуры
	rectangle := Rectangle{Width: 10, Height: 5}
	circle := Circle{Radius: 3}
	
	// Используем интерфейс
	shapes := []Shape{rectangle, circle}
	
	for _, shape := range shapes {
		fmt.Printf("Площадь: %.2f, Периметр: %.2f\n", shape.Area(), shape.Perimeter())
	}
}

// Writer интерфейс для записи данных
type Writer interface {
	Write([]byte) (int, error)
}

// Reader интерфейс для чтения данных
type Reader interface {
	Read([]byte) (int, error)
}

// ReadWriter интерфейс, объединяющий Reader и Writer
type ReadWriter interface {
	Reader
	Writer
}

// StringWriter простая реализация Writer
type StringWriter struct {
	data []byte
}

// Write реализация Writer интерфейса
func (sw *StringWriter) Write(p []byte) (int, error) {
	sw.data = append(sw.data, p...)
	return len(p), nil
}

// String возвращает строковое представление
func (sw *StringWriter) String() string {
	return string(sw.data)
}

// Пример 2: Композиция интерфейсов
func interfaceComposition() {
	fmt.Println("\n=== Композиция интерфейсов ===")
	
	// Создаем StringWriter
	sw := &StringWriter{}
	
	// Используем как Writer
	n, err := sw.Write([]byte("Привет, интерфейсы!"))
	if err != nil {
		fmt.Printf("Ошибка записи: %v\n", err)
		return
	}
	fmt.Printf("Записано %d байт\n", n)
	fmt.Printf("Содержимое: %s\n", sw.String())
}

// Animal интерфейс для животных
type Animal interface {
	Speak() string
	Move() string
}

// Dog структура собаки
type Dog struct{}

// Speak реализация Animal интерфейса
func (d Dog) Speak() string {
	return "Гав-гав!"
}

// Move реализация Animal интерфейса
func (d Dog) Move() string {
	return "Бегает на четырех лапах"
}

// Bird структура птицы
type Bird struct{}

// Speak реализация Animal интерфейса
func (b Bird) Speak() string {
	return "Чирик-чирик!"
}

// Move реализация Animal интерфейса
func (b Bird) Move() string {
	return "Летает в небе"
}

// Пример 3: Полиморфизм с интерфейсами
func polymorphism() {
	fmt.Println("\n=== Полиморфизм с интерфейсами ===")
	
	// Создаем животных
	animals := []Animal{Dog{}, Bird{}}
	
	// Используем полиморфизм
	for _, animal := range animals {
		fmt.Printf("Животное говорит: %s\n", animal.Speak())
		fmt.Printf("Животное двигается: %s\n", animal.Move())
	}
}

// Stringer интерфейс из пакета fmt
type Person struct {
	Name string
	Age  int
}

// String реализация Stringer интерфейса
func (p Person) String() string {
	return fmt.Sprintf("%s (%d лет)", p.Name, p.Age)
}

// Пример 4: Стандартные интерфейсы
func standardInterfaces() {
	fmt.Println("\n=== Стандартные интерфейсы ===")
	
	// Используем Stringer интерфейс
	person := Person{Name: "Иван", Age: 30}
	fmt.Println(person) // Автоматически вызывает String()
	
	// Можно явно вызвать
	fmt.Println("Явный вызов String():", person.String())
}

// EmptyInterface пустой интерфейс (interface{})
type EmptyInterface interface{}

// Пример 5: Пустой интерфейс
func emptyInterface() {
	fmt.Println("\n=== Пустой интерфейс ===")
	
	// Пустой интерфейс может хранить любое значение
	var empty EmptyInterface
	
	empty = 42
	fmt.Printf("Пустой интерфейс содержит: %v (тип: %T)\n", empty, empty)
	
	empty = "строка"
	fmt.Printf("Пустой интерфейс содержит: %v (тип: %T)\n", empty, empty)
	
	empty = []int{1, 2, 3}
	fmt.Printf("Пустой интерфейс содержит: %v (тип: %T)\n", empty, empty)
	
	// Type assertion
	if str, ok := empty.([]int); ok {
		fmt.Printf("Преобразовано в slice: %v\n", str)
	}
}

// Пример 6: Type switch
func typeSwitch() {
	fmt.Println("\n=== Type switch ===")
	
	values := []interface{}{42, "hello", true, 3.14}
	
	for _, v := range values {
		switch val := v.(type) {
		case int:
			fmt.Printf("Целое число: %d\n", val)
		case string:
			fmt.Printf("Строка: %s\n", val)
		case bool:
			fmt.Printf("Булево значение: %t\n", val)
		case float64:
			fmt.Printf("Число с плавающей точкой: %.2f\n", val)
		default:
			fmt.Printf("Неизвестный тип: %T\n", val)
		}
	}
}

// Error интерфейс для ошибок
type CustomError struct {
	Message string
	Code    int
}

// Error реализация error интерфейса
func (e CustomError) Error() string {
	return fmt.Sprintf("Ошибка %d: %s", e.Code, e.Message)
}

// Пример 7: Интерфейс ошибок
func errorInterface() {
	fmt.Println("\n=== Интерфейс ошибок ===")
	
	// Создаем кастомную ошибку
	err := CustomError{
		Message: "Неверные данные",
		Code:    1001,
	}
	
	// Используем как error
	handleError(err)
	
	// Используем стандартную ошибку
	handleError(fmt.Errorf("стандартная ошибка"))
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("Произошла ошибка: %v\n", err)
	}
}

// Repository интерфейс для репозитория данных
type Repository interface {
	Save(data string) error
	Load(id string) (string, error)
	Delete(id string) error
}

// MemoryRepository реализация Repository в памяти
type MemoryRepository struct {
	data map[string]string
}

// NewMemoryRepository создает новый MemoryRepository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]string),
	}
}

// Save реализация Repository интерфейса
func (mr *MemoryRepository) Save(data string) error {
	// В реальной реализации здесь был бы уникальный ID
	id := fmt.Sprintf("id_%d", len(mr.data)+1)
	mr.data[id] = data
	fmt.Printf("Сохранено: %s с ID %s\n", data, id)
	return nil
}

// Load реализация Repository интерфейса
func (mr *MemoryRepository) Load(id string) (string, error) {
	if data, exists := mr.data[id]; exists {
		return data, nil
	}
	return "", fmt.Errorf("данные с ID %s не найдены", id)
}

// Delete реализация Repository интерфейса
func (mr *MemoryRepository) Delete(id string) error {
	if _, exists := mr.data[id]; exists {
		delete(mr.data, id)
		fmt.Printf("Удалено данные с ID %s\n", id)
		return nil
	}
	return fmt.Errorf("данные с ID %s не найдены", id)
}

// Пример 8: Интерфейсы для тестирования (DI)
func dependencyInjection() {
	fmt.Println("\n=== Интерфейсы для тестирования ===")
	
	// Используем репозиторий
	repo := NewMemoryRepository()
	
	// Сохраняем данные
	repo.Save("Первые данные")
	repo.Save("Вторые данные")
	
	// Загружаем данные
	if data, err := repo.Load("id_1"); err == nil {
		fmt.Printf("Загружены данные: %s\n", data)
	}
	
	// Удаляем данные
	repo.Delete("id_1")
	
	// Пытаемся загрузить удаленные данные
	if _, err := repo.Load("id_1"); err != nil {
		fmt.Printf("Ошибка загрузки: %v\n", err)
	}
}

func main() {
	basicInterfaces()
	interfaceComposition()
	polymorphism()
	standardInterfaces()
	emptyInterface()
	typeSwitch()
	errorInterface()
	dependencyInjection()
}