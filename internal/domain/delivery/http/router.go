package http

func (s *Server) createController() {
	public := s.serv.Group("/pr")

	public.GET("/helth", s.Helth)

}
