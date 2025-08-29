package controllers

import (
	"github.com/julietteengel/salesrep-api/pkg/auth"
	"github.com/julietteengel/salesrep-api/pkg/conversation"
	"github.com/julietteengel/salesrep-api/pkg/user"
	"go.uber.org/fx"
)

var PrivateControllers = fx.Options(
	fx.Provide(
		// Auth dependencies
		auth.NewAuthRepository,
		auth.NewAuthService,
		
		// User dependencies
		user.NewUserRepository,
		user.NewUserService,
		AsController(user.NewUserController),
		
		// Conversation dependencies
		conversation.NewConversationRepository,
		conversation.NewConversationService,
		AsController(conversation.NewConversationController),
	),
)
