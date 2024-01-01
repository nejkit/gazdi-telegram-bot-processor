package services

import (
	"context"
	"gazdi-telegram-bot-processor/providers"
)

var (
	AdminUsersSet = "admins"
)

type UserService struct {
	redisClient providers.RedisProvider
}

func NewUserService(redisClient providers.RedisProvider) UserService {
	return UserService{redisClient: redisClient}
}

func (s *UserService) CheckUserAdminRole(userId int) (bool, error) {
	return s.redisClient.CheckMemberOfSet(context.TODO(), AdminUsersSet, userId)
}
