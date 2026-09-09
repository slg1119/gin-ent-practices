package bootstrap

import (
	"gin-start/src/ent"
	"gin-start/src/internal/user"
	"gin-start/src/internal/user/entrepo"
	userhttp "gin-start/src/internal/user/http"
)

// UserModule은 사용자 기능 외부에서 필요한 서비스와 HTTP 핸들러를 묶는다.
type UserModule struct {
	UserService *user.UserService
	UserHandler *userhttp.UserHandler
}

// NewUserModule은 공유 Ent 클라이언트로 사용자 저장소, 서비스, 핸들러를 조립한다.
func NewUserModule(entClient *ent.Client) UserModule {
	userRepository := entrepo.NewUserRepository(entClient)
	userService := user.NewUserService(userRepository)
	userHandler := userhttp.NewUserHandler(userService)

	return UserModule{
		UserService: userService,
		UserHandler: userHandler,
	}
}
