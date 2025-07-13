package controllers

import (
	"github.com/julietteengel/salesrep-api/pkg/auth"
	"github.com/julietteengel/salesrep-api/pkg/conversation"
	"go.uber.org/fx"
)

var PrivateControllers = fx.Options(
	fx.Provide(
		// Auth dependencies
		auth.NewAuthRepository,
		auth.NewAuthService,
		
		// Conversation dependencies
		conversation.NewConversationRepository,
		conversation.NewConversationService,
		AsController(conversation.NewConversationController),
	),
)
