package main

import (
	"github.com/Ruclo/MyMeals/internal/apperrors"
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/database"
	"github.com/Ruclo/MyMeals/internal/events"
	"github.com/Ruclo/MyMeals/internal/handlers"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/Ruclo/MyMeals/internal/services"
	"github.com/Ruclo/MyMeals/internal/storage"
	cloudinary2 "github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	config.InitConfig()

	db := database.InitDB()
	sseServer := events.NewSSEServer()
	cloudinary, err := cloudinary2.NewFromURL(config.ConfigInstance.CloudinaryUrl())
	if err != nil {
		log.Fatal(err)
	}

	orderBroadcaster := sseServer.NewBroadcaster()
	imageStorage := storage.NewCloudinaryStorage(cloudinary)

	mealRepo := repositories.NewMealRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	userRepo := repositories.NewUserRepository(db)

	userService := services.NewUserService(userRepo)
	mealService := services.NewMealService(mealRepo, imageStorage)
	orderService := services.NewOrderService(orderRepo, mealRepo, imageStorage, orderBroadcaster)

	mealsHandler := handlers.NewMealsHandler(mealService)
	ordersHandler := handlers.NewOrdersHandler(orderService)
	usersHandler := handlers.NewUsersHandler(userService)

	regularStaff := models.User{
		Username: "regular",
		Password: "password",
		Role:     models.RegularStaffRole,
	}
	admin := models.User{
		Username: "admin",
		Password: "password",
		Role:     models.AdminRole,
	}
	userService.Create(&regularStaff)
	userService.Create(&admin)

	r := gin.Default()
	r.Use(apperrors.ErrorHandler())

	// Public routes
	r.GET("/api/meals", mealsHandler.GetMeals())
	r.POST("/api/login", usersHandler.Login())
	r.POST("/api/orders", ordersHandler.PostOrder())

	authorized := r.Group("/api")
	authorized.Use(auth.AuthMiddleware())
	authorized.GET("/me", usersHandler.GetMe())
	authorized.GET("/orders/me", ordersHandler.GetMyOrder())

	// AdminRole or RegularStaffRole routes
	staffRoutes := authorized.Group("/")
	staffRoutes.Use(auth.RequireAnyRole(models.RegularStaffRole, models.AdminRole))
	{

		staffRoutes.GET("/orders/pending", ordersHandler.GetPendingOrders())
		staffRoutes.GET("/events/orders", sseServer.Handler()...)
		staffRoutes.PUT("/account/password", usersHandler.ChangePassword())
		staffRoutes.POST("/orders/:orderID/items/:mealID/status", ordersHandler.UpdateStatus())
	}

	// AdminRole only access
	adminRoutes := authorized.Group("/")
	adminRoutes.Use(auth.RequireAnyRole(models.AdminRole))
	{
		adminRoutes.GET("/meals/deleted", mealsHandler.GetMealsWithDeleted())
		adminRoutes.POST("/meals", mealsHandler.PostMeal())
		adminRoutes.POST("/meals/:mealID/replace", mealsHandler.PostMealReplace())
		adminRoutes.DELETE("/meals/:mealID", mealsHandler.DeleteMeal())
		adminRoutes.POST("/users", usersHandler.PostUser())
		adminRoutes.GET("/orders", ordersHandler.GetOrders())
		adminRoutes.GET("/users/staff", usersHandler.GetStaff())
		adminRoutes.DELETE("/users/:username", usersHandler.DeleteUser())
	}

	// Order Creator access only
	orderRoutes := authorized.Group("/orders/:orderID")
	orderRoutes.Use(auth.RequireOrderAccess())
	{
		orderRoutes.POST("/items", ordersHandler.PostOrderItems())
		orderRoutes.POST("/review", ordersHandler.PostOrderReview())
	}

	r.Run()
}
