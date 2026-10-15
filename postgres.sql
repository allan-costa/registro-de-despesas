-- Cria schema comum
CREATE SCHEMA comum AUTHORIZATION postgres;

-- Criar tipos ENUM
DO $$ BEGIN 
    	CREATE TYPE nivel_usuario AS ENUM ('ADMIN', 'READ');
	EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    	CREATE TYPE tipo_despesa AS ENUM ('PADRAO', 'EVENTUAL', 'RECORRENTE', 'PARCELAMENTO');
	EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- 1. Tabela de Usuários (Autenticação)
CREATE TABLE IF NOT EXISTS comum.usuario (
    id          SERIAL PRIMARY KEY,
    nome        VARCHAR NOT NULL,
    email       VARCHAR NOT NULL UNIQUE,
    senha       VARCHAR NOT NULL,
    nivel       nivel_usuario NOT NULL default 'READ',
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Tabela de Categorias
CREATE TABLE IF NOT EXISTS comum.categoria (
id 		SERIAL PRIMARY KEY,
parent_id   	INT NULL DEFAULT NULL,
nome       	VARCHAR NOT NULL,
created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
CONSTRAINT fk_categoria_parent FOREIGN KEY (parent_id) REFERENCES comum.categoria(id) ON DELETE CASCADE
);

-- 3. Tabela de Despesas
CREATE TABLE IF NOT EXISTS comum.despesa (
    id              	SERIAL PRIMARY KEY,
    id_categoria    	INT4 NOT null references comum.categoria(id),
    tipo		tipo_despesa not null default 'PADRAO',
    descricao       	VARCHAR NOT NULL,
    valor	    		DECIMAL(12,2) NOT NULL,
    data		TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    num_parcela 	VARCHAR(20) NULL DEFAULT NULL,
    created_at      	TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Criar extensão do PostgreSQL caso não exista e encripta senha de inicio do usuário Administrador
CREATE EXTENSION IF NOT EXISTS pgcrypto;
SELECT crypt('123456', gen_salt('bf'));

-- Seed: Criar usuario administrador
INSERT INTO comum.usuario (nome, email, senha, nivel) 
VALUES ('Administrador', 'admin@admin.com', '$2a$06$If1Klyr5Lq0lM4dCodPbr.SzV8jfO1t2ycMI51g4qGLolMZuamK1K', 'ADMIN') ON CONFLICT (email) DO NOTHING;
