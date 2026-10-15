package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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

// DespesaFiltro: paginação (Limite/Deslocamento) já vem normalizada pelo service.
type DespesaFiltro struct {
	DataInicial  *time.Time
	DataFinal    *time.Time // inclusiva: o dia inteiro entra
	CategoriaID  int        // 0 = todas; inclui as subcategorias
	Tipo         string     // "" = todos
	Limite       int
	Deslocamento int
}

// where monta o WHERE sobre a tabela com alias d.
func (f DespesaFiltro) where() (string, []any) {
	var conds []string
	var args []any
	add := func(cond string, valor any) {
		args = append(args, valor)
		conds = append(conds, strings.Replace(cond, "?", "$"+strconv.Itoa(len(args)), 1))
	}

	if f.DataInicial != nil {
		add("d.data >= ?", *f.DataInicial)
	}
	if f.DataFinal != nil {
		add("d.data < ?", f.DataFinal.AddDate(0, 0, 1))
	}
	if f.CategoriaID > 0 {
		add(`d.id_categoria IN (
			WITH RECURSIVE sub AS (
				SELECT id FROM comum.categoria WHERE id = ?
				UNION
				SELECT c.id FROM comum.categoria c JOIN sub s ON c.parent_id = s.id
			) SELECT id FROM sub)`, f.CategoriaID)
	}
	if f.Tipo != "" {
		add("d.tipo = ?::tipo_despesa", f.Tipo)
	}

	if len(conds) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(conds, " AND "), args
}

// Totais devolve a quantidade e a soma dos valores de todas as despesas do filtro (sem paginação).
func (r *DespesaRepository) Totais(ctx context.Context, f DespesaFiltro) (int, float64, error) {
	where, args := f.where()
	var n int
	var soma float64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(d.valor), 0) FROM comum.despesa d`+where, args...).Scan(&n, &soma)
	return n, soma, err
}

// Listar devolve a página do filtro em ordem cronológica, com o nome da categoria.
func (r *DespesaRepository) Listar(ctx context.Context, f DespesaFiltro) ([]models.Despesa, error) {
	where, args := f.where()
	args = append(args, f.Limite, f.Deslocamento)

	rows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.id_categoria, c.nome, d.tipo, d.descricao, d.valor, d.data, d.num_parcela, d.created_at
		FROM comum.despesa d JOIN comum.categoria c ON c.id = d.id_categoria`+where+
		fmt.Sprintf(` ORDER BY d.data, d.id LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Despesa{}
	for rows.Next() {
		var d models.Despesa
		if err := rows.Scan(&d.ID, &d.CategoriaID, &d.CategoriaNome, &d.Tipo, &d.Descricao, &d.Valor, &d.Data, &d.NumParcela, &d.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// TotaisPorCategoria soma o valor por categoria (a do lançamento), do maior para o menor.
func (r *DespesaRepository) TotaisPorCategoria(ctx context.Context, f DespesaFiltro) ([]string, []float64, error) {
	where, args := f.where()
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.nome, SUM(d.valor)
		FROM comum.despesa d JOIN comum.categoria c ON c.id = d.id_categoria`+where+`
		GROUP BY c.id, c.nome ORDER BY SUM(d.valor) DESC, c.nome`, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	labels, valores := []string{}, []float64{}
	for rows.Next() {
		var nome string
		var valor float64
		if err := rows.Scan(&nome, &valor); err != nil {
			return nil, nil, err
		}
		labels = append(labels, nome)
		valores = append(valores, valor)
	}
	return labels, valores, rows.Err()
}
