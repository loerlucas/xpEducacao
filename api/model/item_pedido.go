package model

type ItemPedido struct {
	ProdutoID  string  `json:"produto_id" db:"produto_id"`
	Quantidade int     `json:"quantidade" db:"quantidade"`
	PrecoUnit  float64 `json:"preco_unit" db:"preco_unit"`
	Subtotal   float64 `json:"subtotal" db:"subtotal"`
}
