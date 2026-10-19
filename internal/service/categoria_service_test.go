package service

import (
	"context"
	"errors"
	"testing"

	"registro-despesas/internal/models"
	"registro-despesas/internal/repository"
)

type categoriaRepositoryFake struct {
	categorias map[int]*models.Categoria
	proximoID  int
	erroDelete error
}

func novoCategoriaRepositoryFake(categorias ...*models.Categoria) *categoriaRepositoryFake {
	fake := &categoriaRepositoryFake{
		categorias: make(map[int]*models.Categoria),
		proximoID:  1,
	}
	for _, categoria := range categorias {
		fake.categorias[categoria.ID] = categoria
		if categoria.ID >= fake.proximoID {
			fake.proximoID = categoria.ID + 1
		}
	}
	return fake
}

func (f *categoriaRepositoryFake) ListAll(context.Context) ([]*models.Categoria, error) {
	lista := make([]*models.Categoria, 0, len(f.categorias))
	for _, categoria := range f.categorias {
		lista = append(lista, categoria)
	}
	return lista, nil
}

func (f *categoriaRepositoryFake) FindByID(_ context.Context, id int) (*models.Categoria, error) {
	categoria, ok := f.categorias[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return categoria, nil
}

func (f *categoriaRepositoryFake) IsAncestorOrSelf(_ context.Context, ancestorID, nodeID int) (bool, error) {
	for nodeID != 0 {
		if nodeID == ancestorID {
			return true, nil
		}
		categoria := f.categorias[nodeID]
		if categoria == nil || categoria.ParentID == nil {
			break
		}
		nodeID = *categoria.ParentID
	}
	return false, nil
}

func (f *categoriaRepositoryFake) Create(_ context.Context, nome string, parentID *int) (*models.Categoria, error) {
	categoria := &models.Categoria{ID: f.proximoID, Nome: nome, ParentID: parentID}
	f.proximoID++
	f.categorias[categoria.ID] = categoria
	return categoria, nil
}

func (f *categoriaRepositoryFake) Update(_ context.Context, id int, nome string, parentID *int) (*models.Categoria, error) {
	categoria, ok := f.categorias[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	categoria.Nome = nome
	categoria.ParentID = parentID
	return categoria, nil
}

func (f *categoriaRepositoryFake) Delete(_ context.Context, id int) error {
	if f.erroDelete != nil {
		return f.erroDelete
	}
	if _, ok := f.categorias[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.categorias, id)
	return nil
}

func TestCategoriaServiceCRUD(t *testing.T) {
	ctx := context.Background()
	repo := novoCategoriaRepositoryFake()
	service := NewCategoriaService(repo)

	criada, err := service.Salvar(ctx, models.CategoriaRequest{Nome: "  mercado  "})
	if err != nil {
		t.Fatalf("Salvar para criar retornou erro: %v", err)
	}
	if criada.ID != 1 || criada.Nome != "MERCADO" {
		t.Fatalf("categoria criada = %+v; esperado ID 1 e nome MERCADO", criada)
	}

	id := criada.ID
	atualizada, err := service.Salvar(ctx, models.CategoriaRequest{ID: &id, Nome: " feira "})
	if err != nil {
		t.Fatalf("Salvar para editar retornou erro: %v", err)
	}
	if atualizada.Nome != "FEIRA" {
		t.Errorf("nome após edição = %q; esperado FEIRA", atualizada.Nome)
	}

	lista, err := service.Listar(ctx)
	if err != nil {
		t.Fatalf("Listar retornou erro: %v", err)
	}
	if len(lista) != 1 || lista[0].ID != id || lista[0].Nome != "FEIRA" {
		t.Fatalf("lista após edição = %+v; esperada uma categoria FEIRA", lista)
	}

	if err := service.Excluir(ctx, id); err != nil {
		t.Fatalf("Excluir retornou erro: %v", err)
	}
	lista, err = service.Listar(ctx)
	if err != nil {
		t.Fatalf("Listar após exclusão retornou erro: %v", err)
	}
	if len(lista) != 0 {
		t.Errorf("lista após exclusão contém %d categorias; esperado 0", len(lista))
	}
}

func TestCategoriaServiceListarMontaHierarquia(t *testing.T) {
	parentID := 1
	repo := novoCategoriaRepositoryFake(
		&models.Categoria{ID: 1, Nome: "CASA"},
		&models.Categoria{ID: 2, ParentID: &parentID, Nome: "CONTAS"},
	)

	lista, err := NewCategoriaService(repo).Listar(context.Background())
	if err != nil {
		t.Fatalf("Listar retornou erro: %v", err)
	}
	if len(lista) != 1 || lista[0].ID != 1 {
		t.Fatalf("raízes = %+v; esperada apenas a categoria CASA", lista)
	}
	if len(lista[0].Filhos) != 1 || lista[0].Filhos[0].ID != 2 {
		t.Fatalf("filhos de CASA = %+v; esperado CONTAS", lista[0].Filhos)
	}
}

func TestCategoriaServiceSalvarValidaEntrada(t *testing.T) {
	parentID := 99
	id := 1
	childID := 2
	repo := novoCategoriaRepositoryFake(
		&models.Categoria{ID: id, Nome: "CASA"},
		&models.Categoria{ID: childID, ParentID: &id, Nome: "CONTAS"},
	)
	service := NewCategoriaService(repo)
	tests := []struct {
		nome string
		req  models.CategoriaRequest
		want error
	}{
		{nome: "nome vazio", req: models.CategoriaRequest{Nome: "  "}, want: ErrNomeVazio},
		{nome: "pai inexistente", req: models.CategoriaRequest{Nome: "CONTAS", ParentID: &parentID}, want: ErrPaiInexistente},
		{nome: "ciclo", req: models.CategoriaRequest{ID: &id, Nome: "CASA", ParentID: &childID}, want: ErrCicloCategoria},
		{nome: "categoria editada inexistente", req: models.CategoriaRequest{ID: &parentID, Nome: "NOVA"}, want: ErrCategoriaInexistente},
	}

	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			if _, err := service.Salvar(context.Background(), test.req); !errors.Is(err, test.want) {
				t.Errorf("erro = %v; esperado %v", err, test.want)
			}
		})
	}
}

func TestCategoriaServiceExcluirTraduzErrosDoRepositorio(t *testing.T) {
	tests := []struct {
		nome    string
		repoErr error
		want    error
	}{
		{nome: "inexistente", repoErr: repository.ErrNotFound, want: ErrCategoriaInexistente},
		{nome: "vinculada a despesas", repoErr: repository.ErrVinculado, want: ErrCategoriaComDespesas},
	}

	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			repo := novoCategoriaRepositoryFake()
			repo.erroDelete = test.repoErr
			err := NewCategoriaService(repo).Excluir(context.Background(), 1)
			if !errors.Is(err, test.want) {
				t.Errorf("erro = %v; esperado %v", err, test.want)
			}
		})
	}
}
