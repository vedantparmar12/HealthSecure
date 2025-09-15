package handlers

import (
	"net/http"

	"healthsecure/internal/models"
	"healthsecure/internal/services"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	patientService *services.PatientService
}

func NewAIHandler(patientService *services.PatientService) *AIHandler {
	return &AIHandler{
		patientService: patientService,
	}
}

func (h *AIHandler) GetPatients(c *gin.Context) {
	patients, err := h.patientService.GetAllPatients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get patients"})
		return
	}

	// Convert to AI-safe format
	var aiPatients []models.AIPatient
	for _, p := range patients {
		aiPatients = append(aiPatients, models.AIPatient{
			ID:          p.ID,
			Name:        p.GetFullName(),
			Age:         p.GetAge(),
			Description: p.Description,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, aiPatients)
}
