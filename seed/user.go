package seed

import (
	"log"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB) {
	log.Println("[SEED] Seeding users with professional RBAC roles...")

	users := []struct {
		Username    string
		Password    string
		Role        string
		CompanyCode string
	}{
		// Platform level
		{Username: "superadmin", Password: "rahasia123", Role: model.RoleSuperadmin},
		{Username: "tech_admin", Password: "tech123", Role: model.RoleTechAdmin},
		{Username: "admin", Password: "admin123", Role: model.RoleAdmin},

		// Company level (HCML)
		{Username: "hcml_admin", Password: "company123", Role: model.RoleCompanyAdmin, CompanyCode: "hcml"},
		{Username: "hcml_operator", Password: "operator123", Role: model.RoleOperator, CompanyCode: "hcml"},
		{Username: "hcml_viewer", Password: "viewer123", Role: model.RoleViewer, CompanyCode: "hcml"},

		// Company level (Meratus Shipping - for cross-tenant isolation testing)
		{Username: "meratus_operator", Password: "operator123", Role: model.RoleOperator, CompanyCode: "meratus"},

		// Legacy aliases
		{Username: "logistic", Password: "logistic", Role: model.RoleAdmin},
		{Username: "hcml_guest", Password: "hcml_guest", Role: model.RoleViewer, CompanyCode: "hcml"},
	}

	for _, u := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[SEED] Error hashing password for %s: %v", u.Username, err)
			continue
		}

		var compID *uint
		if u.CompanyCode != "" {
			var comp model.MCompany
			if err := db.Where("code = ?", u.CompanyCode).First(&comp).Error; err == nil {
				compID = &comp.ID
			}
		}

		var existing model.User
		if err := db.Where("username = ?", u.Username).First(&existing).Error; err == nil {
			// Update existing user with new password and role
			existing.Password = string(hashedPassword)
			existing.Role = u.Role
			existing.CompanyID = compID
			db.Save(&existing)
			log.Printf("[SEED] Updated user %s (Role: %s)", u.Username, u.Role)
		} else {
			// Create user
			newUser := model.User{
				Username:  u.Username,
				Password:  string(hashedPassword),
				Role:      u.Role,
				CompanyID: compID,
			}
			if err := db.Create(&newUser).Error; err != nil {
				log.Printf("[SEED] Failed to create user %s: %v", u.Username, err)
			} else {
				log.Printf("[SEED] Created user: %s (Role: %s)", u.Username, u.Role)
			}
		}
	}

	log.Println("[SEED] User seeding completed")
}
