package controller

import (
	"api/model"
	"api/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type ProdutoController struct {
	service *service.ProdutoService
}

func NewProdutoController(service *service.ProdutoService) *ProdutoController {
	return &ProdutoController{service: service}
}

// ListarProdutos retorna todos os produtos
// @Summary Lista todos os produtos
// @Description Retorna a lista completa de produtos cadastrados
// @Tags produtos
// @Produce json
// @Success 200 {array} model.Produto
// @Router /produtos [get]
func (c *ProdutoController) ListarProdutos(w http.ResponseWriter, r *http.Request) {
	produtos, err := c.service.BuscarTodosProdutos(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, produtos)
}

// BuscarProdutoPorID retorna um produto específico
// @Summary Busca um produto por ID
// @Description Retorna os detalhes de um produto específico
// @Tags produtos
// @Produce json
// @Param id path string true "ID do Produto"
// @Success 200 {object} model.Produto
// @Failure 404 {string} string "Produto não encontrado"
// @Router /produtos/{id} [get]
func (c *ProdutoController) BuscarProdutoPorID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	produto, err := c.service.BuscarProdutoPorID(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Produto não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, produto)
}

// CriarProduto adiciona um novo produto
// @Summary Adiciona um novo produto
// @Description Cria um novo produto no sistema
// @Tags produtos
// @Accept json
// @Produce json
// @Param produto body model.Produto true "Dados do Produto"
// @Success 201
// @Failure 400 {string} string "Dados inválidos"
// @Failure 409 {string} string "Produto já existe"
// @Router /produtos [post]
func (c *ProdutoController) CriarProduto(w http.ResponseWriter, r *http.Request) {
	var produto model.Produto
	if err := json.NewDecoder(r.Body).Decode(&produto); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if err := c.service.AdicionarProduto(r.Context(), produto); err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrDuplicate:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// AtualizarProduto atualiza um produto existente
// @Summary Atualiza um produto
// @Description Atualiza os dados de um produto existente
// @Tags produtos
// @Accept json
// @Produce json
// @Param id path string true "ID do Produto"
// @Param produto body model.Produto true "Dados atualizados do Produto"
// @Success 200
// @Failure 400 {string} string "Dados inválidos"
// @Failure 404 {string} string "Produto não encontrado"
// @Router /produtos/{id} [put]
func (c *ProdutoController) AtualizarProduto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var produto model.Produto
	if err := json.NewDecoder(r.Body).Decode(&produto); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if err := c.service.AtualizarProduto(r.Context(), id, produto); err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Produto não encontrado", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletarProduto remove um produto
// @Summary Remove um produto
// @Description Remove um produto do sistema
// @Tags produtos
// @Produce json
// @Param id path string true "ID do Produto"
// @Success 204
// @Failure 404 {string} string "Produto não encontrado"
// @Failure 400 {string} string "Produto está em pedidos"
// @Router /produtos/{id} [delete]
func (c *ProdutoController) DeletarProduto(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := c.service.DeletarProduto(r.Context(), id); err != nil {
		switch err {
		case service.ErrNotFound:
			http.Error(w, "Produto não encontrado", http.StatusNotFound)
		case service.ErrDependency:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AtualizarEstoque ajusta o estoque de um produto
// @Summary Atualiza estoque
// @Description Ajusta a quantidade em estoque de um produto (positivo para incrementar, negativo para decrementar)
// @Tags produtos
// @Accept json
// @Produce json
// @Param id path string true "ID do Produto"
// @Param quantidade body int true "Quantidade para ajuste"
// @Success 200
// @Failure 400 {string} string "Quantidade inválida"
// @Failure 404 {string} string "Produto não encontrado"
// @Router /produtos/{id}/estoque [patch]
func (c *ProdutoController) AtualizarEstoque(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var quantidade int
	if err := json.NewDecoder(r.Body).Decode(&quantidade); err != nil {
		http.Error(w, "Quantidade inválida", http.StatusBadRequest)
		return
	}

	if err := c.service.AtualizarEstoque(r.Context(), id, quantidade); err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Produto não encontrado", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CountProdutos retorna o número total de produtos
// @Summary Retorna a contagem total de produtos
// @Description Retorna o número total de produtos cadastrados no sistema
// @Tags produtos
// @Produce json
// @Success 200 {integer} integer "Número total de produtos"
// @Router /produtos/count [get]
func (c *ProdutoController) CountProdutos(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.CountProdutos(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]int{"total": count})
}

// BuscarProdutosPorNome busca produtos por nome
// @Summary Busca produtos por nome
// @Description Retorna os produtos cujos nomes correspondem ao parâmetro de busca
// @Tags produtos
// @Produce json
// @Param nome query string true "Nome ou parte do nome para busca"
// @Success 200 {array} model.Produto
// @Failure 400 {string} string "Nome não pode ser vazio"
// @Router /produtos/search [get]
func (c *ProdutoController) BuscarProdutosPorNome(w http.ResponseWriter, r *http.Request) {
	nome := r.URL.Query().Get("nome")
	if nome == "" {
		http.Error(w, "Parâmetro 'nome' é obrigatório", http.StatusBadRequest)
		return
	}

	produtos, err := c.service.BuscarProdutosPorNome(r.Context(), nome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, produtos)
}
