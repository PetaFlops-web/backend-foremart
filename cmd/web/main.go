package main

import (
	"fmt"

	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/auth"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/product"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/store"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/modules/transaction"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/config"
	"github.com/PetaFlops-web/backend-shop-smbk/internal/shared/middleware"
	module "github.com/PetaFlops-web/backend-shop-smbk/internal/shared/modules"

	_ "github.com/PetaFlops-web/backend-shop-smbk/docs"
)

// @title           AIC Backend API
// @version         1.0
// @description     API Documentation for AIC Backend.
// @termsOfService  http://swagger.io/terms/
// @contact.name    Technical Support
// @contact.url     http://www.swagger.io/support
// @contact.email   [EMAIL_ADDRESS]
// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	app := config.NewFiber(viperConfig)

	// Auth Middleware
	authMiddleware := middleware.AuthMiddleware(viperConfig)

	// Module initialization (ordered by dependency)
	authModule := auth.New(db, log, validate, viperConfig)
	storeModule := store.New(db, validate, log)
	productModule := product.New(db, log, validate, viperConfig)
	transactionModule := transaction.New(db, log, validate, viperConfig, productModule.Client())

	// Register all modules
	modules := []module.Module{
		authModule,
		storeModule,
		productModule,
		transactionModule,
	}

	// Auto-migration (each module migrates its own tables)
	for _, m := range modules {
		if err := m.Migrate(); err != nil {
			log.Fatalf("Failed to migrate: %v", err)
		}
	}

	// Route registration
	api := app.Group("/api")
	for _, m := range modules {
		m.RegisterRoutes(api, authMiddleware)
	}

	// Start server
	port := viperConfig.GetInt("web.port")
	if port == 0 {
		port = 8080
	}

	log.Infof("Server is starting on port :%d", port)

	err := app.Listen(fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}