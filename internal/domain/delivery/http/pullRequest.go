package http

import (
	"net/http"
	"pr-service/internal/domain/entities"

	"github.com/gin-gonic/gin"
)

func (s *Server) HandlePullRequestCreate(c *gin.Context) {
	var req entities.CreatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	pr, err := s.Usecase.CreatePullRequest(c.Request.Context(), req)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"pr": pr,
	})
}

func (s *Server) HandlePullRequestMerge(c *gin.Context) {
	var req entities.MergePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	pr, err := s.Usecase.MergePullRequest(c.Request.Context(), req.PullRequestID)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr": pr,
	})
}

func (s *Server) HandlePullRequestReassign(c *gin.Context) {
	var req entities.ReassignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	pr, replacedBy, err := s.Usecase.ReassignReviewer(
		c.Request.Context(),
		req.PullRequestID,
		req.OldReviewerID,
	)
	if err != nil {
		s.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": replacedBy,
	})
}
