CREATE TABLE IF NOT EXISTS clientes (
    id VARCHAR(36) PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS produtos (
    id VARCHAR(36) PRIMARY KEY,
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    preco DECIMAL(10,2) NOT NULL,
    estoque INTEGER NOT NULL,
    categoria VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS pedidos (
    id VARCHAR(36) PRIMARY KEY,
    cliente_id VARCHAR(36) NOT NULL REFERENCES clientes(id),
    data TIMESTAMP NOT NULL,
    total DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS itens_pedido (
    pedido_id VARCHAR(36) NOT NULL REFERENCES pedidos(id),
    produto_id VARCHAR(36) NOT NULL REFERENCES produtos(id),
    quantidade INTEGER NOT NULL,
    preco_unit DECIMAL(10,2) NOT NULL,
    subtotal DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (pedido_id, produto_id)
);