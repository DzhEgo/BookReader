package database

import (
	"BookStore/internal/database/model"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

var db *gorm.DB

func Connect() error {
	var err error
	dbConn := os.Getenv("DB_CONN")
	if dbConn == "" {
		return fmt.Errorf("DB_CONN environment variable not set")
	}
	db, err = gorm.Open(postgres.Open(dbConn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Connected to database")
	migrate()

	return nil
}

func GetDB() *gorm.DB {
	return db
}

func migrate() {
	if err := db.AutoMigrate(&model.Book{}, &model.User{}, &model.Role{}, &model.ReadingProgress{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	initRoles()
	initAdmin()
}

func initRoles() {
	roles := []string{"admin", "super", "user"}

	for _, role := range roles {
		if err := GetDB().Where("role_name = ?", role).First(&model.Role{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				db.Create(&model.Role{RoleName: role})
				log.Printf("role %s created", role)
			}
		}
	}
}

func initAdmin() {
	var user model.User
	var role model.Role

	adminName := os.Getenv("ADMIN_NAME")
	password := os.Getenv("ADMIN_PASS")

	if adminName == "" || password == "" {
		log.Fatalln("admin name or password is required")
		return
	}

	err := GetDB().Where("login = ?", adminName).First(&user).Error
	if err == nil {
		log.Println("Admin already exists")
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to check admin user: %v", err)
		return
	}

	err = GetDB().Where("role_name = ?", "admin").First(&role).Error
	if err != nil {
		log.Fatal("admin role not found")
		return
	}

	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
		return
	}

	admin := &model.User{
		Login:    adminName,
		Password: string(hashPass),
		RoleID:   role.ID,
	}

	if err := GetDB().Create(&admin).Error; err != nil {
		log.Fatalf("failed to create admin: %v", err)
		return
	}

	log.Println("Admin created")
}
