package controller

import (
	"net/http"
	"validacao-relato/src/model"
	"validacao-relato/src/service"

	"github.com/gin-gonic/gin"
)

type RelatoController struct {
	svc *service.RelatoService
}

func NewRelatoController(r *gin.Engine, svc *service.RelatoService) {
	controller := &RelatoController{svc: svc}

	// Agrupamento de rotas
	api := r.Group("/relatos")
	{
		api.POST("", controller.CriarRelato)
		api.GET("", controller.ListarRelatos)
		api.GET("/:id", controller.BuscarPorID)
		api.PUT("/:id", controller.AtualizarRelato)
		api.DELETE("/:id", controller.DeletarRelato)
		api.POST("/:id/votar", controller.VotarRelato)
		api.PATCH("/:id/estado", controller.AlterarEstado)
	}
}

// Structs auxiliares para receber payloads de entrada (DTOs)
type CriarRelatoRequest struct {
	ID        string  `json:"id" binding:"required"`
	UsuarioID string  `json:"usuario_id" binding:"required"`
	Tipo      string  `json:"tipo" binding:"required"`
	Descricao string  `json:"descricao" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

type VotarRequest struct {
	Positivo bool `json:"positivo"`
}

type AlterarEstadoRequest struct {
	Estado model.EstadoRelato `json:"estado" binding:"required"`
}

func (h *RelatoController) CriarRelato(c *gin.Context) {
	var req CriarRelatoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	relato, err := h.svc.CriarRelato(c.Request.Context(), req.ID, req.UsuarioID, req.Tipo, req.Descricao, req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, relato)
}

func (h *RelatoController) ListarRelatos(c *gin.Context) {
	relatos, err := h.svc.ListarRelatos(c.Request.Context(), 10, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, relatos)
}

func (h *RelatoController) BuscarPorID(c *gin.Context) {
	id := c.Param("id")
	relato, err := h.svc.BuscarPorID(c.Request.Context(), id)
	if err != nil || relato == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Relato não encontrado"})
		return
	}

	c.JSON(http.StatusOK, relato)
}

func (h *RelatoController) AtualizarRelato(c *gin.Context) {
	id := c.Param("id")
	var req CriarRelatoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	relato, err := h.svc.AtualizarRelato(c.Request.Context(), id, req.Tipo, req.Descricao, req.Latitude, req.Longitude, model.PossivelRelato)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, relato)
}

func (h *RelatoController) DeletarRelato(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.DeletarRelato(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Relato não encontrado ou erro ao deletar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Relato deletado com sucesso"})
}

func (h *RelatoController) VotarRelato(c *gin.Context) {
	id := c.Param("id")
	var req VotarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.svc.VotarRelato(c.Request.Context(), id, req.Positivo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Voto registrado com sucesso"})
}

func (h *RelatoController) AlterarEstado(c *gin.Context) {
	id := c.Param("id")
	var req AlterarEstadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Busca o relato existente para manter os demais dados (tipo, descrição, coordenadas)
	relatoAtual, err := h.svc.BuscarPorID(c.Request.Context(), id)
	if err != nil || relatoAtual == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Relato não encontrado"})
		return
	}

	// 2. Chama o método de atualização geral já existente no service
	_, err = h.svc.AtualizarRelato(
		c.Request.Context(),
		id,
		relatoAtual.Tipo,
		relatoAtual.Descricao,
		relatoAtual.Latitude,
		relatoAtual.Longitude,
		req.Estado, // Passa o novo estado recebido na requisição
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Estado alterado com sucesso"})
}
