package http

import (
	"context"
	"net/http"
	"pr-service/config"
	"pr-service/internal/domain/entities"
	"pr-service/internal/domain/usecase"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	logger  *zap.Logger
	cfg     *config.ConfigModel
	serv    *gin.Engine
	Usecase *usecase.Usecase
}

func NewServer(logger *zap.Logger, cfg *config.ConfigModel, uc *usecase.Usecase) (*Server, error) {
	return &Server{
		logger:  logger,
		cfg:     cfg,
		serv:    gin.Default(),
		Usecase: uc,
	}, nil
}

func (s *Server) OnStart(_ context.Context) error {
	s.createController()
	go func() {
		addr := s.cfg.HTTP.Host + ":" + s.cfg.HTTP.Port
		s.logger.Info("HTTP server started", zap.String("addr", addr))
		if err := s.serv.Run(addr); err != nil {
			s.logger.Error("failed to serve", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) OnStop(_ context.Context) error {
	s.logger.Info("http server stopped")
	return nil
}

func (s *Server) handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	if derr, ok := err.(*entities.DomainError); ok {
		s.writeDomainError(c, derr)
		return
	}

	s.logger.Error("internal error", zap.Error(err))

	c.Status(http.StatusInternalServerError)
}

func (s *Server) writeDomainError(c *gin.Context, derr *entities.DomainError) {
	status := http.StatusBadRequest
	switch derr.Code {
	case entities.ErrorCodeTeamExists:
		status = http.StatusBadRequest
	case entities.ErrorCodePRExists:
		status = http.StatusConflict
	case entities.ErrorCodePRMerged, entities.ErrorCodeNotAssigned, entities.ErrorCodeNoCandidate:
		status = http.StatusConflict
	case entities.ErrorCodeNotFound:
		status = http.StatusNotFound
	}

	c.JSON(status, entities.ErrorResponse{
		Error: entities.ErrorBody{
			Code:    derr.Code,
			Message: derr.Message,
		},
	})
}

func (s *Server) Health(c *gin.Context) {
	c.Status(http.StatusOK)
}
