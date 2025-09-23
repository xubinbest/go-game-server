package user

import (
	"context"
	"fmt"
	"time"

	"github.xubinbest.com/go-game-server/internal/auth"
	"github.xubinbest.com/go-game-server/internal/db/models"
	"github.xubinbest.com/go-game-server/internal/pb"
	"github.xubinbest.com/go-game-server/internal/utils"

	"go.uber.org/zap"
)

func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userName := req.Username
	password := req.Password
	email := req.Email

	if userName == "" || password == "" || email == "" {
		return nil, fmt.Errorf("username, password and email are required")
	}

	_, err := h.dbClient.GetUserByUsername(ctx, userName)
	if err == nil {
		return nil, fmt.Errorf("username already exists")
	}

	salt := auth.GenerateSalt()
	hashedPassword := auth.HashPassword(password, salt)

	userId, err := h.generateUserId()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user id: %v", err)
	}

	user := &models.User{
		ID:           userId,
		Username:     userName,
		PasswordHash: hashedPassword,
		Salt:         salt,
		Email:        email,
		CreatedAt:    time.Now(),
		Role:         "user",
	}
	err = h.dbClient.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return &pb.RegisterResponse{
		UserId: int64(userId),
	}, nil
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	username := req.Username
	password := req.Password

	user, err := h.dbClient.GetUserByUsername(ctx, username)
	if err != nil {
		utils.Error("get user by username err", zap.Error(err))
		return nil, fmt.Errorf("invalid credentials")
	}

	if !auth.VerifyPassword(password, user.PasswordHash, user.Salt) {
		utils.Error("password verify err", zap.Error(err))
		return nil, fmt.Errorf("invalid credentials")
	}

	token, expiresAt, err := auth.GenerateToken(user.ID, user.Role, user.Email, h.cfg.Auth.SecretKey, h.cfg.Auth.TokenExpireTime)
	if err != nil {
		utils.Error("generate token err", zap.Error(err))
		return nil, fmt.Errorf("failed to generate token")
	}

	if err := h.cacheClient.SetToken(ctx, user.ID, token, time.Until(expiresAt)); err != nil {
		utils.Error("set token err", zap.Error(err))
		return nil, fmt.Errorf("failed to store token")
	}

	return &pb.LoginResponse{
		UserId:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

func (h *Handler) generateUserId() (int64, error) {
	id, err := h.sf.NextID()
	if err != nil {
		return 0, fmt.Errorf("failed to generate snowflake ID: %w", err)
	}
	return id, nil
}
