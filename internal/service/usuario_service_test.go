package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"registro-despesas/internal/models"
	"registro-despesas/internal/repository"
)

type usuarioRepositoryFake struct {
	usuarios  map[int]*models.Usuario
	proximoID int
}

func novoUsuarioRepositoryFake(usuarios ...*models.Usuario) *usuarioRepositoryFake {
	fake := &usuarioRepositoryFake{usuarios: make(map[int]*models.Usuario), proximoID: 1}
	for _, usuario := range usuarios {
		fake.usuarios[usuario.ID] = usuario
		if usuario.ID >= fake.proximoID {
			fake.proximoID = usuario.ID + 1
		}
	}
	return fake
}

func (f *usuarioRepositoryFake) ListAll(context.Context) ([]*models.Usuario, error) {
	lista := make([]*models.Usuario, 0, len(f.usuarios))
	for _, usuario := range f.usuarios {
		lista = append(lista, usuario)
	}
	return lista, nil
}

func (f *usuarioRepositoryFake) EmailExists(_ context.Context, email string, excludeID int) (bool, error) {
	for id, usuario := range f.usuarios {
		if id != excludeID && strings.EqualFold(usuario.Email, email) {
			return true, nil
		}
	}
	return false, nil
}

func (f *usuarioRepositoryFake) Create(_ context.Context, nome, email, hash, nivel string) (*models.Usuario, error) {
	usuario := &models.Usuario{ID: f.proximoID, Nome: nome, Email: email, Senha: hash, Nivel: nivel}
	f.proximoID++
	f.usuarios[usuario.ID] = usuario
	return usuario, nil
}

func (f *usuarioRepositoryFake) Update(_ context.Context, id int, nome, email, nivel string, hash *string) (*models.Usuario, error) {
	usuario, ok := f.usuarios[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	usuario.Nome = nome
	usuario.Email = email
	usuario.Nivel = nivel
	if hash != nil {
		usuario.Senha = *hash
	}
	return usuario, nil
}

func (f *usuarioRepositoryFake) Delete(_ context.Context, id int) error {
	if _, ok := f.usuarios[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.usuarios, id)
	return nil
}

func TestUsuarioServiceCRUD(t *testing.T) {
	ctx := context.Background()
	repo := novoUsuarioRepositoryFake()
	svc := NewUsuarioService(repo)

	criado, err := svc.Salvar(ctx, 10, models.UsuarioRequest{
		Nome: "  Ana Silva  ", Email: " ANA@EXAMPLE.COM ", Senha: "senha123",
		ConfirmarSenha: "senha123", Nivel: models.NivelRead,
	})
	if err != nil {
		t.Fatalf("Salvar para criar retornou erro: %v", err)
	}
	if criado.ID != 1 || criado.Nome != "Ana Silva" || criado.Email != "ana@example.com" {
		t.Fatalf("usuário criado = %+v; esperado nome e e-mail normalizados", criado)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(criado.Senha), []byte("senha123")); err != nil {
		t.Errorf("senha não foi armazenada como hash bcrypt: %v", err)
	}

	lista, err := svc.Listar(ctx)
	if err != nil || len(lista) != 1 || lista[0].ID != criado.ID {
		t.Fatalf("Listar = %+v, erro = %v; esperado um usuário", lista, err)
	}

	id := criado.ID
	atualizado, err := svc.Salvar(ctx, 10, models.UsuarioRequest{
		ID: &id, Nome: "Ana S.", Email: "ana.nova@example.com", Nivel: models.NivelAdmin,
	})
	if err != nil {
		t.Fatalf("Salvar para editar retornou erro: %v", err)
	}
	if atualizado.Nome != "Ana S." || atualizado.Email != "ana.nova@example.com" || atualizado.Senha != criado.Senha {
		t.Errorf("usuário após edição = %+v; dados ou senha inesperados", atualizado)
	}

	if err := svc.Excluir(ctx, 10, id); err != nil {
		t.Fatalf("Excluir retornou erro: %v", err)
	}
	lista, err = svc.Listar(ctx)
	if err != nil || len(lista) != 0 {
		t.Errorf("Listar após exclusão = %+v, erro = %v; esperado lista vazia", lista, err)
	}
}

func TestUsuarioServiceSalvarRejeitaEntradasInvalidas(t *testing.T) {
	usuario := &models.Usuario{ID: 1, Nome: "Admin", Email: "existente@example.com", Nivel: models.NivelAdmin}
	tests := []struct {
		nome string
		req  models.UsuarioRequest
		want error
	}{
		{nome: "nome vazio", req: models.UsuarioRequest{Email: "novo@example.com"}, want: ErrNomeUsuarioVazio},
		{nome: "senha obrigatória", req: models.UsuarioRequest{Nome: "Novo", Email: "novo@example.com"}, want: ErrSenhaObrigatoria},
		{nome: "confirmação divergente", req: models.UsuarioRequest{Nome: "Novo", Email: "novo@example.com", Senha: "senha123", ConfirmarSenha: "outra123", Nivel: models.NivelRead}, want: ErrSenhaConfirmacao},
		{nome: "senha curta", req: models.UsuarioRequest{Nome: "Novo", Email: "novo@example.com", Senha: "123", ConfirmarSenha: "123", Nivel: models.NivelRead}, want: ErrSenhaCurta},
		{nome: "e-mail duplicado", req: models.UsuarioRequest{Nome: "Novo", Email: " EXISTENTE@EXAMPLE.COM ", Senha: "senha123", ConfirmarSenha: "senha123", Nivel: models.NivelRead}, want: ErrEmailDuplicado},
		{nome: "autorrebaixamento", req: models.UsuarioRequest{ID: &usuario.ID, Nome: "Admin", Email: usuario.Email, Nivel: models.NivelRead}, want: ErrAutoRebaixamento},
	}

	for _, test := range tests {
		t.Run(test.nome, func(t *testing.T) {
			repo := novoUsuarioRepositoryFake(&models.Usuario{ID: usuario.ID, Nome: usuario.Nome, Email: usuario.Email, Nivel: usuario.Nivel})
			_, err := NewUsuarioService(repo).Salvar(context.Background(), usuario.ID, test.req)
			if !errors.Is(err, test.want) {
				t.Errorf("erro = %v; esperado %v", err, test.want)
			}
		})
	}
}

func TestUsuarioServiceExcluirProtegeContaAtual(t *testing.T) {
	svc := NewUsuarioService(novoUsuarioRepositoryFake())
	if err := svc.Excluir(context.Background(), 3, 3); !errors.Is(err, ErrAutoExclusao) {
		t.Errorf("erro = %v; esperado %v", err, ErrAutoExclusao)
	}
	if err := svc.Excluir(context.Background(), 3, 4); !errors.Is(err, ErrUsuarioInexistente) {
		t.Errorf("erro = %v; esperado %v", err, ErrUsuarioInexistente)
	}
}
