package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"

	"github.com/codeselfstudy/timer-server/internal/protocol"
	"github.com/codeselfstudy/timer-server/internal/room"
)

var (
	jwtSecret    []byte
	roomManager  *room.Manager
	adjectives   = []string{"happy", "swift", "bright", "calm", "bold", "cool", "warm", "soft", "wild", "quiet", "stormy", "sunny", "misty", "dusty", "frosty"}
	nouns        = []string{"tiger", "eagle", "river", "mountain", "forest", "ocean", "meadow", "canyon", "valley", "sunset", "salad", "robot", "cat", "panda", "phoenix"}
)

func main() {
	// Load .env.local for development (silently ignore if not found)
	// Try project root first (when run from monorepo root), then parent paths
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../../.env.local") // When run from packages/timer-server

	// Load JWT secret from environment
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("Warning: JWT_SECRET not set, using random secret (tokens won't persist across restarts)")
		secret = generateRandomString(32)
	}
	jwtSecret = []byte(secret)

	// Initialize room manager
	roomManager = room.NewManager()

	// Set up HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/ws/{roomID}", handleWebSocket)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/rooms", handleRooms) // For debugging

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleaned := roomManager.CleanupEmpty()
			if cleaned > 0 {
				log.Printf("Cleaned up %d empty rooms", cleaned)
			}
		}
	}()

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	log.Printf("Timer server starting on port %s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server error: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomID")
	if roomID == "" {
		http.Error(w, "Room ID required", http.StatusBadRequest)
		return
	}

	// Accept WebSocket connection
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // Configure appropriately for production
	})
	if err != nil {
		log.Printf("Failed to accept WebSocket: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Wait for JOIN message
	_, data, err := conn.Read(ctx)
	if err != nil {
		log.Printf("Failed to read JOIN message: %v", err)
		return
	}

	var joinMsg protocol.Message
	if err := json.Unmarshal(data, &joinMsg); err != nil || joinMsg.Type != protocol.MsgJoin {
		sendError(ctx, conn, protocol.ErrCodeInvalidMessage, "Expected JOIN message")
		return
	}

	// Parse join payload
	token, _ := joinMsg.Payload["token"].(string)
	name, _ := joinMsg.Payload["name"].(string)

	// Determine if user is admin and get their ID
	userID, isAdmin, err := authenticateUser(token, roomID)
	if err != nil {
		log.Printf("Auth error: %v", err)
		// For guests, generate an ID
		userID = generateRandomString(8)
		isAdmin = false
	}

	// Generate name for anonymous users
	if name == "" {
		if isAdmin {
			name = "Admin"
		} else {
			name = generateGuestName(userID)
		}
	}

	// Get or create room
	rm, exists := roomManager.Get(roomID)
	if !exists {
		// Room doesn't exist - only admin can create
		if !isAdmin {
			sendError(ctx, conn, protocol.ErrCodeRoomNotFound, "Room not found")
			return
		}
		// Create room with default settings
		rm = roomManager.Create(roomID, userID, protocol.DefaultSettings())
		log.Printf("Created room %s by user %s", roomID, userID)
	}

	// Create client
	client := &room.Client{
		ID:      userID,
		Name:    name,
		IsAdmin: isAdmin,
		Conn:    conn,
		Send:    make(chan []byte, 256),
	}

	// Register client with room
	rm.Register(client)
	defer rm.Unregister(client)

	log.Printf("User %s (%s) joined room %s (admin: %v)", userID, name, roomID, isAdmin)

	// Start write loop
	go room.WriteLoop(ctx, client)

	// Read loop
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				log.Printf("User %s left room %s normally", userID, roomID)
			} else {
				log.Printf("User %s disconnected from room %s: %v", userID, roomID, err)
			}
			return
		}

		rm.HandleMessage(client, data)
	}
}

func authenticateUser(token, roomID string) (userID string, isAdmin bool, err error) {
	if token == "" {
		return "", false, fmt.Errorf("no token provided")
	}

	// Parse JWT
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return "", false, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return "", false, fmt.Errorf("invalid claims")
	}

	// Verify room ID matches
	tokenRoomID, _ := claims["roomId"].(string)
	if tokenRoomID != roomID {
		return "", false, fmt.Errorf("room ID mismatch")
	}

	userID, _ = claims["sub"].(string)
	isAdmin, _ = claims["isAdmin"].(bool)

	return userID, isAdmin, nil
}

func generateGuestName(seed string) string {
	// Use seed to deterministically pick adjective and noun
	adj := adjectives[hashString(seed)%len(adjectives)]
	noun := nouns[hashString(seed+"noun")%len(nouns)]
	num := hashString(seed+"num") % 10000

	return fmt.Sprintf("%s_%s_%04d", adj, noun, num)
}

func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h
}

func generateRandomString(length int) string {
	bytes := make([]byte, length/2+1)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func sendError(ctx context.Context, conn *websocket.Conn, code, message string) {
	msg := protocol.NewMessage(protocol.MsgError, map[string]any{
		"code":    code,
		"message": message,
	})
	data, _ := json.Marshal(msg)
	conn.Write(ctx, websocket.MessageText, data)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "ok",
		"rooms":     roomManager.Count(),
		"timestamp": time.Now().Unix(),
	})
}

func handleRooms(w http.ResponseWriter, r *http.Request) {
	// Only allow in development
	if os.Getenv("NODE_ENV") == "production" {
		http.Error(w, "Not available in production", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	rooms := roomManager.List()
	roomInfo := make([]map[string]any, 0, len(rooms))

	for _, id := range rooms {
		if rm, ok := roomManager.Get(id); ok {
			state := rm.TimerState()
			roomInfo = append(roomInfo, map[string]any{
				"id":          id,
				"clients":     rm.ClientCount(),
				"phase":       state.Phase,
				"isRunning":   state.IsRunning,
				"remainingMs": state.RemainingMs,
			})
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"rooms": roomInfo,
		"count": len(rooms),
	})
}

// GetJWTSecret returns the JWT secret (for use in generating tokens from the Node app)
func GetJWTSecret() []byte {
	return jwtSecret
}

// Helper to check if a string contains only valid room ID characters
func isValidRoomID(id string) bool {
	if len(id) < 3 || len(id) > 50 {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// Helper for case-insensitive string comparison
func equalFold(a, b string) bool {
	return strings.EqualFold(a, b)
}
