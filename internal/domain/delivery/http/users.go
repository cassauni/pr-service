package http

import (
	"net/http"
	"pr-service/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleSetIsActive(c *gin.Context) {
	var req entities.SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := s.Usecase.SetUserIsActive(c.Request.Context(), req.UserID, *req.IsActive)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (s *Server) HandleGetUserReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	resp, err := s.Usecase.GetUserReviews(c.Request.Context(), userID)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
