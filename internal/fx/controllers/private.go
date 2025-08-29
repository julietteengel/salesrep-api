package controllers

import (
	"github.com/julietteengel/salesrep-api/pkg/auth"
	"github.com/julietteengel/salesrep-api/pkg/call"
	"github.com/julietteengel/salesrep-api/pkg/conversation"
	"github.com/julietteengel/salesrep-api/pkg/insights"
	"github.com/julietteengel/salesrep-api/pkg/storage"
	"github.com/julietteengel/salesrep-api/pkg/transcription"
	"github.com/julietteengel/salesrep-api/pkg/user"
	"go.uber.org/fx"
)

var PrivateControllers = fx.Options(
	fx.Provide(
		// Storage dependencies
		storage.NewS3Service,
		
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
		
		// Call dependencies
		call.NewCallRepository,
		call.NewCallService,
		AsController(call.NewCallController),
		
		// Transcription dependencies
		transcription.NewTranscriptionService,
		AsController(transcription.NewTranscriptionController),
		
		// Insights dependencies
		insights.NewInsightsService,
	),
)
