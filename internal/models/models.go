package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	NivelAdmin = "ADMIN"
	NivelRead  = "READ"
)

const (
	TipoPadrao       = "PADRAO"
	TipoEventual     = "EVENTUAL"
	TipoRecorrente   = "RECORRENTE"
	TipoParcelamento = "PARCELAMENTO"
)

// JWTClaims: payload id, email, nivel, exp (exp vem de RegisteredClaims).
type JWTClaims struct {
	UserID int    `json:"id"`
	Email  string `json:"email"`
	Nivel  string `json:"nivel"`
	jwt.RegisteredClaims
}

var _ jwt.Claims = (*JWTClaims)(nil)

// --- Entidades (comum.*) ---

type Usuario struct {
	ID        int       `json:"id"`
	Nome      string    `json:"nome"`
	Email     string    `json:"email"`
	Senha     string    `json:"-"`
	Nivel     string    `json:"nivel"`
	CreatedAt time.Time `json:"created_at"`
}

type Categoria struct {
	ID        int          `json:"id"`
	ParentID  *int         `json:"parent_id"`
	Nome      string       `json:"nome"`
	CreatedAt time.Time    `json:"created_at"`
	Filhos    []*Categoria `json:"filhos"`
}

type Despesa struct {
	ID            int       `json:"id"`
	CategoriaID   int       `json:"categoria_id"`
	CategoriaNome string    `json:"categoria_nome,omitempty"`
	Tipo          string    `json:"tipo"`
	Descricao     string    `json:"descricao"`
	Valor         float64   `json:"valor"`
	Data          time.Time `json:"data"`
	NumParcela    *string   `json:"num_parcela"`
	CreatedAt     time.Time `json:"created_at"`
}

// --- Auth ---

type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Senha string `json:"senha" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Nivel string `json:"nivel"`
}

type AlterarSenhaRequest struct {
	SenhaAtual string `json:"senha_atual" binding:"required"`
	SenhaNova  string `json:"senha_nova" binding:"required,min=6"`
}

// --- Categorias ---

// CategoriaRequest: com ID presente edita, sem ID cria.
type CategoriaRequest struct {
	ID       *int   `json:"id"`
	Nome     string `json:"nome" binding:"required"`
	ParentID *int   `json:"parent_id"`
}

// --- Usuários ---

// UsuarioRequest: com ID presente edita (senha opcional), sem ID cria.
type UsuarioRequest struct {
	ID             *int   `json:"id"`
	Nome           string `json:"nome" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Senha          string `json:"senha"`
	ConfirmarSenha string `json:"confirmar_senha"`
	Nivel          string `json:"nivel" binding:"required,oneof=ADMIN READ"`
}

// --- Despesas ---

type DespesaRequest struct {
	CategoriaID   int     `json:"categoria_id" binding:"required"`
	Descricao     string  `json:"descricao" binding:"required"`
	Valor         float64 `json:"valor" binding:"required,gt=0"`
	Data          string  `json:"data" binding:"required"` // YYYY-MM-DD HH-mm-ss
	Tipo          string  `json:"tipo" binding:"required,oneof=PADRAO EVENTUAL RECORRENTE PARCELAMENTO"`
	TotalParcelas int     `json:"total_parcelas"`
}

type DespesaListResponse struct {
	Data           []Despesa `json:"data"`
	Pagina         int       `json:"pagina"`
	TotalPaginas   int       `json:"total_paginas"`
	TotalRegistros int       `json:"total_registros"`
	TotalPagina    float64   `json:"total_pagina"`
	TotalDespesas  float64   `json:"total_despesas"`
}

type GraficoResponse struct {
	Labels  []string  `json:"labels"`
	Valores []float64 `json:"valores"`
}
