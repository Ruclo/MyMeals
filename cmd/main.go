package main

import (
	"github.com/Ruclo/MyMeals/internal/auth"
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/database"
	"github.com/Ruclo/MyMeals/internal/handlers"
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/Ruclo/MyMeals/internal/repositories"
	"github.com/gin-gonic/gin"
)

//TODO: sse, update status endpoint

func main() {
	config.InitConfig()

	db := database.InitDB()

	r := gin.Default()

	mealRepo := repositories.NewMealRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	userRepo := repositories.NewUserRepository(db)

	mealsHandler := handlers.NewMealsHandler(mealRepo)
	ordersHandler := handlers.NewOrdersHandler(orderRepo)
	usersHandler := handlers.NewUsersHandler(userRepo)

	// Public routes
	r.GET("/api/meals", mealsHandler.GetMeals())
	r.POST("/api/login", usersHandler.Login())
	r.POST("/api/orders", ordersHandler.PostOrder())

	authorized := r.Group("/api")
	authorized.Use(auth.AuthMiddleware())

	// AdminRole or RegularStaffRole routes
	staffRoutes := authorized.Group("/")
	staffRoutes.Use(auth.RequireAnyRole(models.RegularStaffRole, models.AdminRole))
	{
		staffRoutes.GET("/orders/pending", ordersHandler.GetPendingOrders())
		staffRoutes.PUT("/account/password", usersHandler.ChangePassword())
	}

	// AdminRole only access
	adminRoutes := authorized.Group("/")
	adminRoutes.Use(auth.RequireAnyRole(models.AdminRole))
	{
		adminRoutes.POST("/meals", mealsHandler.PostMeal())
		adminRoutes.PUT("/meals/:id", mealsHandler.PutMeal())
		adminRoutes.DELETE("/meals/:id", mealsHandler.DeleteMeal())
		adminRoutes.POST("/users", usersHandler.PostUser())
	}

	// Order Creator access only
	orderRoutes := authorized.Group("/orders/:orderID")
	orderRoutes.Use(auth.RequireOrderAccess())
	{
		orderRoutes.POST("/items", ordersHandler.PostOrderItem())
		orderRoutes.POST("/review", ordersHandler.PostOrderReview())
	}

	r.Run()
}
