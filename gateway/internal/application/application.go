package application

import (
	ns "gateway/internal/clients/notification_service"
	os "gateway/internal/clients/order_service"
	ps "gateway/internal/clients/product_service"
	us "gateway/internal/clients/user_service"
	"gateway/internal/config"
	"gateway/internal/service"
	"gateway/internal/transport/http/api"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Cfg    interface{}
	Server *http.Server
}

func New(cfgPath string) (*App, error) {
	// loading service configuration
	appCfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	// creating router
	r := chi.NewRouter()

	// creating clients
	productService, err := ps.NewProductService(appCfg.ProductServiceAddr)
	if err != nil {
		return nil, err
	}

	userService, err := us.NewUserService(appCfg.UserServiceAddr)
	if err != nil {
		return nil, err
	}

	orderService, err := os.NewOrderService(appCfg.OrderServiceAddr)
	if err != nil {
		return nil, err
	}

	notificationService, err := ns.NewNotificationService(appCfg.NotificationServiceAddr)
	if err != nil {
		return nil, err
	}

	// creating service
	gateway := service.New(productService, userService, orderService, notificationService)

	// register routes
	r.Get("/api/v1/products-catalog", api.ProductCatalogHandler(gateway))
	r.Get("/api/v1/products", api.ProductsByIDHandler(gateway))

	r.Post("/api/v1/users", api.CreateUserHandler(gateway))
	r.Delete("/api/v1/users", api.DeleteUserHandler(gateway))

	r.Post("/api/v1/users/basket", api.AddProductToBasketHandler(gateway))

	r.Post("/api/v1/users/wallets", api.AddMoneyHandler(gateway))

	r.Post("/api/v1/orders", api.CreateOrderHandler(gateway))

	r.Post("/api/v1/payments", api.PayForOrderHandler(gateway))
	r.Get("/api/v1/notifications", api.GetUserNotificationsHandler(gateway))

	// creating http server
	s := &http.Server{
		Addr:    appCfg.HTTPAddress,
		Handler: r,
	}

	app := &App{
		Cfg:    appCfg,
		Server: s,
	}

	return app, nil
}
