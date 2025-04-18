package main

import (
	"github.com/Ruclo/MyMeals/internal/config"
	"github.com/Ruclo/MyMeals/internal/database"
	"github.com/Ruclo/MyMeals/internal/errors"
	"github.com/Ruclo/MyMeals/internal/events"
	"github.com/Ruclo/MyMeals/internal/handlers"
	"github.com/Ruclo/MyMeals/internal/repositories"
	cloudinary2 "github.com/cloudinary/cloudinary-go/v2"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	config.InitConfig()

	db := database.InitDB()
	orderEvents := events.NewServer()
	cloudinary, err := cloudinary2.NewFromURL(config.ConfigInstance.CloudinaryUrl())
	if err != nil {
		log.Fatal(err)
	}

	mealRepo := repositories.NewMealRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	userRepo := repositories.NewUserRepository(db)
	broadcastingOrderRepo := repositories.NewBroadcastingOrderRepository(orderRepo, orderEvents.Message)

	mealsHandler := handlers.NewMealsHandler(mealRepo, cloudinary)
	ordersHandler := handlers.NewOrdersHandler(broadcastingOrderRepo, cloudinary)
	usersHandler := handlers.NewUsersHandler(userRepo)

	r := gin.Default()
	r.Use(errors.ErrorHandler())
	// Public routes
	r.LoadHTMLGlob("templates/*")

	// Route to serve the HTML file
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "MyMeals",
		})
	})
	r.GET("/api/meals", mealsHandler.GetMeals())
	r.POST("/api/login", usersHandler.Login())
	r.POST("/api/orders", ordersHandler.PostOrder())

	authorized := r.Group("/api")
	//authorized.Use(auth.AuthMiddleware())

	// AdminRole or RegularStaffRole routes
	staffRoutes := authorized.Group("/")
	//staffRoutes.Use(auth.RequireAnyRole(models.RegularStaffRole, models.AdminRole))
	{
		staffRoutes.GET("/orders/pending", ordersHandler.GetPendingOrders())
		staffRoutes.GET("/api/events/orders", orderEvents.Handler()...)
		staffRoutes.PUT("/account/password", usersHandler.ChangePassword())
		staffRoutes.PUT("/orders/:orderID/items/:mealID/status", ordersHandler.UpdateStatus())
	}

	// AdminRole only access
	adminRoutes := authorized.Group("/")
	//adminRoutes.Use(auth.RequireAnyRole(models.AdminRole))
	{
		adminRoutes.POST("/meals", mealsHandler.PostMeal())
		adminRoutes.PUT("/meals/:mealID", mealsHandler.PutMeal())
		adminRoutes.DELETE("/meals/:mealID", mealsHandler.DeleteMeal())
		adminRoutes.POST("/users", usersHandler.PostUser())
		adminRoutes.GET("/orders", ordersHandler.GetOrders())
	}

	// Order Creator access only
	orderRoutes := authorized.Group("/orders/:orderID")
	//orderRoutes.Use(auth.RequireOrderAccess())
	{
		orderRoutes.POST("/items", ordersHandler.PostOrderItem())
		orderRoutes.POST("/review", ordersHandler.PostOrderReview())
	}

	r.Run()
}
