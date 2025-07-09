package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/julietteengel/salesrep-api/pkg/salesrep"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func init() {
	err := godotenv.Load("/app/cmd/server/.env")
	if err != nil {
		log.Println("⚠️ Could not load .env file")
	}
}

func NewGormDB() *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		viper.GetString("DB_HOST"),
		viper.GetString("DB_PORT"),
		viper.GetString("DB_USER"),
		viper.GetString("DB_DATABASE"),
		viper.GetString("DB_PASSWORD"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Printf("Could not connect to database: %v", err)
		return nil
	}

	log.Println("Connected to database")

	if err := db.AutoMigrate(&salesrep.SalesRep{}); err != nil {
		log.Printf("Migration failed: %v", err)
		return nil
	}

	a := 5
	fmt.Printf("Location of a: %p", &a)

	log.Println("Database migrated successfully")
	return db
}

//func DatabaseProvider(lc fx.Lifecycle) (*sqlx.DB, error) {
//	url, ok := os.LookupEnv("DATABASE_URL")
//	if !ok {
//		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
//	}
//
//	var err error
//	db, err := sqlx.Connect("postgres", url)
//	if err != nil {
//		return nil, err
//	}
//
//	lc.Append(fx.Hook{
//		OnStop: func(ctx context.Context) error {
//			return db.Close()
//		},
//	})
//
//	return db, nil
//}
