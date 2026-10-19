package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"registro-despesas/internal/models"
	"registro-despesas/internal/repository"
)

type despesaRepositoryFake struct {
	despesas  []models.Despesa
	proximoID int
}

func novoDespesaRepositoryFake() *despesaRepositoryFake {
	return &despesaRepositoryFake{proximoID: 1}
}

func (f *despesaRepositoryFake) CreateBatch(_ context.Context, items []models.Despesa) ([]models.Despesa, error) {
	criados := make([]models.Despesa, 0, len(items))
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, item := range items {
		item.ID = f.proximoID
		item.CreatedAt = createdAt
		f.proximoID++
		f.despesas = append(f.despesas, item)
		criados = append(criados, item)
	}
	return criados, nil
}

func (f *despesaRepositoryFake) UpdateDescricaoValor(_ context.Context, id int, descricao string, valor float64) (*models.Despesa, error) {
	for i := range f.despesas {
		if f.despesas[i].ID == id {
			f.despesas[i].Descricao = descricao
			f.despesas[i].Valor = valor
			return &f.despesas[i], nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *despesaRepositoryFake) Delete(_ context.Context, id int, group bool) (int, error) {
	for i := range f.despesas {
		if f.despesas[i].ID == id {
			f.despesas = append(f.despesas[:i], f.despesas[i+1:]...)
			return 1, nil
		}
	}
	return 0, repository.ErrNotFound
}

func (f *despesaRepositoryFake) Totais(context.Context, repository.DespesaFiltro) (int, float64, error) {
	var soma float64
	for _, despesa := range f.despesas {
		soma += despesa.Valor
	}
	return len(f.despesas), soma, nil
}

func (f *despesaRepositoryFake) Listar(_ context.Context, filtro repository.DespesaFiltro) ([]models.Despesa, error) {
	inicio := min(filtro.Deslocamento, len(f.despesas))
	fim := min(inicio+filtro.Limite, len(f.despesas))
	return append([]models.Despesa(nil), f.despesas[inicio:fim]...), nil
}

func (f *despesaRepositoryFake) TotaisPorCategoria(context.Context, repository.DespesaFiltro) ([]string, []float64, error) {
	return nil, nil, nil
}

func TestDespesaServiceCRUD(t *testing.T) {
	ctx := context.Background()
	repo := novoDespesaRepositoryFake()
	svc := NewDespesaService(repo)

	criadas, err := svc.Criar(ctx, models.DespesaRequest{
		CategoriaID: 1, Tipo: models.TipoPadrao, Descricao: "  Almoço  ", Valor: 25.50, Data: "2026-10-01",
	})
	if err != nil {
		t.Fatalf("Criar retornou erro: %v", err)
	}
	if len(criadas) != 1 || criadas[0].Descricao != "Almoço" || criadas[0].Valor != 25.50 {
		t.Fatalf("despesas criadas = %+v; esperado um lançamento normalizado", criadas)
	}

	lista, err := svc.Listar(ctx, models.DespesaFiltroRequest{Pagina: 1, PorPagina: 10})
	if err != nil {
		t.Fatalf("Listar retornou erro: %v", err)
	}
	if len(lista.Data) != 1 || lista.TotalRegistros != 1 || lista.TotalDespesas != 25.50 {
		t.Fatalf("resposta da consulta = %+v; esperado um registro e total 25.50", lista)
	}

	atualizada, err := svc.Editar(ctx, criadas[0].ID, models.DespesaRequest{Descricao: "Almoço de trabalho", Valor: 30})
	if err != nil {
		t.Fatalf("Editar retornou erro: %v", err)
	}
	if atualizada.Descricao != "Almoço de trabalho" || atualizada.Valor != 30 {
		t.Errorf("despesa após edição = %+v; dados inesperados", atualizada)
	}

	excluidas, err := svc.Excluir(ctx, criadas[0].ID, false)
	if err != nil || excluidas != 1 {
		t.Fatalf("Excluir = %d, erro = %v; esperado 1 registro removido", excluidas, err)
	}
	lista, err = svc.Listar(ctx, models.DespesaFiltroRequest{})
	if err != nil || lista.TotalRegistros != 0 {
		t.Errorf("consulta após exclusão = %+v, erro = %v; esperado nenhum registro", lista, err)
	}
}

func TestDespesaServiceCriarParcelamento(t *testing.T) {
	repo := novoDespesaRepositoryFake()
	despesas, err := NewDespesaService(repo).Criar(context.Background(), models.DespesaRequest{
		CategoriaID: 1, Tipo: models.TipoParcelamento, Descricao: "Compra", Valor: 100,
		Data: "2026-01-31", TotalParcelas: 2,
	})
	if err != nil {
		t.Fatalf("Criar parcelamento retornou erro: %v", err)
	}
	if len(despesas) != 2 || despesas[0].Descricao != "Parcela Compra 1/2" || despesas[1].Descricao != "Parcela Compra 2/2" {
		t.Fatalf("parcelas = %+v; esperado duas parcelas numeradas", despesas)
	}
	if despesas[1].Data.Month() != time.February || despesas[1].Data.Day() != 28 {
		t.Errorf("data da segunda parcela = %s; esperado 28/02/2026", despesas[1].Data.Format("2006-01-02"))
	}
}

func TestDespesaServiceRejeitaEntradasInvalidas(t *testing.T) {
	tests := []struct {
		nome string
		req  models.DespesaRequest
		want error
	}{
		{nome: "descrição vazia", req: models.DespesaRequest{CategoriaID: 1, Tipo: models.TipoPadrao, Valor: 10, Data: "2026-10-01"}, want: ErrDescricaoInvalida},
		{nome: "valor zero", req: models.DespesaRequest{CategoriaID: 1, Tipo: models.TipoPadrao, Descricao: "Compra", Valor: 0, Data: "2026-10-01"}, want: ErrValorInvalido},
		{nome: "categoria inválida", req: models.DespesaRequest{Tipo: models.TipoPadrao, Descricao: "Compra", Valor: 10, Data: "2026-10-01"}, want: ErrCategoriaInexistente},
		{nome: "data inválida", req: models.DespesaRequest{CategoriaID: 1, Tipo: models.TipoPadrao, Descricao: "Compra", Valor: 10, Data: "31/10/2026"}, want: ErrDataInvalida},
		{nome: "tipo inválido", req: models.DespesaRequest{CategoriaID: 1, Tipo: "OUTRO", Descricao: "Compra", Valor: 10, Data: "2026-10-01"}, want: ErrTipoInvalido},
		{nome: "parcelas inválidas", req: models.DespesaRequest{CategoriaID: 1, Tipo: models.TipoParcelamento, Descricao: "Compra", Valor: 10, Data: "2026-10-01", TotalParcelas: 1}, want: ErrParcelasInvalidas},
	}

	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			_, err := NewDespesaService(novoDespesaRepositoryFake()).Criar(context.Background(), test.req)
			if !errors.Is(err, test.want) {
				t.Errorf("erro = %v; esperado %v", err, test.want)
			}
		})
	}
}

func TestDespesaServiceMapeiaRegistroInexistente(t *testing.T) {
	svc := NewDespesaService(novoDespesaRepositoryFake())
	if _, err := svc.Editar(context.Background(), 1, models.DespesaRequest{Descricao: "Compra", Valor: 10}); !errors.Is(err, ErrDespesaInexistente) {
		t.Errorf("erro ao editar = %v; esperado %v", err, ErrDespesaInexistente)
	}
	if _, err := svc.Excluir(context.Background(), 1, false); !errors.Is(err, ErrDespesaInexistente) {
		t.Errorf("erro ao excluir = %v; esperado %v", err, ErrDespesaInexistente)
	}
}
