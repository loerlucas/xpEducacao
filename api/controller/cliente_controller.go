package controller

import (
	"api/model"
	"api/service"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type ClienteController struct {
	service *service.ClienteService
}

func NewClienteController(service *service.ClienteService) *ClienteController {
	return &ClienteController{service: service}
}

// ListarClientes retorna todos os clientes
// @Summary Lista todos os clientes
// @Description Retorna a lista completa de clientes cadastrados
// @Tags clientes
// @Produce json
// @Success 200 {array} model.Cliente
// @Router /clientes [get]
func (c *ClienteController) ListarClientes(w http.ResponseWriter, r *http.Request) {
	clientes, err := c.service.BuscarTodosClientes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, clientes)
}

// BuscarClientePorID retorna um cliente específico
// @Summary Busca um cliente por ID
// @Description Retorna os detalhes de um cliente específico
// @Tags clientes
// @Produce json
// @Param id path string true "ID do Cliente"
// @Success 200 {object} model.Cliente
// @Failure 404 {string} string "Cliente não encontrado"
// @Router /clientes/{id} [get]
func (c *ClienteController) BuscarClientePorID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	cliente, err := c.service.BuscarClientePorID(r.Context(), id)
	if err != nil {
		if err == service.ErrNotFound {
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, cliente)
}

// CriarCliente adiciona um novo cliente
// @Summary Adiciona um novo cliente
// @Description Cria um novo cliente no sistema
// @Tags clientes
// @Accept json
// @Produce json
// @Param cliente body model.Cliente true "Dados do Cliente"
// @Success 201
// @Failure 400 {string} string "Dados inválidos"
// @Failure 409 {string} string "Cliente já existe"
// @Router /clientes [post]
func (c *ClienteController) CriarCliente(w http.ResponseWriter, r *http.Request) {
	var cliente model.Cliente
	if err := json.NewDecoder(r.Body).Decode(&cliente); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if err := c.service.AdicionarCliente(r.Context(), cliente); err != nil {
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

// AtualizarCliente atualiza um cliente existente
// @Summary Atualiza um cliente
// @Description Atualiza os dados de um cliente existente
// @Tags clientes
// @Accept json
// @Produce json
// @Param id path string true "ID do Cliente"
// @Param cliente body model.Cliente true "Dados atualizados do Cliente"
// @Success 200
// @Failure 400 {string} string "Dados inválidos"
// @Failure 404 {string} string "Cliente não encontrado"
// @Router /clientes/{id} [put]
func (c *ClienteController) AtualizarCliente(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var cliente model.Cliente
	if err := json.NewDecoder(r.Body).Decode(&cliente); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if err := c.service.AtualizarCliente(r.Context(), id, cliente); err != nil {
		switch err {
		case service.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case service.ErrNotFound:
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeletarCliente remove um cliente
// @Summary Remove um cliente
// @Description Remove um cliente do sistema
// @Tags clientes
// @Produce json
// @Param id path string true "ID do Cliente"
// @Success 204
// @Failure 404 {string} string "Cliente não encontrado"
// @Failure 400 {string} string "Cliente possui pedidos associados"
// @Router /clientes/{id} [delete]
func (c *ClienteController) DeletarCliente(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := c.service.DeletarCliente(r.Context(), id); err != nil {
		switch err {
		case service.ErrNotFound:
			http.Error(w, "Cliente não encontrado", http.StatusNotFound)
		case service.ErrDependency:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function para enviar respostas JSON
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

// CountClientes retorna o número total de clientes
// @Summary Retorna a contagem total de clientes
// @Description Retorna o número total de clientes cadastrados no sistema
// @Tags clientes
// @Produce json
// @Success 200 {integer} integer "Número total de clientes"
// @Router /clientes/count [get]
func (c *ClienteController) CountClientes(w http.ResponseWriter, r *http.Request) {
	count, err := c.service.CountClientes(r.Context())
	if err != nil {
		// Log do erro para debug
		log.Printf("Erro ao contar clientes: %v", err)

		// Mensagem de erro mais amigável
		http.Error(w, "Não foi possível obter a contagem de clientes", http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]int{"total": count})
}

// BuscarClientesPorNome busca clientes por nome
// @Summary Busca clientes por nome
// @Description Retorna os clientes cujos nomes correspondem ao parâmetro de busca
// @Tags clientes
// @Produce json
// @Param nome query string true "Nome ou parte do nome para busca"
// @Success 200 {array} model.Cliente
// @Failure 400 {string} string "Nome não pode ser vazio"
// @Router /clientes/search [get]
func (c *ClienteController) BuscarClientesPorNome(w http.ResponseWriter, r *http.Request) {
	nome := r.URL.Query().Get("nome")
	if nome == "" {
		http.Error(w, "Parâmetro 'nome' é obrigatório", http.StatusBadRequest)
		return
	}

	clientes, err := c.service.BuscarClientesPorNome(r.Context(), nome)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respondWithJSON(w, http.StatusOK, clientes)
}
