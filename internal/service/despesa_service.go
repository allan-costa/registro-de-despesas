package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"registro-despesas/internal/models"
	"registro-despesas/internal/repository"
)

const (
	minParcelas  = 2
	maxParcelas  = 120
	maxDescricao = 200
	maxValor     = 9999999999.99 // limite de NUMERIC(12,2)
)

// Formatos aceitos para data; o primeiro é o oficial (YYYY-MM-DD HH-mm-ss).
var dataLayouts = []string{"2006-01-02 15-04-05", "2006-01-02 15:04:05", "2006-01-02"}

var (
	ErrDescricaoInvalida  = errors.New("a descrição é obrigatória e deve ter até 200 caracteres")
	ErrValorInvalido      = errors.New("o valor deve ser maior que zero e menor que 10 bilhões")
	ErrDataInvalida       = errors.New("data inválida; use o formato YYYY-MM-DD HH-mm-ss")
	ErrTipoInvalido       = errors.New("tipo de despesa inválido")
	ErrParcelasInvalidas  = errors.New("informe o total de parcelas (entre 2 e 120)")
	ErrDespesaInexistente = errors.New("despesa não encontrada")
)

type DespesaService struct {
	repo *repository.DespesaRepository
}

func NewDespesaService(repo *repository.DespesaRepository) *DespesaService {
	return &DespesaService{repo: repo}
}

// Criar gera 1 registro (PADRAO/EVENTUAL) ou N mensais consecutivos (RECORRENTE/PARCELAMENTO),
// todos com o mesmo valor informado, gravados em uma única transação.
func (s *DespesaService) Criar(ctx context.Context, req models.DespesaRequest) ([]models.Despesa, error) {
	descricao, valor, err := validarDescricaoValor(req.Descricao, req.Valor)
	if err != nil {
		return nil, err
	}
	if req.CategoriaID <= 0 {
		return nil, ErrCategoriaInexistente
	}

	data, err := parseData(req.Data)
	if err != nil {
		return nil, err
	}

	base := models.Despesa{CategoriaID: req.CategoriaID, Tipo: req.Tipo, Descricao: descricao, Valor: valor, Data: data}

	var items []models.Despesa
	switch req.Tipo {
	case models.TipoPadrao, models.TipoEventual:
		items = []models.Despesa{base}
	case models.TipoRecorrente, models.TipoParcelamento:
		n := req.TotalParcelas
		if n < minParcelas || n > maxParcelas {
			return nil, ErrParcelasInvalidas
		}
		items = make([]models.Despesa, 0, n)
		for y := 1; y <= n; y++ {
			d := base
			d.Data = addMonths(data, y-1)
			num := fmt.Sprintf("%d/%d", y, n)
			d.NumParcela = &num
			if req.Tipo == models.TipoRecorrente {
				d.Descricao = fmt.Sprintf("%s - Recorrência %s", descricao, num)
			} else {
				d.Descricao = fmt.Sprintf("Parcela %s %s", descricao, num)
			}
			items = append(items, d)
		}
	default:
		return nil, ErrTipoInvalido
	}

	out, err := s.repo.CreateBatch(ctx, items)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrCategoriaInexistente
	}
	return out, err
}

// Editar altera somente descrição e valor da despesa informada.
func (s *DespesaService) Editar(ctx context.Context, id int, req models.DespesaRequest) (*models.Despesa, error) {
	descricao, valor, err := validarDescricaoValor(req.Descricao, req.Valor)
	if err != nil {
		return nil, err
	}

	d, err := s.repo.UpdateDescricaoValor(ctx, id, descricao, valor)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrDespesaInexistente
	}
	return d, err
}

// Excluir remove a despesa ou, com grupo, todo o lote; devolve quantos registros foram removidos.
func (s *DespesaService) Excluir(ctx context.Context, id int, grupo bool) (int, error) {
	n, err := s.repo.Delete(ctx, id, grupo)
	if errors.Is(err, repository.ErrNotFound) {
		return 0, ErrDespesaInexistente
	}
	return n, err
}

func validarDescricaoValor(descricao string, valor float64) (string, float64, error) {
	descricao = strings.TrimSpace(descricao)
	if n := utf8.RuneCountInString(descricao); n == 0 || n > maxDescricao {
		return "", 0, ErrDescricaoInvalida
	}
	valor = math.Round(valor*100) / 100
	if math.IsNaN(valor) || valor <= 0 || valor > maxValor {
		return "", 0, ErrValorInvalido
	}
	return descricao, valor, nil
}

func parseData(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range dataLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrDataInvalida
}

// addMonths soma n meses à data original, limitando o dia ao último do mês (31/01 + 1 mês = 28/02).
func addMonths(t time.Time, n int) time.Time {
	first := time.Date(t.Year(), t.Month()+time.Month(n), 1, 0, 0, 0, 0, t.Location())
	lastDay := time.Date(first.Year(), first.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
	day := min(t.Day(), lastDay)
	return time.Date(first.Year(), first.Month(), day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}
