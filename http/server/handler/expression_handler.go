package handler

import (
	"net/http"
	"strconv"

	"github.com/PavelBradnitski/calc_go/internal/services"
	"github.com/PavelBradnitski/calc_go/pkg/calculation"
	"github.com/gin-gonic/gin"
)

type RateHandler struct {
	service services.RateServiceInterface
}

func NewRateHandler(service services.RateServiceInterface) *RateHandler {
	return &RateHandler{service: service}
}
func (h *RateHandler) RegisterRoutes(router *gin.Engine) {
	rateGroup := router.Group("/api/v1")
	{
		rateGroup.POST("/calculate/", h.Add)
		rateGroup.GET("/expressions/", h.Get)
		rateGroup.GET("/expressions/:id", h.GetByID)
	}
}

type Request struct {
	Expression string `json:"expression"`
}

func (h *RateHandler) Add(c *gin.Context) {
	request := new(Request)
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	expressionInSlice, err := calculation.ParseExpression(request.Expression)
	if err != nil {
		m := map[string]interface{}{"error": calculation.ErrInvalidExpression}
		c.JSON(http.StatusUnprocessableEntity, m)
		return
	}
	result, err := calculation.Calculator(expressionInSlice)
	if err != nil {
		m := map[string]interface{}{"error": calculation.ErrInvalidExpression}
		c.JSON(http.StatusUnprocessableEntity, m)
		return
	}
	id, err := h.service.Add(c, result)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rates"})
		return
	}
	if id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID is blank"})
		return
	}
	m := map[string]interface{}{"id": id}
	c.JSON(http.StatusCreated, m)
}
func (h *RateHandler) Get(c *gin.Context) {
	rates, err := h.service.Get(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rates"})
		return
	}
	if len(rates) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rates not found"})
		return
	}
	c.JSON(http.StatusOK, rates)
}

func (h *RateHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	n, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect id format. Expect number"})
		return
	}
	foundExpression, err := h.service.GetById(c, n)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expression not found"})
		return
	}
	if foundExpression != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expression not found"})
		return
	}
	c.JSON(http.StatusOK, foundExpression)
}
