package model

type Produto struct {
	ID        string  `json:"id"`
	Nome      string  `json:"nome"`
	Descricao string  `json:"descricao"`
	Preco     float64 `json:"preco"`
	Estoque   int     `json:"estoque"`
	Categoria string  `json:"categoria"`
}
