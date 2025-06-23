package controller

import (
	"api/model"
	"api/service"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type PedidoController struct {
	service *service.PedidoService
}

func NewPedidoController(service *service.PedidoService) *PedidoController {
	return &PedidoController{service: service}
}

// ListarPedidos retorna todos os pedidos
// @Summary Lista todos os pedidos
// @Description Retorna a lista completa de pedidos cadastrados
// @Tags pedidos
// @Produce json
// @Success 200 {array} model.Pedido
// @Router /pedidos [get]
func (c *PedidoController) ListarPedidos(w http.ResponseWriter, r *http.Request) {
	pedidos, err := c.service.BuscarTodosPedidos(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, pedidos)
}

// BuscarPedidoPorID retorna um pedido específico
// @Summary Busca um pedido por ID
// @Description Retorna os detalhes de um pedido específico
// @Tags pedidos
// @Produce json
// @Param id path string true "ID do Pedido"
// @Success 200 {object} model.Pedido
// @Failure 404 {string} string "Pedido não encontrado"
// @Router /pedidos/{id} [get]
func (c *PedidoController) BuscarPedidoPorID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	pedido, err := c.service.BuscarPedidoPorID(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Pedido não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, pedido)
}

// CriarPedido adiciona um novo pedido
// @Summary Adiciona um novo pedido
// @Description Cria um novo pedido no sistema
// @Tags pedidos
// @Accept json
// @Produce json
// @Param pedido body model.Pedido true "Dados do Pedido"
// @Success 201
// @Failure 400 {string} string "Dados inválidos"
// @Failure 404 {string} string "Cliente ou produto não encontrado"
// @Failure 422 {string} string "Estoque insuficiente"
// @Router /pedidos [post]
func (c *PedidoController) CriarPedido(w http.ResponseWriter, r *http.Request) {
	var pedido model.Pedido
	if err := json.NewDecoder(r.Body).Decode(&pedido); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if err := c.service.AdicionarPedido(r.Context(), pedido); err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case service.ErrInsufficientStock:
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// AtualizarStatusPedido altera o status de um pedido
// @Summary Atualiza status do pedido
// @Description Altera o status de um pedido existente
// @Tags pedidos
// @Accept json
// @Produce json
// @Param id path string true "ID do Pedido"
// @Param status body string true "Novo status"
// @Success 200
// @Failure 400 {string} string "Status inválido"
// @Failure 404 {string} string "Pedido não encontrado"
// @Router /pedidos/{id}/status [put]
func (c *PedidoController) AtualizarStatusPedido(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var status struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, "Status inválido", http.StatusBadRequest)
		return
	}

	if err := c.service.AtualizarStatusPedido(r.Context(), id, status.Status); err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Pedido não encontrado", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CancelarPedido cancela um pedido existente
// @Summary Cancela um pedido
// @Description Cancela um pedido e devolve os produtos ao estoque
// @Tags pedidos
// @Produce json
// @Param id path string true "ID do Pedido"
// @Success 200
// @Failure 404 {string} string "Pedido não encontrado"
// @Router /pedidos/{id}/cancelar [post]
func (c *PedidoController) CancelarPedido(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := c.service.CancelarPedido(r.Context(), id); err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Pedido não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletarPedido remove um pedido
// @Summary Remove um pedido
// @Description Remove um pedido do sistema (apenas pedidos cancelados)
// @Tags pedidos
// @Produce json
// @Param id path string true "ID do Pedido"
// @Success 204
// @Failure 404 {string} string "Pedido não encontrado"
// @Failure 400 {string} string "Pedido não pode ser deletado"
// @Router /pedidos/{id} [delete]
func (c *PedidoController) DeletarPedido(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := c.service.DeletarPedido(r.Context(), id); err != nil {
		switch err {
		case service.ErrNotFound:
			http.Error(w, "Pedido não encontrado", http.StatusNotFound)
		case service.ErrInvalidOperation:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CountPedidos retorna o número total de pedidos
// @Summary Retorna a contagem total de pedidos
// @Description Retorna o número total de pedidos cadastrados no sistema
// @Tags pedidos
// @Produce json
// @Success 200 {integer} integer "Número total de pedidos"
// @Router /pedidos/count [get]
func (c *PedidoController) CountPedidos(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.CountPedidos(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]int{"total": count})
}

// BuscarPedidosPorNomeCliente busca pedidos por nome do cliente
// @Summary Busca pedidos por nome do cliente
// @Description Retorna os pedidos cujos clientes têm nomes que correspondem ao parâmetro de busca
// @Tags pedidos
// @Produce json
// @Param nome query string true "Nome ou parte do nome do cliente para busca"
// @Success 200 {array} model.Pedido
// @Failure 400 {string} string "Nome não pode ser vazio"
// @Router /pedidos/search [get]
func (c *PedidoController) BuscarPedidosPorNomeCliente(w http.ResponseWriter, r *http.Request) {
	nome := r.URL.Query().Get("nome")
	if nome == "" {
		http.Error(w, "Parâmetro 'nome' é obrigatório", http.StatusBadRequest)
		return
	}

	pedidos, err := c.service.BuscarPedidosPorNomeCliente(r.Context(), nome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, pedidos)
}
