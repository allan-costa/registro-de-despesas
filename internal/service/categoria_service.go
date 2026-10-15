package service

import (
	"context"
	"errors"
	"strings"

	"registro-despesas/internal/models"
	"registro-despesas/internal/repository"
)

var (
	ErrNomeVazio            = errors.New("o nome da categoria é obrigatório")
	ErrPaiInexistente       = errors.New("categoria pai não encontrada")
	ErrCicloCategoria       = errors.New("a categoria não pode ser pai de si mesma nem de uma de suas subcategorias")
	ErrCategoriaInexistente = errors.New("categoria não encontrada")
	ErrCategoriaComDespesas = errors.New("não é possível excluir: a categoria (ou uma subcategoria) possui despesas vinculadas")
)

type CategoriaService struct {
	repo *repository.CategoriaRepository
}

func NewCategoriaService(repo *repository.CategoriaRepository) *CategoriaService {
	return &CategoriaService{repo: repo}
}

// Listar devolve as raízes (parent_id nulo) com Filhos populados recursivamente.
func (s *CategoriaService) Listar(ctx context.Context) ([]*models.Categoria, error) {
	all, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	byID := make(map[int]*models.Categoria, len(all))
	for _, c := range all {
		byID[c.ID] = c
	}

	roots := []*models.Categoria{}
	for _, c := range all {
		if c.ParentID == nil {
			roots = append(roots, c)
		} else if p, ok := byID[*c.ParentID]; ok {
			p.Filhos = append(p.Filhos, c)
		}
	}
	return roots, nil
}

// Salvar cria a categoria ou, com req.ID presente, edita a existente.
func (s *CategoriaService) Salvar(ctx context.Context, req models.CategoriaRequest) (*models.Categoria, error) {
	nome := strings.ToUpper(strings.TrimSpace(req.Nome))
	if nome == "" {
		return nil, ErrNomeVazio
	}

	if req.ParentID != nil {
		if _, err := s.repo.FindByID(ctx, *req.ParentID); errors.Is(err, repository.ErrNotFound) {
			return nil, ErrPaiInexistente
		} else if err != nil {
			return nil, err
		}

		if req.ID != nil {
			ciclo, err := s.repo.IsAncestorOrSelf(ctx, *req.ID, *req.ParentID)
			if err != nil {
				return nil, err
			}
			if ciclo {
				return nil, ErrCicloCategoria
			}
		}
	}

	if req.ID == nil {
		return s.repo.Create(ctx, nome, req.ParentID)
	}

	c, err := s.repo.Update(ctx, *req.ID, nome, req.ParentID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrCategoriaInexistente
	}
	return c, err
}

func (s *CategoriaService) Excluir(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return ErrCategoriaInexistente
	case errors.Is(err, repository.ErrVinculado):
		return ErrCategoriaComDespesas
	}
	return err
}
