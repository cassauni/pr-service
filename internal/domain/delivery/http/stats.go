package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandleAssignmentsStats(c *gin.Context) {
	resp, err := s.Usecase.GetAssignmentsStats(c.Request.Context())
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
