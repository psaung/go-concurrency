package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/psaung/go-concurrency/models"
	"github.com/psaung/go-concurrency/repo"
)

type handler struct {
	repo repo.Repo
	once sync.Once
}

type Handler interface {
	Index(w http.ResponseWriter, r *http.Request)
	ProductIndex(w http.ResponseWriter, r *http.Request)
	OrderShow(w http.ResponseWriter, r *http.Request)
	OrderInsert(w http.ResponseWriter, r *http.Request)
	Close(w http.ResponseWriter, r *http.Request)
	Stats(w http.ResponseWriter, r *http.Request)
	OrderReverse(w http.ResponseWriter, r *http.Request)
}

func New() (Handler, error) {
	r, err := repo.New()
	if err != nil {
		return nil, err
	}
	h := handler{
		repo: r,
	}
	return &h, nil
}

func (h *handler) Index(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, "Welcome to the Orders App!", nil)
}

func (h *handler) ProductIndex(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, h.repo.GetAllProducts(), nil)
}

func (h *handler) OrderShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId := vars["orderId"]
	o, err := h.repo.GetOrder(orderId)
	if err != nil {
		writeResponse(w, http.StatusNotFound, nil, err)
		return
	}
	writeResponse(w, http.StatusOK, o, nil)
}

func (h *handler) OrderInsert(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeResponse(w, http.StatusBadRequest, nil, fmt.Errorf("invalid order body:%v", err))
		return
	}
	order, err := h.repo.CreateOrder(item)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, nil, err)
		return
	}
	writeResponse(w, http.StatusOK, order, nil)
}

func (h *handler) Close(w http.ResponseWriter, r *http.Request) {
	h.invokeClose()
	writeResponse(w, http.StatusOK, "The Orders App is now closed!", nil)
}

func (h *handler) Stats(w http.ResponseWriter, r *http.Request) {
	reqCtx := r.Context()
	ctx, cancel := context.WithTimeout(reqCtx, 600*time.Millisecond)
	defer cancel()
	stats, err := h.repo.GetOrderStats(ctx)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, nil, err)
		return
	}
	writeResponse(w, http.StatusOK, stats, nil)
}

func (h *handler) OrderReverse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderId := vars["orderId"]
	order, err := h.repo.RequestReversal(orderId)
	if err != nil {
		writeResponse(w, http.StatusInternalServerError, nil, err)
		return
	}
	writeResponse(w, http.StatusOK, order, nil)
}

func (h *handler) invokeClose() {
	h.once.Do(func() {
		h.repo.Close()
	})
}
