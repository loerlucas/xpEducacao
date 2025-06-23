package repository

import (
	"api/model"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type PedidoRepository struct {
	db *sqlx.DB
}

func NewPedidoRepository(db *sqlx.DB) *PedidoRepository {
	return &PedidoRepository{db: db}
}

func (r *PedidoRepository) GetAll(ctx context.Context) ([]model.Pedido, error) {
	const query = `
        SELECT 
            p.id,
            p.cliente_id,
            p.data,
            p.total,
            p.status
        FROM pedidos p
        ORDER BY p.data DESC
    `

	var pedidos []model.Pedido
	err := r.db.SelectContext(ctx, &pedidos, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos: %w", err)
	}

	// Carrega os itens para cada pedido
	for i := range pedidos {
		itens, err := r.getItensPedido(ctx, pedidos[i].ID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar itens do pedido %s: %w", pedidos[i].ID, err)
		}
		pedidos[i].Itens = itens
	}

	return pedidos, nil
}

func (r *PedidoRepository) GetByID(ctx context.Context, id string) (*model.Pedido, error) {
	const query = `SELECT id, cliente_id, data, total, status FROM pedidos WHERE id = $1`
	var pedido model.Pedido
	err := r.db.GetContext(ctx, &pedido, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("erro ao buscar pedido: %w", err)
	}

	itens, err := r.getItensPedido(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar itens do pedido: %w", err)
	}
	pedido.Itens = itens

	return &pedido, nil
}

func (r *PedidoRepository) getItensPedido(ctx context.Context, pedidoID string) ([]model.ItemPedido, error) {
	const query = `
        SELECT 
            produto_id AS "produto_id",
            quantidade AS "quantidade",
            preco_unit AS "preco_unit",
            subtotal AS "subtotal"
        FROM itens_pedido
        WHERE pedido_id = $1
    `

	var itens []model.ItemPedido
	err := r.db.SelectContext(ctx, &itens, query, pedidoID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar itens do pedido: %w", err)
	}
	return itens, nil
}

func (r *PedidoRepository) AddWithTx(ctx context.Context, tx *sqlx.Tx, pedido model.Pedido) error {
	const pedidoQuery = `INSERT INTO pedidos (id, cliente_id, data, total, status) 
		VALUES ($1, $2, $3, $4, $5)`
	_, err := tx.ExecContext(ctx, pedidoQuery,
		pedido.ID,
		pedido.ClienteID,
		pedido.Data,
		pedido.Total,
		pedido.Status)
	if err != nil {
		return fmt.Errorf("erro ao inserir pedido: %w", err)
	}

	// Inserir itens do pedido
	const itemQuery = `INSERT INTO itens_pedido 
		(pedido_id, produto_id, quantidade, preco_unit, subtotal) 
		VALUES ($1, $2, $3, $4, $5)`
	for _, item := range pedido.Itens {
		_, err := tx.ExecContext(ctx, itemQuery,
			pedido.ID,
			item.ProdutoID,
			item.Quantidade,
			item.PrecoUnit,
			item.Subtotal)
		if err != nil {
			return fmt.Errorf("erro ao inserir item do pedido: %w", err)
		}
	}

	return nil
}

func (r *PedidoRepository) Update(ctx context.Context, id string, pedido model.Pedido) error {
	const query = `UPDATE pedidos SET 
		cliente_id = $1, 
		data = $2, 
		total = $3, 
		status = $4 
		WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query,
		pedido.ClienteID,
		pedido.Data,
		pedido.Total,
		pedido.Status,
		id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar pedido: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PedidoRepository) UpdateWithTx(ctx context.Context, tx *sqlx.Tx, id string, pedido model.Pedido) error {
	const query = `UPDATE pedidos SET 
		cliente_id = $1, 
		data = $2, 
		total = $3, 
		status = $4 
		WHERE id = $5`
	result, err := tx.ExecContext(ctx, query,
		pedido.ClienteID,
		pedido.Data,
		pedido.Total,
		pedido.Status,
		id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar pedido: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PedidoRepository) Delete(ctx context.Context, id string) error {
	// Primeiro deletar os itens do pedido
	const deleteItensQuery = `DELETE FROM itens_pedido WHERE pedido_id = $1`
	_, err := r.db.ExecContext(ctx, deleteItensQuery, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar itens do pedido: %w", err)
	}

	// Depois deletar o pedido
	const deletePedidoQuery = `DELETE FROM pedidos WHERE id = $1`
	result, err := r.db.ExecContext(ctx, deletePedidoQuery, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar pedido: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *PedidoRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *PedidoRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM pedidos`
	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("erro ao contar pedidos: %w", err)
	}
	return count, nil
}

func (r *PedidoRepository) FindByClienteName(ctx context.Context, nome string) ([]model.Pedido, error) {
	const query = `
        SELECT p.id, p.cliente_id, p.data, p.total, p.status
        FROM pedidos p
        JOIN clientes c ON p.cliente_id = c.id
        WHERE c.nome LIKE $1
        ORDER BY p.data DESC
    `

	var pedidos []model.Pedido
	err := r.db.SelectContext(ctx, &pedidos, query, "%"+nome+"%")
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar pedidos por nome do cliente: %w", err)
	}

	// Carrega os itens para cada pedido
	for i := range pedidos {
		itens, err := r.getItensPedido(ctx, pedidos[i].ID)
		if err != nil {
			return nil, fmt.Errorf("erro ao buscar itens do pedido %s: %w", pedidos[i].ID, err)
		}
		pedidos[i].Itens = itens
	}

	return pedidos, nil
}
