package repository

import (
	"api/model"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ProdutoRepository struct {
	db *sqlx.DB
}

func NewProdutoRepository(db *sqlx.DB) *ProdutoRepository {
	return &ProdutoRepository{db: db}
}

func (r *ProdutoRepository) GetAll(ctx context.Context) ([]model.Produto, error) {
	const query = `SELECT id, nome, descricao, preco, estoque, categoria FROM produtos ORDER BY nome`
	var produtos []model.Produto
	err := r.db.SelectContext(ctx, &produtos, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar produtos: %w", err)
	}
	return produtos, nil
}

func (r *ProdutoRepository) GetByID(ctx context.Context, id string) (*model.Produto, error) {
	const query = `SELECT id, nome, descricao, preco, estoque, categoria FROM produtos WHERE id = $1`
	var produto model.Produto
	err := r.db.GetContext(ctx, &produto, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("erro ao buscar produto: %w", err)
	}
	return &produto, nil
}

func (r *ProdutoRepository) Add(ctx context.Context, produto model.Produto) error {
	const query = `INSERT INTO produtos (id, nome, descricao, preco, estoque, categoria) 
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query,
		produto.ID,
		produto.Nome,
		produto.Descricao,
		produto.Preco,
		produto.Estoque,
		produto.Categoria)
	if err != nil {
		return fmt.Errorf("erro ao inserir produto: %w", err)
	}
	return nil
}

func (r *ProdutoRepository) Update(ctx context.Context, id string, produto model.Produto) error {
	const query = `UPDATE produtos SET 
		nome = $1, 
		descricao = $2, 
		preco = $3, 
		estoque = $4, 
		categoria = $5 
		WHERE id = $6`
	result, err := r.db.ExecContext(ctx, query,
		produto.Nome,
		produto.Descricao,
		produto.Preco,
		produto.Estoque,
		produto.Categoria,
		id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar produto: %w", err)
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

func (r *ProdutoRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM produtos WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar produto: %w", err)
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

func (r *ProdutoRepository) IncrementarEstoque(ctx context.Context, id string, quantidade int) error {
	const query = `UPDATE produtos SET estoque = estoque + $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, quantidade, id)
	if err != nil {
		return fmt.Errorf("erro ao incrementar estoque: %w", err)
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

func (r *ProdutoRepository) DecrementarEstoque(ctx context.Context, id string, quantidade int) error {
	const query = `UPDATE produtos SET estoque = estoque - $1 
		WHERE id = $2 AND estoque >= $1`
	result, err := r.db.ExecContext(ctx, query, quantidade, id)
	if err != nil {
		return fmt.Errorf("erro ao decrementar estoque: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("estoque insuficiente ou produto não encontrado")
	}

	return nil
}

func (r *ProdutoRepository) ProdutoEmPedidos(ctx context.Context, produtoID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM itens_pedido WHERE produto_id = $1)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, produtoID)
	if err != nil {
		return false, fmt.Errorf("erro ao verificar pedidos do produto: %w", err)
	}
	return exists, nil
}

func (r *ProdutoRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) FROM produtos`
	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("erro ao contar produtos: %w", err)
	}
	return count, nil
}

func (r *ProdutoRepository) FindByName(ctx context.Context, name string) ([]model.Produto, error) {
	const query = `SELECT id, nome, descricao, preco, estoque FROM produtos WHERE nome LIKE $1`
	var produtos []model.Produto
	err := r.db.SelectContext(ctx, &produtos, query, "%"+name+"%")
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar produtos por nome: %w", err)
	}
	return produtos, nil
}
