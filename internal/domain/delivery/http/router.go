package http

func (s *Server) createController() {
	s.serv.GET("/health", s.Health)

	s.serv.POST("/team/add", s.HandleTeamAdd)
	s.serv.GET("/team/get", s.HandleTeamGet)
	s.serv.POST("/team/bulkDeactivate", s.HandleTeamBulkDeactivate)

	s.serv.POST("/users/setIsActive", s.HandleSetIsActive)
	s.serv.GET("/users/getReview", s.HandleGetUserReview)

	s.serv.POST("/pullRequest/create", s.HandlePullRequestCreate)
	s.serv.POST("/pullRequest/merge", s.HandlePullRequestMerge)
	s.serv.POST("/pullRequest/reassign", s.HandlePullRequestReassign)

	s.serv.GET("/stats/assignments", s.HandleAssignmentsStats)
}
