package users

import (
	"BookStore/internal/control/model"
	dbmodel "BookStore/internal/database/model"
	"BookStore/internal/database/table"
	"fmt"
)

type UserService interface {
	Users() ([]*dbmodel.User, error)
	GetUser(id int) (*dbmodel.User, error)
	DeleteUser(id int) error
	UpdateRole(cred model.SetRole) error
	GetRoles() ([]*dbmodel.Role, error)
	GetRole(id int) (*dbmodel.Role, error)
}

type Option func(service *userService)

type userService struct{}

func NewService(opts ...Option) UserService {
	s := userService{}
	for _, opt := range opts {
		opt(&s)
	}
	return &s
}

func (a *userService) UpdateRole(cred model.SetRole) error {
	user, err := table.GetUserByID(cred.UserId)
	if err != nil {
		return err
	}

	if user.RoleID == cred.RoleId {
		return fmt.Errorf("role already exists")
	}

	if err = table.UpdateRole(user.Login, cred.RoleId); err != nil {
		return err
	}

	return nil
}

func (a *userService) GetRoles() ([]*dbmodel.Role, error) {
	roles, err := table.GetRoles()
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (a *userService) GetRole(id int) (*dbmodel.Role, error) {
	role, err := table.GetRoleByID(id)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (a *userService) DeleteUser(id int) error {
	if err := table.DeleteUser(id); err != nil {
		return err
	}

	return nil
}

func (a *userService) Users() ([]*dbmodel.User, error) {
	users, err := table.GetUsers()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

func (a *userService) GetUser(id int) (*dbmodel.User, error) {
	user, err := table.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	user.Password = ""

	return user, nil
}
