package view

import "github.com/muety/broilerplate/models"

type DashboardViewModel struct {
	User      *models.User
	AvatarURL string
	Success   string
	Error     string
}

func (s *DashboardViewModel) WithSuccess(m string) *DashboardViewModel {
	s.Success = m
	return s
}

func (s *DashboardViewModel) WithError(m string) *DashboardViewModel {
	s.Error = m
	return s
}
