package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"registro-despesas/internal/middleware"
	"registro-despesas/internal/models"
	"registro-despesas/internal/service"
)

const maxBodyBytes = 1 << 20

// bindJSON limita o corpo e responde 400 com msg se o JSON ou a validação falhar.
func bindJSON(c *gin.Context, dst any, msg string) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBodyBytes)
	if err := c.ShouldBindJSON(dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": msg})
		return false
	}
	return true
}

func internalError(c *gin.Context, err error) {
	log.Printf("erro interno em %s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro interno do servidor"})
}

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if !bindJSON(c, &req, "informe e-mail e senha válidos") {
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req)
	if errors.Is(err, service.ErrCredenciais) {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": err.Error()})
		return
	}
	if err != nil {
		internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) AlterarSenha(c *gin.Context) {
	var req models.AlterarSenhaRequest
	if !bindJSON(c, &req, "informe a senha atual e uma nova senha com no mínimo 6 caracteres") {
		return
	}

	err := h.svc.AlterarSenha(c.Request.Context(), c.GetInt(middleware.CtxUserID), req)
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"mensagem": "senha alterada com sucesso"})
	case errors.Is(err, service.ErrSenhaAtual),
		errors.Is(err, service.ErrSenhaCurta),
		errors.Is(err, service.ErrSenhaLonga):
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
	case errors.Is(err, service.ErrUsuarioInexistente):
		c.JSON(http.StatusUnauthorized, gin.H{"erro": err.Error()})
	default:
		internalError(c, err)
	}
}
