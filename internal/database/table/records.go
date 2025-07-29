package table

import (
	"BookStore/internal/database"
	"BookStore/internal/database/model"
)

func Upsert(model interface{}) error {
	err := database.GetDB().Save(model).Error
	if err != nil {
		return err
	}
	return nil
}

func GetUserByLogin(username string) (*model.User, error) {
	var user *model.User

	err := database.GetDB().Model(&model.User{}).Where("login = ?", username).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}

	return user, err
}

func GetUserByID(id int) (*model.User, error) {
	var user *model.User
	err := database.GetDB().Model(&model.User{}).Where("id = ?", id).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, err
}

func GetUsers() ([]*model.User, error) {
	var users []*model.User
	err := database.GetDB().Model(&model.User{}).Preload("Role").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, err
}

func GetRoleByID(id int) (*model.Role, error) {
	var role *model.Role
	err := database.GetDB().Model(&model.Role{}).Where("id = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}

	return role, err
}

func GetRoles() ([]*model.Role, error) {
	var roles []*model.Role
	err := database.GetDB().Model(&model.Role{}).Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, err
}

func GetRoleID(roleName string) (int, error) {
	var roleId int
	err := database.GetDB().Model(&model.Role{}).Select("id").Where("role_name = ?", roleName).First(&roleId).Error
	if err != nil {
		return 0, err
	}
	return roleId, nil
}

func UpdateRole(username string, roleId int) error {
	err := database.GetDB().Model(&model.User{}).Where("login = ?", username).Update("role_id", roleId).Error
	if err != nil {
		return err
	}
	return nil
}

func GetBook(id int) (*model.Book, error) {
	var book *model.Book
	err := database.GetDB().Model(&model.Book{}).Where("id = ?", id).First(&book).Error
	if err != nil {
		return nil, err
	}
	return book, err
}

func GetBooks() ([]*model.Book, error) {
	var books []*model.Book
	err := database.GetDB().Model(&model.Book{}).Find(&books).Error
	if err != nil {
		return nil, err
	}
	return books, err
}

func DeleteBook(id int) error {
	if err := database.GetDB().Delete(&model.Book{}, id).Error; err != nil {
		return err
	}
	return nil
}

func DeleteUser(id int) error {
	err := database.GetDB().Model(&model.User{}).Where("id = ?", id).Delete(&model.User{}).Error
	if err != nil {
		return err
	}

	return nil
}
