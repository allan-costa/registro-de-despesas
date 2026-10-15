package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"registro-despesas/internal/models"
)

// ErrVinculado indica violação de FK ao excluir (existem despesas vinculadas).
var ErrVinculado = errors.New("registro possui vínculos")

const (
	categoriaCols = `id, parent_id, nome, created_at`
	pgForeignKey  = "23503"
)

type rowScanner interface {
	Scan(dest ...any) error
}

// scanCategoria devolve Filhos vazio (e não nil) para serializar como [].
func scanCategoria(row rowScanner) (*models.Categoria, error) {
	c := models.Categoria{Filhos: []*models.Categoria{}}
	err := row.Scan(&c.ID, &c.ParentID, &c.Nome, &c.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type CategoriaRepository struct {
	db *sql.DB
}

func NewCategoriaRepository(db *sql.DB) *CategoriaRepository {
	return &CategoriaRepository{db: db}
}

func (r *CategoriaRepository) ListAll(ctx context.Context) ([]*models.Categoria, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+categoriaCols+` FROM comum.categoria ORDER BY nome, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.Categoria
	for rows.Next() {
		c, err := scanCategoria(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CategoriaRepository) FindByID(ctx context.Context, id int) (*models.Categoria, error) {
	return scanCategoria(r.db.QueryRowContext(ctx,
		`SELECT `+categoriaCols+` FROM comum.categoria WHERE id = $1`, id))
}

func (r *CategoriaRepository) Create(ctx context.Context, nome string, parentID *int) (*models.Categoria, error) {
	return scanCategoria(r.db.QueryRowContext(ctx,
		`INSERT INTO comum.categoria (parent_id, nome) VALUES ($1, $2) RETURNING `+categoriaCols,
		parentID, nome))
}

func (r *CategoriaRepository) Update(ctx context.Context, id int, nome string, parentID *int) (*models.Categoria, error) {
	return scanCategoria(r.db.QueryRowContext(ctx,
		`UPDATE comum.categoria SET parent_id = $1, nome = $2 WHERE id = $3 RETURNING `+categoriaCols,
		parentID, nome, id))
}

// IsAncestorOrSelf informa se ancestorID é nodeID ou algum ancestral dele (detecta ciclos ao mudar o pai).
func (r *CategoriaRepository) IsAncestorOrSelf(ctx context.Context, ancestorID, nodeID int) (bool, error) {
	var ok bool
	err := r.db.QueryRowContext(ctx, `
		WITH RECURSIVE anc AS (
			SELECT id, parent_id FROM comum.categoria WHERE id = $1
			UNION
			SELECT c.id, c.parent_id FROM comum.categoria c JOIN anc a ON c.id = a.parent_id
		)
		SELECT EXISTS (SELECT 1 FROM anc WHERE id = $2)`, nodeID, ancestorID).Scan(&ok)
	return ok, err
}

// Delete remove a categoria e, por cascade, as subcategorias; falha com ErrVinculado se
// ela ou qualquer descendente tiver despesas (a FK de despesa impede e o comando é revertido).
func (r *CategoriaRepository) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM comum.categoria WHERE id = $1`, id)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == pgForeignKey {
		return ErrVinculado
	}
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
