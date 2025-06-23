package service

import (
	"api/model"
	"api/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ClienteService struct {
	repo *repository.ClienteRepository
}

func NewClienteService(repo *repository.ClienteRepository) *ClienteService {
	return &ClienteService{repo: repo}
}

func (s *ClienteService) BuscarTodosClientes(ctx context.Context) ([]model.Cliente, error) {
	return s.repo.GetAll(ctx)
}

func (s *ClienteService) BuscarClientePorID(ctx context.Context, id string) (*model.Cliente, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ClienteService) AdicionarCliente(ctx context.Context, cliente model.Cliente) error {
	// Validações básicas
	if cliente.ID == "" {
		return fmt.Errorf("ID do cliente é obrigatório")
	}
	if cliente.Nome == "" {
		return fmt.Errorf("nome do cliente é obrigatório")
	}
	if cliente.Email == "" {
		return fmt.Errorf("email do cliente é obrigatório")
	}

	// Verificar se email já existe
	existente, err := s.repo.GetByEmail(ctx, cliente.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("erro ao verificar email existente: %w", err)
	}
	if existente != nil {
		return fmt.Errorf("email %s já está em uso", cliente.Email)
	}

	return s.repo.Add(ctx, cliente)
}

func (s *ClienteService) AtualizarCliente(ctx context.Context, id string, clienteAtualizado model.Cliente) error {
	// Validações básicas
	if clienteAtualizado.Nome == "" {
		return fmt.Errorf("nome do cliente é obrigatório")
	}
	if clienteAtualizado.Email == "" {
		return fmt.Errorf("email do cliente é obrigatório")
	}

	// Verificar se cliente existe
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("cliente com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar cliente: %w", err)
	}

	// Verificar se novo email já está em uso (por outro cliente)
	existente, err := s.repo.GetByEmail(ctx, clienteAtualizado.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("erro ao verificar email existente: %w", err)
	}
	if existente != nil && existente.ID != id {
		return fmt.Errorf("email %s já está em uso por outro cliente", clienteAtualizado.Email)
	}

	return s.repo.Update(ctx, id, clienteAtualizado)
}

func (s *ClienteService) DeletarCliente(ctx context.Context, id string) error {
	// Verificar se cliente existe antes de deletar
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("cliente com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar cliente: %w", err)
	}

	// Verificar se cliente tem pedidos associados
	temPedidos, err := s.repo.ClienteTemPedidos(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar pedidos do cliente: %w", err)
	}
	if temPedidos {
		return fmt.Errorf("não é possível deletar cliente com pedidos associados")
	}

	return s.repo.Delete(ctx, id)
}

func (s *ClienteService) CountClientes(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *ClienteService) BuscarClientesPorNome(ctx context.Context, nome string) ([]model.Cliente, error) {
	if nome == "" {
		return nil, fmt.Errorf("nome não pode ser vazio")
	}
	return s.repo.FindByName(ctx, nome)
}
