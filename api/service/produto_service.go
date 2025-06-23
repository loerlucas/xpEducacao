package service

import (
	"api/model"
	"api/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type ProdutoService struct {
	repo *repository.ProdutoRepository
}

func NewProdutoService(repo *repository.ProdutoRepository) *ProdutoService {
	return &ProdutoService{repo: repo}
}

func (s *ProdutoService) BuscarTodosProdutos(ctx context.Context) ([]model.Produto, error) {
	return s.repo.GetAll(ctx)
}

func (s *ProdutoService) BuscarProdutoPorID(ctx context.Context, id string) (*model.Produto, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProdutoService) AdicionarProduto(ctx context.Context, produto model.Produto) error {
	// Validações básicas
	if produto.ID == "" {
		return fmt.Errorf("ID do produto é obrigatório")
	}
	if produto.Nome == "" {
		return fmt.Errorf("nome do produto é obrigatório")
	}
	if produto.Preco <= 0 {
		return fmt.Errorf("preço do produto deve ser maior que zero")
	}
	if produto.Estoque < 0 {
		return fmt.Errorf("estoque do produto não pode ser negativo")
	}

	// Verificar se produto com mesmo ID já existe
	_, err := s.repo.GetByID(ctx, produto.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("erro ao verificar produto existente: %w", err)
	}
	if err == nil {
		return fmt.Errorf("produto com ID %s já existe", produto.ID)
	}

	return s.repo.Add(ctx, produto)
}

func (s *ProdutoService) AtualizarProduto(ctx context.Context, id string, produtoAtualizado model.Produto) error {
	// Validações básicas
	if produtoAtualizado.Nome == "" {
		return fmt.Errorf("nome do produto é obrigatório")
	}
	if produtoAtualizado.Preco <= 0 {
		return fmt.Errorf("preço do produto deve ser maior que zero")
	}
	if produtoAtualizado.Estoque < 0 {
		return fmt.Errorf("estoque do produto não pode ser negativo")
	}

	// Verificar se produto existe
	produtoExistente, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("produto com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar produto: %w", err)
	}

	// Manter o ID original
	produtoAtualizado.ID = produtoExistente.ID

	return s.repo.Update(ctx, id, produtoAtualizado)
}

func (s *ProdutoService) DeletarProduto(ctx context.Context, id string) error {
	// Verificar se produto existe
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("produto com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar produto: %w", err)
	}

	// Verificar se produto está em algum pedido
	emPedidos, err := s.repo.ProdutoEmPedidos(ctx, id)
	if err != nil {
		return fmt.Errorf("erro ao verificar pedidos do produto: %w", err)
	}
	if emPedidos {
		return fmt.Errorf("não é possível deletar produto associado a pedidos")
	}

	return s.repo.Delete(ctx, id)
}

func (s *ProdutoService) AtualizarEstoque(ctx context.Context, id string, quantidade int) error {
	if quantidade == 0 {
		return nil
	}

	// Verificar se produto existe
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("produto com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar produto: %w", err)
	}

	if quantidade > 0 {
		return s.repo.IncrementarEstoque(ctx, id, quantidade)
	}
	return s.repo.DecrementarEstoque(ctx, id, -quantidade)
}

func (s *ProdutoService) CountProdutos(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *ProdutoService) BuscarProdutosPorNome(ctx context.Context, nome string) ([]model.Produto, error) {
	if nome == "" {
		return nil, fmt.Errorf("nome não pode ser vazio")
	}
	return s.repo.FindByName(ctx, nome)
}
