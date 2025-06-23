package repository

import (
	"api/model"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ClienteRepository struct {
	db *sqlx.DB
}

func NewClienteRepository(db *sqlx.DB) *ClienteRepository {
	return &ClienteRepository{db: db}
}

func (r *ClienteRepository) GetAll(ctx context.Context) ([]model.Cliente, error) {
	const query = `SELECT id, nome, email FROM clientes`
	var clientes []model.Cliente
	err := r.db.SelectContext(ctx, &clientes, query)
	return clientes, err
}

func (r *ClienteRepository) GetByID(ctx context.Context, id string) (*model.Cliente, error) {
	const query = `SELECT id, nome, email FROM clientes WHERE id = $1`
	var cliente model.Cliente
	err := r.db.GetContext(ctx, &cliente, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("erro ao buscar cliente: %w", err)
	}
	return &cliente, nil
}

func (r *ClienteRepository) GetByEmail(ctx context.Context, email string) (*model.Cliente, error) {
	const query = `SELECT id, nome, email FROM clientes WHERE email = $1`
	var cliente model.Cliente
	err := r.db.GetContext(ctx, &cliente, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("erro ao buscar cliente por email: %w", err)
	}
	return &cliente, nil
}

func (r *ClienteRepository) Add(ctx context.Context, cliente model.Cliente) error {
	const query = `INSERT INTO clientes (id, nome, email) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, cliente.ID, cliente.Nome, cliente.Email)
	if err != nil {
		return fmt.Errorf("erro ao inserir cliente: %w", err)
	}
	return nil
}

func (r *ClienteRepository) Update(ctx context.Context, id string, cliente model.Cliente) error {
	const query = `UPDATE clientes SET nome = $1, email = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, cliente.Nome, cliente.Email, id)
	if err != nil {
		return fmt.Errorf("erro ao atualizar cliente: %w", err)
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

func (r *ClienteRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM clientes WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("erro ao deletar cliente: %w", err)
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

func (r *ClienteRepository) ClienteTemPedidos(ctx context.Context, clienteID string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM pedidos WHERE cliente_id = $1)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, clienteID)
	if err != nil {
		return false, fmt.Errorf("erro ao verificar pedidos do cliente: %w", err)
	}
	return exists, nil
}

func (r *ClienteRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return r.db.BeginTxx(ctx, nil)
}

func (r *ClienteRepository) Count(ctx context.Context) (int, error) {
	const query = `SELECT COUNT(*) as count FROM clientes`
	var result struct {
		Count int `db:"count"`
	}

	err := r.db.GetContext(ctx, &result, query)
	if err != nil {
		// Mesmo sem registros, COUNT deve retornar 0, não um erro
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("erro ao contar clientes: %w", err)
	}
	return result.Count, nil
}

func (r *ClienteRepository) FindByName(ctx context.Context, name string) ([]model.Cliente, error) {
	const query = `SELECT id, nome, email FROM clientes WHERE nome LIKE $1`
	var clientes []model.Cliente
	err := r.db.SelectContext(ctx, &clientes, query, "%"+name+"%")
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar clientes por nome: %w", err)
	}
	return clientes, nil
}
