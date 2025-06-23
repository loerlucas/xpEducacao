package main

import (
	"api/config"
	"api/controller"
	_ "api/docs" // Import para documentação Swagger
	"api/repository"
	"api/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // Driver PostgreSQL
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title API de E-commerce
// @version 1.0
// @description API para gerenciamento de e-commerce
// @termsOfService http://swagger.io/terms/

// @contact.name Suporte API
// @contact.url http://www.suporte.com
// @contact.email suporte@api.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {

	// Carrega variáveis do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar arquivo .env")
	}

	// Desligamento gracioso
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Inicializar conexão com o banco de dados
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// Garantir que a conexão seja fechada ao final
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Erro ao fechar conexão com o banco: %v", err)
		}
	}()

	// Verificar conexão com o banco
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Erro ao verificar conexão com o banco: %v", err)
	}

	// Inicializar repositórios
	clienteRepo := repository.NewClienteRepository(db)
	produtoRepo := repository.NewProdutoRepository(db)
	pedidoRepo := repository.NewPedidoRepository(db)

	// Inicializar services
	clienteService := service.NewClienteService(clienteRepo)
	produtoService := service.NewProdutoService(produtoRepo)
	pedidoService := service.NewPedidoService(pedidoRepo, clienteRepo, produtoRepo, produtoService)

	// Inicializar controllers
	clienteController := controller.NewClienteController(clienteService)
	produtoController := controller.NewProdutoController(produtoService)
	pedidoController := controller.NewPedidoController(pedidoService)

	// Configurar roteador
	r := mux.NewRouter()

	// Middlewares
	r.Use(loggingMiddleware)
	r.Use(contentTypeMiddleware)

	// Rotas de Clientes
	clienteRouter := r.PathPrefix("/clientes").Subrouter()
	r.HandleFunc("/clientes", clienteController.ListarClientes).Methods("GET")
	clienteRouter.HandleFunc("/count", clienteController.CountClientes).Methods("GET")
	clienteRouter.HandleFunc("/search", clienteController.BuscarClientesPorNome).Methods("GET")
	clienteRouter.HandleFunc("", clienteController.CriarCliente).Methods("POST")
	clienteRouter.HandleFunc("/{id}", clienteController.BuscarClientePorID).Methods("GET")
	clienteRouter.HandleFunc("/{id}", clienteController.AtualizarCliente).Methods("PUT")
	clienteRouter.HandleFunc("/{id}", clienteController.DeletarCliente).Methods("DELETE")

	// Rotas de Produtos
	produtoRouter := r.PathPrefix("/produtos").Subrouter()
	produtoRouter.HandleFunc("", produtoController.ListarProdutos).Methods("GET")
	produtoRouter.HandleFunc("/count", produtoController.CountProdutos).Methods("GET")
	produtoRouter.HandleFunc("/search", produtoController.BuscarProdutosPorNome).Methods("GET")
	produtoRouter.HandleFunc("", produtoController.CriarProduto).Methods("POST")
	produtoRouter.HandleFunc("/{id}", produtoController.BuscarProdutoPorID).Methods("GET")
	produtoRouter.HandleFunc("/{id}", produtoController.AtualizarProduto).Methods("PUT")
	produtoRouter.HandleFunc("/{id}", produtoController.DeletarProduto).Methods("DELETE")

	// Rotas de Pedidos
	pedidoRouter := r.PathPrefix("/pedidos").Subrouter()
	pedidoRouter.HandleFunc("", pedidoController.ListarPedidos).Methods("GET")
	pedidoRouter.HandleFunc("/count", pedidoController.CountPedidos).Methods("GET")
	pedidoRouter.HandleFunc("/search", pedidoController.BuscarPedidosPorNomeCliente).Methods("GET")
	pedidoRouter.HandleFunc("", pedidoController.CriarPedido).Methods("POST")
	pedidoRouter.HandleFunc("/{id}", pedidoController.BuscarPedidoPorID).Methods("GET")
	pedidoRouter.HandleFunc("/{id}/status", pedidoController.AtualizarStatusPedido).Methods("PUT")
	pedidoRouter.HandleFunc("/{id}/cancelar", pedidoController.CancelarPedido).Methods("POST")
	pedidoRouter.HandleFunc("/{id}", pedidoController.DeletarPedido).Methods("DELETE")

	// Documentação Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Configurar servidor HTTP
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Iniciar servidor em goroutine
	go func() {
		log.Printf("Servidor iniciado em http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	// Aguardar sinal de desligamento
	<-ctx.Done()

	// Configurar timeout para desligamento gracioso
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Desligar servidor
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Erro durante desligamento do servidor: %v", err)
	} else {
		log.Println("Servidor desligado graciosamente")
	}
}

// loggingMiddleware registra informações sobre as requisições
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf(
			"Iniciando requisição %s %s",
			r.Method,
			r.RequestURI,
		)

		next.ServeHTTP(w, r)

		log.Printf(
			"Requisição concluída %s %s em %v",
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

// contentTypeMiddleware define o Content-Type como application/json
func contentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
