package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"registro-despesas/internal/models"
)

const despesaCols = `id, id_categoria, tipo, descricao, valor, data, num_parcela, created_at`

func scanDespesa(row rowScanner) (*models.Despesa, error) {
	var d models.Despesa
	err := row.Scan(&d.ID, &d.CategoriaID, &d.Tipo, &d.Descricao, &d.Valor, &d.Data, &d.NumParcela, &d.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

type DespesaRepository struct {
	db *sql.DB
}

func NewDespesaRepository(db *sql.DB) *DespesaRepository {
	return &DespesaRepository{db: db}
}

// CreateBatch insere todas as despesas em uma única transação. Por isso elas compartilham o mesmo
// created_at (CURRENT_TIMESTAMP é fixo por transação), que identifica o lote na exclusão em grupo.
// Retorna ErrNotFound se a categoria não existir (violação da FK).
func (r *DespesaRepository) CreateBatch(ctx context.Context, items []models.Despesa) ([]models.Despesa, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO comum.despesa (id_categoria, tipo, descricao, valor, data, num_parcela)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+despesaCols)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	out := make([]models.Despesa, 0, len(items))
	for _, it := range items {
		d, err := scanDespesa(stmt.QueryRowContext(ctx,
			it.CategoriaID, it.Tipo, it.Descricao, it.Valor, it.Data, it.NumParcela))
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == pgForeignKey {
			return nil, ErrNotFound
		}
		if err != nil {
			return nil, err
		}
		out = append(out, *d)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateDescricaoValor altera somente descrição e valor.
func (r *DespesaRepository) UpdateDescricaoValor(ctx context.Context, id int, descricao string, valor float64) (*models.Despesa, error) {
	return scanDespesa(r.db.QueryRowContext(ctx,
		`UPDATE comum.despesa SET descricao = $1, valor = $2 WHERE id = $3 RETURNING `+despesaCols,
		descricao, valor, id))
}

// Delete remove a despesa; com group, remove todo o lote (mesmo tipo, categoria e created_at) de
// uma despesa parcelada/recorrente. Despesas sem parcela são sempre removidas individualmente.
// Retorna a quantidade de registros removidos.
func (r *DespesaRepository) Delete(ctx context.Context, id int, group bool) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var n int64
	if group {
		res, err := tx.ExecContext(ctx, `
			DELETE FROM comum.despesa d USING comum.despesa o
			WHERE o.id = $1 AND o.num_parcela IS NOT NULL
			  AND d.tipo = o.tipo AND d.id_categoria = o.id_categoria AND d.created_at = o.created_at`, id)
		if err != nil {
			return 0, err
		}
		if n, err = res.RowsAffected(); err != nil {
			return 0, err
		}
	}

	if n == 0 {
		res, err := tx.ExecContext(ctx, `DELETE FROM comum.despesa WHERE id = $1`, id)
		if err != nil {
			return 0, err
		}
		if n, err = res.RowsAffected(); err != nil {
			return 0, err
		}
	}

	if n == 0 {
		return 0, ErrNotFound
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return int(n), nil
}
