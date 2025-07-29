package auth

import (
	"BookStore/internal/control/model"
	dbmodel "BookStore/internal/database/model"
	"BookStore/internal/database/table"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	CreateUser(user model.Creditionals) error
	GetUser(login model.Creditionals) (*dbmodel.User, error)
	GetUserContext(login model.Creditionals) (*model.UserContext, error)
	ValidateUser(dbUser *dbmodel.User, cred model.Creditionals) (bool, error)
}
type Option func(service *authService)

type authService struct{}

func NewService(opts ...Option) AuthService {
	s := authService{}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

func (a *authService) CreateUser(user model.Creditionals) error {
	hash, err := a.hashPass(user)
	if err != nil {
		return err
	}

	roleId, err := table.GetRoleID("user")
	if err != nil {
		return err
	}

	userDb := &dbmodel.User{
		Login:    user.Login,
		Password: hash,
		Email:    user.Email,
		RoleID:   roleId,
	}

	if err = table.Upsert(userDb); err != nil {
		return err
	}

	return nil
}

func (a *authService) hashPass(user model.Creditionals) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 0)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (a *authService) GetUser(login model.Creditionals) (*dbmodel.User, error) {
	user, err := table.GetUserByLogin(login.Login)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (a *authService) GetUserContext(login model.Creditionals) (*model.UserContext, error) {
	user, err := table.GetUserByLogin(login.Login)
	if err != nil {
		return nil, err
	}

	userContext := &model.UserContext{
		ID:    user.ID,
		Login: user.Login,
		Role:  user.Role.RoleName,
	}

	return userContext, nil
}

func (a *authService) ValidateUser(dbUser *dbmodel.User, cred model.Creditionals) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(cred.Password)); err != nil {
		return false, err
	}
	return true, nil
}
