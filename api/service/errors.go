package service

import "errors"

// Erros customizados do serviço
var (
	// ErrNotFound indica que o recurso solicitado não foi encontrado
	ErrNotFound = errors.New("recurso não encontrado")

	// ErrInvalidInput indica que os dados de entrada são inválidos
	ErrInvalidInput = errors.New("dados de entrada inválidos")

	// ErrInsufficientStock indica que não há estoque suficiente para a operação
	ErrInsufficientStock = errors.New("estoque insuficiente")

	// ErrDependency indica que o recurso possui dependências e não pode ser excluído/modificado
	ErrDependency = errors.New("o recurso possui dependências e não pode ser processado")

	// ErrDuplicate indica que já existe um recurso com os mesmos dados
	ErrDuplicate = errors.New("recurso duplicado")

	// ErrInvalidOperation indica que a operação solicitada não é válida
	ErrInvalidOperation = errors.New("operação inválida")

	// ErrUnauthorized indica que o usuário não tem permissão para a operação
	ErrUnauthorized = errors.New("não autorizado")

	// ErrConflict indica um conflito na operação
	ErrConflict = errors.New("conflito na operação")
)

// ServiceError representa um erro customizado do serviço com detalhes adicionais
type ServiceError struct {
	// Código do erro (pode ser usado para mapear para HTTP status codes)
	Code string `json:"code"`

	// Mensagem do erro
	Message string `json:"message"`

	// Detalhes adicionais (opcional)
	Details interface{} `json:"details,omitempty"`

	// Erro original (para logging/debug)
	Err error `json:"-"`
}

// Error implementa a interface error
func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// NewServiceError cria um novo ServiceError
func NewServiceError(code, message string, err error) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// IsServiceError verifica se um erro é do tipo ServiceError
func IsServiceError(err error) bool {
	_, ok := err.(*ServiceError)
	return ok
}

// Helper functions para criar erros específicos
func NewNotFoundError(resource string, id interface{}) *ServiceError {
	return NewServiceError(
		"not_found",
		resource+" não encontrado(a)",
		nil,
	)
}

func NewValidationError(field, reason string) *ServiceError {
	return NewServiceError(
		"invalid_input",
		"validação falhou para o campo "+field,
		nil,
	)
}

func NewDuplicateError(resource string) *ServiceError {
	return NewServiceError(
		"duplicate",
		resource+" já existe",
		nil,
	)
}

func NewDependencyError(resource string) *ServiceError {
	return NewServiceError(
		"dependency",
		resource+" possui dependências e não pode ser processado",
		nil,
	)
}
