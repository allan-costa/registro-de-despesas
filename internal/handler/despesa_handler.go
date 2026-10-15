package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"registro-despesas/internal/models"
	"registro-despesas/internal/service"
)

type DespesaHandler struct {
	svc *service.DespesaService
}

func NewDespesaHandler(svc *service.DespesaService) *DespesaHandler {
	return &DespesaHandler{svc: svc}
}

// Salvar cria (201, lista das despesas geradas) ou, com id no corpo, edita descrição e valor (200).
func (h *DespesaHandler) Salvar(c *gin.Context) {
	var req models.DespesaRequest
	if !bindJSON(c, &req, "informe a descrição, um valor maior que zero e um tipo válido (PADRAO, EVENTUAL, RECORRENTE ou PARCELAMENTO)") {
		return
	}

	var (
		result any
		status = http.StatusCreated
		err    error
	)
	if req.ID != nil {
		status = http.StatusOK
		result, err = h.svc.Editar(c.Request.Context(), *req.ID, req)
	} else {
		result, err = h.svc.Criar(c.Request.Context(), req)
	}

	switch {
	case err == nil:
		c.JSON(status, result)
	case errors.Is(err, service.ErrDescricaoInvalida),
		errors.Is(err, service.ErrValorInvalido),
		errors.Is(err, service.ErrDataInvalida),
		errors.Is(err, service.ErrTipoInvalido),
		errors.Is(err, service.ErrParcelasInvalidas),
		errors.Is(err, service.ErrCategoriaInexistente):
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
	case errors.Is(err, service.ErrDespesaInexistente):
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
	default:
		internalError(c, err)
	}
}

// Excluir remove uma despesa ou, com ?excluir_grupo=true, todo o lote parcelado/recorrente.
func (h *DespesaHandler) Excluir(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	n, err := h.svc.Excluir(c.Request.Context(), id, c.Query("excluir_grupo") == "true")
	switch {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"mensagem": "despesa excluída com sucesso", "excluidos": n})
	case errors.Is(err, service.ErrDespesaInexistente):
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
	default:
		internalError(c, err)
	}
}
