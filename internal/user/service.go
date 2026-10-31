package user

import (
	"errors"

	"github.com/Alugara-Inovasi-Utama/emerti/model"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Create(username, password, role string, companyID *uint) (*UserResponse, error)
	GetAll(companyID *uint) ([]UserResponse, error)
	GetByID(id uint) (*UserResponse, error)
	Update(id uint, username, password, role string, companyID *uint) (*UserResponse, error)
	Delete(id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func toUserResponse(u *model.User) UserResponse {
	res := UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CompanyID: u.CompanyID,
	}
	if u.Company != nil {
		res.CompanyName = u.Company.Name
		res.CompanyCode = u.Company.Code
	}
	return res
}

func (s *service) Create(username, password, role string, companyID *uint) (*UserResponse, error) {
	// Check if user exists
	_, err := s.repo.FindByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &model.User{
		Username:  username,
		Password:  string(hashedPassword),
		Role:      role,
		CompanyID: companyID,
	}

	err = s.repo.Create(user)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	created, err := s.repo.GetByID(user.ID)
	if err != nil {
		resp := toUserResponse(user)
		return &resp, nil
	}

	resp := toUserResponse(created)
	return &resp, nil
}

func (s *service) GetAll(companyID *uint) ([]UserResponse, error) {
	users, err := s.repo.GetAll(companyID)
	if err != nil {
		return nil, err
	}

	var response []UserResponse
	for _, user := range users {
		response = append(response, toUserResponse(&user))
	}

	return response, nil
}

func (s *service) GetByID(id uint) (*UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("User not added yet")
	}

	resp := toUserResponse(user)
	return &resp, nil
}

func (s *service) Update(id uint, username, password, role string, companyID *uint) (*UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("User not found")
	}

	// Update username if provided
	if username != "" {
		// Check if new username already exists
		existingUser, err := s.repo.FindByUsername(username)
		if err == nil && existingUser.ID != id {
			return nil, errors.New("Username already exists")
		}
		user.Username = username
	}

	// Update password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("Failed to hash password")
		}
		user.Password = string(hashedPassword)
	}

	// Update role if provided
	if role != "" {
		user.Role = role
	}

	// Update companyID if provided
	if companyID != nil {
		if *companyID == 0 {
			user.CompanyID = nil
		} else {
			user.CompanyID = companyID
		}
	}

	err = s.repo.Update(user)
	if err != nil {
		return nil, errors.New("Failed to update user")
	}

	updated, err := s.repo.GetByID(user.ID)
	if err != nil {
		resp := toUserResponse(user)
		return &resp, nil
	}

	resp := toUserResponse(updated)
	return &resp, nil
}

func (s *service) Delete(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("User not found")
	}

	return s.repo.Delete(id)
}
