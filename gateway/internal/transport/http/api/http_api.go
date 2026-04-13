package api

import (
	"encoding/json"
	"errors"
	"gateway/internal/apperrors"
	"gateway/internal/service"
	"net/http"
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
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		products, err := gateway.ProductCatalogPaginated(r.Context(), in.Page, in.Size)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				http.Error(w, "no data", http.StatusNoContent)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.MarshalIndent(&products, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		products, err := gateway.ProductService.ProductsByIDList(r.Context(), in.IDList)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				http.Error(w, "no data", http.StatusNoContent)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.MarshalIndent(&products, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		userID, err := gateway.CreateNewUser(r.Context(), in.Username, in.Password)
		if err != nil {
			if errors.Is(err, apperrors.ErrAlreadyExists) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			if errors.Is(err, apperrors.ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			if errors.Is(err, apperrors.ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = gateway.AddProductToBasket(r.Context(), in.UserID, in.ProductID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			if errors.Is(err, apperrors.ErrFailedPrecondition) {
				http.Error(w, "unable to create order. basket is empty", http.StatusBadRequest)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
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
