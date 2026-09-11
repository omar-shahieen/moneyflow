package repository

import "github.com/omar-shahieen/moneyflow/internal/server"

type Repositories struct{}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{}
}
