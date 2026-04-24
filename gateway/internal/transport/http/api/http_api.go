package api

import (
	"encoding/json"
	"gateway/internal/service"
	"gateway/internal/transport/http/errmap"
	"net/http"
	"strconv"
)

func ProductCatalogHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		Page uint64 `json:"page"`
		Size uint64 `json:"size"`
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := Request{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, "invalid input data", http.StatusBadRequest)
			return
		}

		products, err := gateway.ProductCatalogPaginated(r.Context(), in.Page, in.Size)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		data, err := json.MarshalIndent(&products, "", "  ")
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}

	return h
}

func ProductsByIDHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		IDList []uint64 `json:"id_list"`
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := &Request{}

		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, "invalid input data", http.StatusBadRequest)
			return
		}

		products, err := gateway.ProductService.ProductsByIDList(r.Context(), in.IDList)
		if err != nil {
			msg, status := errmap.MapError(err)

			http.Error(w, msg, status)
			return
		}

		data, err := json.MarshalIndent(&products, "", "  ")
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}

	return h
}

func CreateUserHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	h := func(w http.ResponseWriter, r *http.Request) {
		in := Request{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, "invalid input data", http.StatusBadRequest)
			return
		}

		userID, err := gateway.CreateNewUser(r.Context(), in.Username, in.Password)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		m := map[string]uint64{"user_id": userID}
		data, err := json.MarshalIndent(&m, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(data)
	}

	return h
}

func DeleteUserHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		UserID uint64 `json:"user_id"`
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := Request{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		err = gateway.UserService.DeleteUser(r.Context(), in.UserID)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	return h
}

func AddMoneyHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		UserID      uint64 `json:"user_id"`
		MoneyAmount uint64 `json:"money_amount"`
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := Request{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = gateway.AddMoney(r.Context(), in.UserID, in.MoneyAmount)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	return h
}

func AddProductToBasketHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		UserID    uint64 `json:"user_id"`
		ProductID uint64 `json:"product_id"`
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := Request{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, "invalid input data", http.StatusBadRequest)
			return
		}

		err = gateway.AddProductToBasket(r.Context(), in.UserID, in.ProductID)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		w.WriteHeader(http.StatusOK)
	}

	return h
}

func CreateOrderHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		UserID uint64 `json:"user_id"`
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := Request{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		orderID, err := gateway.CreateOrder(r.Context(), in.UserID)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		resp := map[string]uint64{
			"order_id": orderID,
		}

		data, err := json.MarshalIndent(&resp, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}

	return h
}

func PayForOrderHandler(gateway *service.Gateway) http.HandlerFunc {
	type Request struct {
		OrderID uint64 `json:"order_id"`
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		in := Request{}
		err := json.NewDecoder(r.Body).Decode(&in)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		paymentID, err := gateway.PayForOrder(r.Context(), in.OrderID)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		m := map[string]uint64{"payment_id": paymentID}
		data, err := json.MarshalIndent(&m, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}

	return h
}

func GetUserNotificationsHandler(gateway *service.Gateway) http.HandlerFunc {
	h := func(w http.ResponseWriter, r *http.Request) {
		userIDValue := r.URL.Query().Get("user_id")
		if userIDValue == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		userID, err := strconv.ParseUint(userIDValue, 10, 64)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}

		limit := uint64(20)
		limitValue := r.URL.Query().Get("limit")
		if limitValue != "" {
			parsedLimit, parseErr := strconv.ParseUint(limitValue, 10, 64)
			if parseErr != nil {
				http.Error(w, "invalid limit", http.StatusBadRequest)
				return
			}
			limit = parsedLimit
		}

		offset := uint64(0)
		offsetValue := r.URL.Query().Get("offset")
		if offsetValue != "" {
			parsedOffset, parseErr := strconv.ParseUint(offsetValue, 10, 64)
			if parseErr != nil {
				http.Error(w, "invalid offset", http.StatusBadRequest)
				return
			}
			offset = parsedOffset
		}

		notifications, err := gateway.NotificationsByUser(r.Context(), userID, limit, offset)
		if err != nil {
			msg, status := errmap.MapError(err)
			http.Error(w, msg, status)
			return
		}

		data, err := json.MarshalIndent(&notifications, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}

	return h
}
