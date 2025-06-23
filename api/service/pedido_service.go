package service

import (
	"api/model"
	"api/repository"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PedidoService struct {
	pedidoRepo  *repository.PedidoRepository
	clienteRepo *repository.ClienteRepository
	produtoRepo *repository.ProdutoRepository
	estoqueSvc  *ProdutoService
}

func NewPedidoService(
	pedidoRepo *repository.PedidoRepository,
	clienteRepo *repository.ClienteRepository,
	produtoRepo *repository.ProdutoRepository,
	estoqueSvc *ProdutoService,
) *PedidoService {
	return &PedidoService{
		pedidoRepo:  pedidoRepo,
		clienteRepo: clienteRepo,
		produtoRepo: produtoRepo,
		estoqueSvc:  estoqueSvc,
	}
}

func (s *PedidoService) BuscarTodosPedidos(ctx context.Context) ([]model.Pedido, error) {
	return s.pedidoRepo.GetAll(ctx)
}

func (s *PedidoService) BuscarPedidoPorID(ctx context.Context, id string) (*model.Pedido, error) {
	return s.pedidoRepo.GetByID(ctx, id)
}

func (s *PedidoService) AdicionarPedido(ctx context.Context, pedido model.Pedido) error {
	// Validações básicas
	if pedido.ID == "" {
		return fmt.Errorf("ID do pedido é obrigatório")
	}
	if pedido.ClienteID == "" {
		return fmt.Errorf("cliente_id é obrigatório")
	}
	if len(pedido.Itens) == 0 {
		return fmt.Errorf("pedido deve conter pelo menos um item")
	}
	if pedido.Status == "" {
		return fmt.Errorf("status do pedido é obrigatório")
	}

	// Verificar se cliente existe
	_, err := s.clienteRepo.GetByID(ctx, pedido.ClienteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("cliente com ID %s não encontrado", pedido.ClienteID)
		}
		return fmt.Errorf("erro ao verificar cliente: %w", err)
	}

	// Validar itens e calcular total
	var totalCalculado float64
	produtosMap := make(map[string]*model.Produto)

	for i, item := range pedido.Itens {
		// Buscar produto
		produto, err := s.produtoRepo.GetByID(ctx, item.ProdutoID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("produto com ID %s não encontrado", item.ProdutoID)
			}
			return fmt.Errorf("erro ao buscar produto %s: %w", item.ProdutoID, err)
		}

		// Validar quantidade
		if item.Quantidade <= 0 {
			return fmt.Errorf("quantidade inválida para o produto %s", produto.Nome)
		}

		// Verificar estoque
		if produto.Estoque < item.Quantidade {
			return fmt.Errorf("estoque insuficiente para o produto %s (disponível: %d, solicitado: %d)",
				produto.Nome, produto.Estoque, item.Quantidade)
		}

		// Calcular valores
		pedido.Itens[i].PrecoUnit = produto.Preco
		pedido.Itens[i].Subtotal = produto.Preco * float64(item.Quantidade)
		totalCalculado += pedido.Itens[i].Subtotal
		produtosMap[produto.ID] = produto
	}

	// Validar total
	if pedido.Total != totalCalculado {
		return fmt.Errorf("total do pedido (%.2f) não corresponde à soma dos itens (%.2f)",
			pedido.Total, totalCalculado)
	}

	// Definir data atual se não informada
	if pedido.Data == "" {
		pedido.Data = time.Now().Format(time.RFC3339)
	}

	// Usar transação para garantir atomicidade
	tx, err := s.pedidoRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Adicionar pedido
	if err := s.pedidoRepo.AddWithTx(ctx, tx, pedido); err != nil {
		return fmt.Errorf("erro ao adicionar pedido: %w", err)
	}

	// Atualizar estoque dos produtos
	for _, item := range pedido.Itens {
		if err := s.estoqueSvc.AtualizarEstoque(ctx, item.ProdutoID, -item.Quantidade); err != nil {
			return fmt.Errorf("erro ao atualizar estoque do produto %s: %w", item.ProdutoID, err)
		}
	}

	// Confirmar transação
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return nil
}

func (s *PedidoService) AtualizarStatusPedido(ctx context.Context, id string, novoStatus string) error {
	// Verificar se pedido existe
	pedido, err := s.pedidoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("pedido com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	// Validar novo status
	if novoStatus == "" {
		return fmt.Errorf("novo status é obrigatório")
	}

	// Atualizar apenas o status
	pedido.Status = novoStatus
	return s.pedidoRepo.Update(ctx, id, *pedido)
}

func (s *PedidoService) CancelarPedido(ctx context.Context, id string) error {
	// Verificar se pedido existe
	pedido, err := s.pedidoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("pedido com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	// Verificar se já está cancelado
	if pedido.Status == "Cancelado" {
		return nil
	}

	// Usar transação para garantir atomicidade
	tx, err := s.pedidoRepo.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação: %w", err)
	}
	defer tx.Rollback()

	// Atualizar status do pedido
	pedido.Status = "Cancelado"
	if err := s.pedidoRepo.UpdateWithTx(ctx, tx, id, *pedido); err != nil {
		return fmt.Errorf("erro ao atualizar pedido: %w", err)
	}

	// Devolver produtos ao estoque
	for _, item := range pedido.Itens {
		if err := s.estoqueSvc.AtualizarEstoque(ctx, item.ProdutoID, item.Quantidade); err != nil {
			return fmt.Errorf("erro ao devolver estoque do produto %s: %w", item.ProdutoID, err)
		}
	}

	// Confirmar transação
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao confirmar transação: %w", err)
	}

	return nil
}

func (s *PedidoService) DeletarPedido(ctx context.Context, id string) error {
	// Verificar se pedido existe
	_, err := s.pedidoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("pedido com ID %s não encontrado", id)
		}
		return fmt.Errorf("erro ao buscar pedido: %w", err)
	}
	// Verificar se o pedido está cancelado
	return s.pedidoRepo.Delete(ctx, id)
}

func (s *PedidoService) CountPedidos(ctx context.Context) (int, error) {
	return s.pedidoRepo.Count(ctx)
}

func (s *PedidoService) BuscarPedidosPorNomeCliente(ctx context.Context, nome string) ([]model.Pedido, error) {
	if nome == "" {
		return nil, fmt.Errorf("nome do cliente não pode ser vazio")
	}
	return s.pedidoRepo.FindByClienteName(ctx, nome)
}
