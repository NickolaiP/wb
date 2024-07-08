package handler

import (
	"html/template"
	"log"
	"net/http"
	"test/internal/storage/cache"
)

// структура обработчика
type Handler struct {
	Cache *cache.OrdersCache
}

// создание нового обработчика
func NewHandler(inMemory *cache.OrdersCache) Handler {
	return Handler{
		Cache: inMemory,
	}
}

// обработчик возвращающий заказ по id
func (h *Handler) GetByID(w http.ResponseWriter, req *http.Request) {
	log.Println("New get by id request")
	// доступен только get метод
	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid method. Only GET method is supported"))
		return
	}
	// получаем id
	data, exists := h.Cache.Get(req.PathValue("id"))
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Order not found"))
		return
	}
	tmpl, err := template.ParseFiles("/home/mandom/go-test/internal/handler/template.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "orders", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
