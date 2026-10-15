package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"registro-despesas/internal/models"
)

// ErrDuplicado indica violação de unicidade (e-mail já cadastrado).
var ErrDuplicado = errors.New("registro duplicado")

const pgUniqueViolation = "23505"

type UsuarioRepository struct {
	db *sql.DB
}

func NewUsuarioRepository(db *sql.DB) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

// ListAll não seleciona a senha.
func (r *UsuarioRepository) ListAll(ctx context.Context) ([]*models.Usuario, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, nome, email, nivel, created_at FROM comum.usuario ORDER BY nome, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*models.Usuario{}
	for rows.Next() {
		var u models.Usuario
		if err := rows.Scan(&u.ID, &u.Nome, &u.Email, &u.Nivel, &u.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &u)
	}
	return out, rows.Err()
}

// EmailExists compara sem diferenciar maiúsculas; excludeID ignora o próprio registro na edição (0 na criação).
func (r *UsuarioRepository) EmailExists(ctx context.Context, email string, excludeID int) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM comum.usuario WHERE LOWER(email) = LOWER($1) AND id <> $2)`,
		email, excludeID).Scan(&ok)
	return ok, err
}

func (r *UsuarioRepository) Create(ctx context.Context, nome, email, hash, nivel string) (*models.Usuario, error) {
	u, err := scanUsuario(r.db.QueryRowContext(ctx,
		`INSERT INTO comum.usuario (nome, email, senha, nivel) VALUES ($1, $2, $3, $4) RETURNING `+usuarioCols,
		nome, email, hash, nivel))
	return u, mapUnique(err)
}

// Update mantém a senha atual quando hash é nil.
func (r *UsuarioRepository) Update(ctx context.Context, id int, nome, email, nivel string, hash *string) (*models.Usuario, error) {
	u, err := scanUsuario(r.db.QueryRowContext(ctx,
		`UPDATE comum.usuario SET nome = $1, email = $2, nivel = $3, senha = COALESCE($4, senha)
		 WHERE id = $5 RETURNING `+usuarioCols,
		nome, email, nivel, hash, id))
	return u, mapUnique(err)
}

func (r *UsuarioRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM comum.usuario WHERE id = $1`, id)
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

func mapUnique(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgUniqueViolation {
		return ErrDuplicado
	}
	return err
}
