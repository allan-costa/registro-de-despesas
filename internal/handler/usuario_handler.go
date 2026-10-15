package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"registro-despesas/internal/middleware"
	"registro-despesas/internal/models"
	"registro-despesas/internal/service"
)

type UsuarioHandler struct {
	svc *service.UsuarioService
}

func NewUsuarioHandler(svc *service.UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{svc: svc}
}

func (h *UsuarioHandler) Listar(c *gin.Context) {
	usuarios, err := h.svc.Listar(c.Request.Context())
	if err != nil {
		internalError(c, err)
		return
	}
	c.JSON(http.StatusOK, usuarios)
}

// Salvar cria (201) ou, com id no corpo, edita (200).
func (h *UsuarioHandler) Salvar(c *gin.Context) {
	var req models.UsuarioRequest
	if !bindJSON(c, &req, "informe nome, e-mail válido e nível (ADMIN ou READ)") {
		return
	}

	u, err := h.svc.Salvar(c.Request.Context(), c.GetInt(middleware.CtxUserID), req)
	switch {
	case err == nil:
		status := http.StatusCreated
		if req.ID != nil {
			status = http.StatusOK
		}
		c.JSON(status, u)
	case errors.Is(err, service.ErrNomeUsuarioVazio),
		errors.Is(err, service.ErrSenhaObrigatoria),
		errors.Is(err, service.ErrSenhaConfirmacao),
		errors.Is(err, service.ErrSenhaCurta),
		errors.Is(err, service.ErrSenhaLonga),
		errors.Is(err, service.ErrAutoRebaixamento):
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
	case errors.Is(err, service.ErrEmailDuplicado):
		c.JSON(http.StatusConflict, gin.H{"erro": err.Error()})
	case errors.Is(err, service.ErrUsuarioInexistente):
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
	default:
		internalError(c, err)
	}
}

func (h *UsuarioHandler) Excluir(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	err = h.svc.Excluir(c.Request.Context(), c.GetInt(middleware.CtxUserID), id)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"mensagem": "usuário excluído com sucesso"})
	case errors.Is(err, service.ErrAutoExclusao):
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
	case errors.Is(err, service.ErrUsuarioInexistente):
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
	default:
		internalError(c, err)
	}
}
