package repository

import (
	"context"
	"database/sql"
	"errors"

	"registro-despesas/internal/models"
)

var ErrNotFound = errors.New("registro não encontrado")

const usuarioCols = `id, nome, email, senha, nivel, created_at`

func scanUsuario(row *sql.Row) (*models.Usuario, error) {
	var u models.Usuario
	err := row.Scan(&u.ID, &u.Nome, &u.Email, &u.Senha, &u.Nivel, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*models.Usuario, error) {
	return scanUsuario(r.db.QueryRowContext(ctx,
		`SELECT `+usuarioCols+` FROM comum.usuario WHERE LOWER(email) = LOWER($1)`, email))
}

func (r *AuthRepository) FindByID(ctx context.Context, id int) (*models.Usuario, error) {
	return scanUsuario(r.db.QueryRowContext(ctx,
		`SELECT `+usuarioCols+` FROM comum.usuario WHERE id = $1`, id))
}

func (r *AuthRepository) UpdateSenha(ctx context.Context, id int, hash string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE comum.usuario SET senha = $1 WHERE id = $2`, hash, id)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrNotFound
	}
	return nil
}
