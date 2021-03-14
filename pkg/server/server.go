package server

import (
	"github.com/OpenSlides/openslides-manage-service/pkg/create_user"
	"github.com/OpenSlides/openslides-manage-service/pkg/server/serverutil"
	"github.com/OpenSlides/openslides-manage-service/pkg/set_password"
)

type Server struct {
	set_password.SetPassworder
	create_user.CreateUserer
}

func newServer(cfg *serverutil.Config) Server {
	return Server{
		SetPassworder: set_password.SetPassworder{Config: cfg},
	}
}
