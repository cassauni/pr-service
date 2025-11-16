package http

import (
	"net/http"
	"pr-service/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleTeamAdd(c *gin.Context) {
	var req entities.Team
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	team, err := s.Usecase.CreateTeam(c.Request.Context(), req)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": team,
	})
}

func (s *Server) HandleTeamGet(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	team, err := s.Usecase.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}
