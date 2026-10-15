package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"registro-despesas/internal/models"
	"registro-despesas/internal/service"
)

type CategoriaHandler struct {
	svc *service.CategoriaService
}

func NewCategoriaHandler(svc *service.CategoriaService) *CategoriaHandler {
	return &CategoriaHandler{svc: svc}
}

func (h *CategoriaHandler) Listar(c *gin.Context) {
	tree, err := h.svc.Listar(c.Request.Context())
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, tree)
}

// Salvar cria (201) ou, com id no corpo, edita (200).
func (h *CategoriaHandler) Salvar(c *gin.Context) {
	var req models.CategoriaRequest
	if !bindJSON(c, &req, "informe o nome da categoria") {
		return
	}

	cat, err := h.svc.Salvar(c.Request.Context(), req)
	switch {
	case err == nil:
		status := http.StatusCreated
		if req.ID != nil {
			status = http.StatusOK
		}
		c.JSON(status, cat)
	case errors.Is(err, service.ErrNomeVazio),
		errors.Is(err, service.ErrPaiInexistente),
		errors.Is(err, service.ErrCicloCategoria):
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
	case errors.Is(err, service.ErrCategoriaInexistente):
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
	default:
		internalError(c, err)
	}
}

func (h *CategoriaHandler) Excluir(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	err = h.svc.Excluir(c.Request.Context(), id)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"mensagem": "categoria excluída com sucesso"})
	case errors.Is(err, service.ErrCategoriaInexistente):
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
	case errors.Is(err, service.ErrCategoriaComDespesas):
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
	default:
		internalError(c, err)
	}
}
