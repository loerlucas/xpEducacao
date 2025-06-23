package model

type Pedido struct {
	ID        string       `json:"id" db:"id"`
	ClienteID string       `json:"cliente_id" db:"cliente_id"`
	Data      string       `json:"data" db:"data"`
	Total     float64      `json:"total" db:"total"`
	Status    string       `json:"status" db:"status"`
	Itens     []ItemPedido `json:"itens"`
}
