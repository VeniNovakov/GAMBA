package ticket

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RegisterRoutes(r chi.Router) {
	r.Get("/tickets", c.GetAll)
	r.Post("/tickets", c.Create)
	r.Get("/tickets/{id}", c.GetByID)
	r.Put("/tickets/{id}", c.Update)
	r.Post("/tickets/{id}/close", c.Close)
	r.Post("/tickets/{id}/messages", c.AddMessage)
}

func (c *Controller) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, isAdmin := getUserFromContext(r)

	tickets, err := c.service.GetAll(userID, isAdmin)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, tickets)
}

func (c *Controller) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	userID, isAdmin := getUserFromContext(r)

	ticket, err := c.service.GetByID(id, userID, isAdmin)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserFromContext(r)

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.service.Create(userID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, ticket)
}

func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	userID, isAdmin := getUserFromContext(r)

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	ticket, err := c.service.Update(id, userID, isAdmin, &req)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ticket)
}

func (c *Controller) Close(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	userID, isAdmin := getUserFromContext(r)

	if err := c.service.Close(id, userID, isAdmin); err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ticket closed"})
}

func (c *Controller) AddMessage(w http.ResponseWriter, r *http.Request) {
	ticketID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	userID, isAdmin := getUserFromContext(r)

	var req AddMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	message, err := c.service.AddMessage(ticketID, userID, isAdmin, &req)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, message)
}

// getUserFromContext extracts user info from request context (set by auth middleware)
func getUserFromContext(r *http.Request) (uuid.UUID, bool) {
	userID, _ := r.Context().Value("user_id").(uuid.UUID)
	role, _ := r.Context().Value("role").(string)
	return userID, role == "administrator"
}

func handleError(w http.ResponseWriter, err error) {
	switch err {
	case ErrTicketNotFound:
		writeError(w, http.StatusNotFound, err.Error())
	case ErrUnauthorized:
		writeError(w, http.StatusForbidden, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
