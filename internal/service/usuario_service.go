package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"registro-despesas/internal/models"
	"registro-despesas/internal/repository"
)

var (
	ErrNomeUsuarioVazio = errors.New("o nome é obrigatório")
	ErrEmailDuplicado   = errors.New("já existe um usuário com este e-mail")
	ErrSenhaObrigatoria = errors.New("a senha é obrigatória para novos usuários")
	ErrSenhaConfirmacao = errors.New("a confirmação da senha não confere")
	ErrAutoRebaixamento = errors.New("você não pode rebaixar o seu próprio nível de acesso")
	ErrAutoExclusao     = errors.New("você não pode excluir a si mesmo")
)

type UsuarioService struct {
	repo *repository.UsuarioRepository
}

func NewUsuarioService(repo *repository.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repo: repo}
}

func (s *UsuarioService) Listar(ctx context.Context) ([]*models.Usuario, error) {
	return s.repo.ListAll(ctx)
}

// Salvar cria o usuário ou, com req.ID presente, edita (senha vazia mantém a atual).
// actorID é o usuário autenticado, usado para impedir o autorrebaixamento.
func (s *UsuarioService) Salvar(ctx context.Context, actorID int, req models.UsuarioRequest) (*models.Usuario, error) {
	nome := strings.TrimSpace(req.Nome)
	if nome == "" {
		return nil, ErrNomeUsuarioVazio
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	editing := req.ID != nil

	if editing && *req.ID == actorID && req.Nivel != models.NivelAdmin {
		return nil, ErrAutoRebaixamento
	}

	var hash *string
	if !editing && req.Senha == "" {
		return nil, ErrSenhaObrigatoria
	}
	if req.Senha != "" {
		if err := validarSenha(req.Senha); err != nil {
			return nil, err
		}
		if req.Senha != req.ConfirmarSenha {
			return nil, ErrSenhaConfirmacao
		}
		b, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("gerar hash: %w", err)
		}
		h := string(b)
		hash = &h
	}

	excludeID := 0
	if editing {
		excludeID = *req.ID
	}
	exists, err := s.repo.EmailExists(ctx, email, excludeID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailDuplicado
	}

	var u *models.Usuario
	if editing {
		u, err = s.repo.Update(ctx, *req.ID, nome, email, req.Nivel, hash)
	} else {
		u, err = s.repo.Create(ctx, nome, email, *hash, req.Nivel)
	}
	switch {
	case errors.Is(err, repository.ErrDuplicado):
		return nil, ErrEmailDuplicado
	case errors.Is(err, repository.ErrNotFound):
		return nil, ErrUsuarioInexistente
	}
	return u, err
}

func (s *UsuarioService) Excluir(ctx context.Context, actorID, id int) error {
	if id == actorID {
		return ErrAutoExclusao
	}
	if err := s.repo.Delete(ctx, id); errors.Is(err, repository.ErrNotFound) {
		return ErrUsuarioInexistente
	} else if err != nil {
		return err
	}
	return nil
}
