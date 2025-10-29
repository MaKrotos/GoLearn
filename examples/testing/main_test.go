package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Calculator простой калькулятор для тестирования
type Calculator struct{}

// Add складывает два числа
func (c Calculator) Add(a, b int) int {
	return a + b
}

// Subtract вычитает второе число из первого
func (c Calculator) Subtract(a, b int) int {
	return a - b
}

// Multiply умножает два числа
func (c Calculator) Multiply(a, b int) int {
	return a * b
}

// Divide делит первое число на второе
func (c Calculator) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("деление на ноль")
	}
	return a / b, nil
}

// Пример 1: Базовое тестирование
func TestCalculator_Add(t *testing.T) {
	calc := Calculator{}
	
	// Тестовый случай
	result := calc.Add(2, 3)
	expected := 5
	
	if result != expected {
		t.Errorf("Add(2, 3) = %d; expected %d", result, expected)
	}
}

// Пример 2: Табличные тесты
func TestCalculator_Subtract(t *testing.T) {
	calc := Calculator{}
	
	// Таблица тестовых случаев
	tests := []struct {
		a        int
		b        int
		expected int
		name     string
	}{
		{5, 3, 2, "positive numbers"},
		{0, 0, 0, "zeros"},
		{-5, -3, -2, "negative numbers"},
		{10, -5, 15, "mixed signs"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Subtract(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Subtract(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Пример 3: Тестирование с ошибками
func TestCalculator_Divide(t *testing.T) {
	calc := Calculator{}
	
	t.Run("normal division", func(t *testing.T) {
		result, err := calc.Divide(10, 2)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if result != 5 {
			t.Errorf("Divide(10, 2) = %d; expected 5", result)
		}
	})
	
	t.Run("division by zero", func(t *testing.T) {
		_, err := calc.Divide(10, 0)
		if err == nil {
			t.Fatal("Expected error for division by zero, but got none")
		}
		if err.Error() != "деление на ноль" {
			t.Errorf("Expected 'деление на ноль' error, got %v", err)
		}
	})
}

// Пример 4: Параллельное тестирование
func TestCalculator_Multiply(t *testing.T) {
	calc := Calculator{}
	
	// Запускаем тест параллельно
	t.Parallel()
	
	tests := []struct {
		a, b, expected int
	}{
		{2, 3, 6},
		{0, 5, 0},
		{-2, 3, -6},
		{-2, -3, 6},
	}
	
	for _, tt := range tests {
		tt := tt // Захватываем переменную для параллельного теста
		t.Run(fmt.Sprintf("%dx%d", tt.a, tt.b), func(t *testing.T) {
			t.Parallel()
			result := calc.Multiply(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Multiply(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// Пример 5: httptest для тестирования HTTP обработчиков
func TestHTTPHandlers(t *testing.T) {
	// Создаем тестовый сервер
	handler := http.NewServeMux()
	handler.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok"}`)
	})
	
	handler.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"id": 1, "name": "John"}]`)
	})
	
	// Тест для /health endpoint
	t.Run("health check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		
		handler.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		expected := `{"status": "ok"}`
		if w.Body.String() != expected {
			t.Errorf("Expected body %s, got %s", expected, w.Body.String())
		}
	})
	
	// Тест для /api/users endpoint
	t.Run("get users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		w := httptest.NewRecorder()
		
		handler.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}
	})
	
	// Тест для неразрешенного метода
	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
		w := httptest.NewRecorder()
		
		handler.ServeHTTP(w, req)
		
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status 405, got %d", w.Code)
		}
	})
}

// Пример 6: Mock объекты
type UserRepository interface {
	GetUser(id int) (string, error)
	SaveUser(id int, name string) error
}

type MockUserRepository struct {
	users map[int]string
}

func (m *MockUserRepository) GetUser(id int) (string, error) {
	if name, exists := m.users[id]; exists {
		return name, nil
	}
	return "", fmt.Errorf("user not found")
}

func (m *MockUserRepository) SaveUser(id int, name string) error {
	if m.users == nil {
		m.users = make(map[int]string)
	}
	m.users[id] = name
	return nil
}

// UserService сервис, использующий UserRepository
type UserService struct {
	repo UserRepository
}

func (s *UserService) GetUserName(id int) (string, error) {
	return s.repo.GetUser(id)
}

func (s *UserService) CreateUserName(id int, name string) error {
	return s.repo.SaveUser(id, name)
}

func TestUserService(t *testing.T) {
	// Создаем mock репозиторий
	mockRepo := &MockUserRepository{
		users: map[int]string{1: "John"},
	}
	
	// Создаем сервис с mock репозиторием
	service := &UserService{repo: mockRepo}
	
	t.Run("get existing user", func(t *testing.T) {
		name, err := service.GetUserName(1)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if name != "John" {
			t.Errorf("Expected 'John', got '%s'", name)
		}
	})
	
	t.Run("get non-existing user", func(t *testing.T) {
		_, err := service.GetUserName(999)
		if err == nil {
			t.Fatal("Expected error for non-existing user, but got none")
		}
	})
	
	t.Run("create user", func(t *testing.T) {
		err := service.CreateUserName(2, "Jane")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		// Проверяем, что пользователь был создан
		name, err := service.GetUserName(2)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if name != "Jane" {
			t.Errorf("Expected 'Jane', got '%s'", name)
		}
	})
}

// Пример 7: Бенчмарки
func BenchmarkCalculator_Add(b *testing.B) {
	calc := Calculator{}
	
	// Запускаем бенчмарк
	for i := 0; i < b.N; i++ {
		calc.Add(1, 2)
	}
}

func BenchmarkCalculator_Multiply(b *testing.B) {
	calc := Calculator{}
	
	for i := 0; i < b.N; i++ {
		calc.Multiply(i, 2)
	}
}

// Пример 8: Примеры для fuzz тестирования (Go 1.18+)
// func FuzzCalculator_Add(f *testing.F) {
// 	// Добавляем seed corpus
// 	f.Add(1, 2)
// 	f.Add(0, 0)
// 	f.Add(-1, 1)
// 	
// 	calc := Calculator{}
// 	
// 	f.Fuzz(func(t *testing.T, a, b int) {
// 		result := calc.Add(a, b)
// 		// Проверяем, что результат корректен
// 		if result != a+b {
// 			t.Errorf("Add(%d, %d) = %d; expected %d", a, b, result, a+b)
// 		}
// 	})
// }

// Пример 9: Тестирование с использованием testify (внешняя библиотека)
// Для использования нужно выполнить: go get github.com/stretchr/testify
//
// import (
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )
//
// func TestCalculator_WithTestify(t *testing.T) {
// 	calc := Calculator{}
// 	
// 	// Использование assert
// 	result := calc.Add(2, 3)
// 	assert.Equal(t, 5, result, "2 + 3 should equal 5")
// 	
// 	// Использование require
// 	result, err := calc.Divide(10, 2)
// 	require.NoError(t, err, "Division should not error")
// 	assert.Equal(t, 5, result, "10 / 2 should equal 5")
// 	
// 	// Тест с ошибкой
// 	_, err = calc.Divide(10, 0)
// 	assert.Error(t, err, "Division by zero should error")
// 	assert.Equal(t, "деление на ноль", err.Error())
// }

// Пример 10: Setup и teardown
func setupTest() *Calculator {
	fmt.Println("Setup test")
	return Calculator{}
}

func teardownTest() {
	fmt.Println("Teardown test")
}

func TestCalculator_WithSetupTeardown(t *testing.T) {
	calc := setupTest()
	defer teardownTest()
	
	result := calc.Add(1, 1)
	if result != 2 {
		t.Errorf("Expected 2, got %d", result)
	}
}

// Пример 11: Тестирование с подтестами и разными сценариями
func TestCalculator_Comprehensive(t *testing.T) {
	calc := Calculator{}
	
	tests := map[string]struct {
		a, b     int
		op       string
		expected int
		hasError bool
	}{
		"add positive":     {a: 2, b: 3, op: "add", expected: 5, hasError: false},
		"add negative":     {a: -2, b: -3, op: "add", expected: -5, hasError: false},
		"subtract positive": {a: 5, b: 3, op: "sub", expected: 2, hasError: false},
		"multiply positive": {a: 3, b: 4, op: "mul", expected: 12, hasError: false},
		"divide normal":     {a: 10, b: 2, op: "div", expected: 5, hasError: false},
		"divide by zero":    {a: 10, b: 0, op: "div", expected: 0, hasError: true},
	}
	
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var result int
			var err error
			
			switch tc.op {
			case "add":
				result = calc.Add(tc.a, tc.b)
			case "sub":
				result = calc.Subtract(tc.a, tc.b)
			case "mul":
				result = calc.Multiply(tc.a, tc.b)
			case "div":
				result, err = calc.Divide(tc.a, tc.b)
			}
			
			if tc.hasError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tc.expected {
					t.Errorf("Expected %d, got %d", tc.expected, result)
				}
			}
		})
	}
}