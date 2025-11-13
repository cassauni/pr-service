package http

import (
	"context"
	"net/http"
	"pr-service/config"
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
	return &Server{logger: logger, cfg: cfg, serv: gin.Default(), Usecase: uc}, nil
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

func (s *Server) OnStop(_ context.Context) error { s.logger.Info("http server stopped"); return nil }

func (s *Server) Helth(c *gin.Context) {
	c.Status(http.StatusOK)
}
