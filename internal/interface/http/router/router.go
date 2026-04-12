package router

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"room-booking/internal/interface/http/handler"
	"room-booking/internal/interface/http/middleware"
	"room-booking/internal/service"
)

func NewRouter(
	jwtService *service.JWTService,
	roomService service.RoomService,
	scheduleService handler.ScheduleCreator,
	slotService handler.SlotLister,
	bookingService handler.BookingCreator,
	logger handler.Logger,
) *http.ServeMux {
	mux := http.NewServeMux()

	authHandler := handler.NewAuthHandler(jwtService)
	roomHandler := handler.NewRoomHandler(roomService, logger)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	slotHandler := handler.NewSlotListHandler(slotService)
	bookingHandler := handler.NewBookingCreateHandler(bookingService)

	registerBasicRoutes(mux)
	registerAuthRoutes(mux, authHandler)
	registerRoomAndScheduleRoutes(mux, roomHandler, scheduleHandler, slotHandler, jwtService)
	registerBookingRoutes(mux, bookingHandler, jwtService)
	return mux
}

func registerAuthRoutes(mux *http.ServeMux, h *handler.AuthHandler) {
	mux.HandleFunc("/dummyLogin", h.DummyLogin)
}

func registerBasicRoutes(mux *http.ServeMux) {
	mux.Handle("/health", handler.NewHealthHandler())
	mux.Handle("/_info", handler.NewInfoHandler())
}

func registerRoomAndScheduleRoutes(
	mux *http.ServeMux,
	roomHandler *handler.RoomHandler,
	scheduleHandler *handler.ScheduleCreateHandler,
	slotHandler *handler.SlotListHandler,
	jwtService *service.JWTService,
) {
	auth := middleware.AuthMiddleware(jwtService)
	adminOnly := middleware.RequireRole("admin")
	userOrAdmin := middleware.RequireAnyRole("user", "admin")

	mux.Handle("/rooms/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		switch {
		case path == "/rooms/list" && method == http.MethodGet:
			auth(userOrAdmin(http.HandlerFunc(roomHandler.ListRooms))).ServeHTTP(w, r)

		case path == "/rooms/create" && method == http.MethodPost:
			auth(adminOnly(http.HandlerFunc(roomHandler.CreateRoom))).ServeHTTP(w, r)

		case method == http.MethodPost && strings.HasSuffix(path, "/schedule/create"):
			parts := strings.Split(strings.Trim(path, "/"), "/")
			if len(parts) != 4 || parts[0] != "rooms" || parts[2] != "schedule" || parts[3] != "create" {
				respondError(w, http.StatusBadRequest, "InvalidRequest", "invalid path format")
				return
			}

			roomId := parts[1]
			if roomId == "" {
				respondError(w, http.StatusBadRequest, "InvalidRequest", "roomId is required in path")
				return
			}

			ctx := context.WithValue(r.Context(), handler.CtxRoomID, roomId)
			reqWithCtx := r.WithContext(ctx)

			wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				scheduleHandler.ServeHTTP(w, reqWithCtx)
			})
			auth(adminOnly(wrappedHandler)).ServeHTTP(w, r)

		case method == http.MethodGet && strings.HasSuffix(path, "/slots/list"):
			parts := strings.Split(strings.Trim(path, "/"), "/")
			if len(parts) != 4 || parts[0] != "rooms" || parts[2] != "slots" || parts[3] != "list" {
				respondError(w, http.StatusBadRequest, "InvalidRequest", "invalid path format")
				return
			}
			roomId := parts[1]
			if roomId == "" {
				respondError(w, http.StatusBadRequest, "InvalidRequest", "roomId is required in path")
				return
			}
			ctx := context.WithValue(r.Context(), handler.CtxRoomID, roomId)
			reqWithCtx := r.WithContext(ctx)

			wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				slotHandler.ServeHTTP(w, reqWithCtx)
			})
			auth(userOrAdmin(wrappedHandler)).ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

func registerBookingRoutes(
	mux *http.ServeMux,
	bookingHandler *handler.BookingCreateHandler,
	jwtService *service.JWTService,
) {
	auth := middleware.AuthMiddleware(jwtService)
	userOnly := middleware.RequireRole("user")

	mux.Handle("/bookings/create", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		auth(userOnly(bookingHandler)).ServeHTTP(w, r)
	}))
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
