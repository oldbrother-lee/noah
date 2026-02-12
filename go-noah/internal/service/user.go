package service

import (
	"context"
	"errors"
	"go-noah/api"
	"go-noah/internal/model"
	"go-noah/internal/repository"
	"go-noah/pkg/global"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService 用户业务逻辑层
type UserService struct{}

var UserServiceApp = new(UserService)

// getUserRepo 获取 UserRepository（在方法内部创建）
func (s *UserService) getUserRepo() *repository.UserRepository {
	// 直接创建 Repository，避免循环导入
	repo := repository.NewRepository(global.Logger, global.DB, global.Enforcer)
	return repository.NewUserRepository(repo)
}

func (s *UserService) GetUser(ctx context.Context, uid uint64) (*model.User, error) {
	repo := s.getUserRepo()
	return repo.GetUser(ctx, uid)
}

func (s *UserService) GetUsers(ctx context.Context, req *api.GetUsersRequest) (*api.GetUsersResponseData, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}

	repo := s.getUserRepo()
	users, total, err := repo.GetUsers(ctx, req.Page, req.PageSize, req.Search, req.OrganizationKey, req.RoleID)
	if err != nil {
		return nil, err
	}

	list := make([]api.UserData, 0, len(users))
	for _, user := range users {
		lastLogin := ""
		if user.LastLogin != nil {
			lastLogin = user.LastLogin.Format("2006-01-02 15:04:05")
		}
		dateJoined := ""
		if user.DateJoined != nil {
			dateJoined = user.DateJoined.Format("2006-01-02 15:04:05")
		}

		list = append(list, api.UserData{
			Uid:         user.Uid,
			Username:    user.Username,
			Email:       user.Email,
			NickName:    user.NickName,
			Mobile:      user.Mobile,
			AvatarFile:  user.AvatarFile,
			RoleID:      user.RoleID,
			IsSuperuser: user.IsSuperuser,
			IsActive:    user.IsActive,
			IsStaff:     user.IsStaff,
			IsTwoFA:     user.IsTwoFA,
			LastLogin:   lastLogin,
			DateJoined:  dateJoined,
			UpdatedAt:   user.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &api.GetUsersResponseData{
		List:  list,
		Total: total,
	}, nil
}

func (s *UserService) CreateUser(ctx context.Context, req *api.CreateUserRequest) error {
	repo := s.getUserRepo()

	// 检查用户名是否已存在
	existing, _ := repo.GetUserByUsername(ctx, req.Username)
	if existing != nil {
		return errors.New("用户已存在")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &model.User{
		Username:    req.Username,
		Password:    string(hashedPassword),
		Email:       req.Email,
		NickName:    req.NickName,
		Mobile:      req.Mobile,
		RoleID:      req.RoleID,
		IsTwoFA:     req.IsTwoFA,
		IsSuperuser: req.IsSuperuser,
		IsActive:    req.IsActive,
		IsStaff:     false,
		AvatarFile:  "/static/avatar2.jpg",
		OtpSecret:   "",
	}

	return repo.CreateUser(ctx, user)
}

func (s *UserService) UpdateUser(ctx context.Context, uid uint64, req *api.UpdateUserRequest) error {
	repo := s.getUserRepo()

	// 检查用户是否存在
	user, err := repo.GetUser(ctx, uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("用户不存在")
		}
		return err
	}

	// 如果用户名改变，检查新用户名是否已存在
	if req.Username != user.Username {
		existing, _ := repo.GetUserByUsername(ctx, req.Username)
		if existing != nil {
			return errors.New("用户名已存在")
		}
	}

	updates := map[string]interface{}{
		"username":     req.Username,
		"email":        req.Email,
		"nick_name":    req.NickName,
		"mobile":       req.Mobile,
		"role_id":      req.RoleID,
		"is_two_fa":    req.IsTwoFA,
		"is_superuser": req.IsSuperuser,
		"is_active":    req.IsActive,
	}

	return repo.UpdateUser(ctx, uid, updates)
}

func (s *UserService) DeleteUser(ctx context.Context, uid uint64) error {
	repo := s.getUserRepo()
	return repo.DeleteUser(ctx, uid)
}

func (s *UserService) ChangePassword(ctx context.Context, req *api.ChangePasswordRequest) error {
	if req.Password != req.VerifyPassword {
		return errors.New("两次输入的密码不一致")
	}

	repo := s.getUserRepo()

	// 检查用户是否存在
	_, err := repo.GetUser(ctx, req.UID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("用户不存在")
		}
		return err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return repo.UpdatePassword(ctx, req.UID, string(hashedPassword))
}
