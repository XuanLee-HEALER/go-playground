package ws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Hub coordinates registered clients and message broadcasts.
type Hub struct {
	register       chan *Client
	unregister     chan *Client
	broadcast      chan broadcastMessage
	clientsByGroup map[string]map[*Client]struct{}
	clientsByUser  map[string]map[*Client]struct{}
	instanceID     string
	redis          *redis.Client
	redisChannel   string
	userSubs       map[string]*userSubscription
}

// NewHub constructs a hub with initialized channels.
func NewHub() *Hub {
	hub := &Hub{
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		broadcast:      make(chan broadcastMessage, 128),
		clientsByGroup: make(map[string]map[*Client]struct{}),
		clientsByUser:  make(map[string]map[*Client]struct{}),
		instanceID:     newInstanceID(),
		userSubs:       make(map[string]*userSubscription),
	}

	// Redis is optional; default to localhost for local multi-node testing.
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	hub.redis = redis.NewClient(&redis.Options{Addr: redisAddr})
	hub.redisChannel = "ws:broadcast"
	if channel := os.Getenv("REDIS_CHANNEL"); channel != "" {
		hub.redisChannel = channel
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := hub.redis.Ping(ctx).Err(); err != nil {
		log.Printf("redis unavailable at %s: %v", redisAddr, err)
		hub.redis = nil
		return hub
	}

	hub.startRedisSubscriber()
	return hub
}

// Run processes all hub events in a single goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Track by group and by user id for scoped broadcasts.
			if h.clientsByGroup[client.group] == nil {
				h.clientsByGroup[client.group] = make(map[*Client]struct{})
			}
			h.clientsByGroup[client.group][client] = struct{}{}
			if h.clientsByUser[client.id] == nil {
				h.clientsByUser[client.id] = make(map[*Client]struct{})
			}
			h.clientsByUser[client.id][client] = struct{}{}
			h.ensureUserSubscription(client.id)
		case client := <-h.unregister:
			group := h.clientsByGroup[client.group]
			if _, ok := group[client]; ok {
				delete(group, client)
				close(client.send)
				if len(group) == 0 {
					delete(h.clientsByGroup, client.group)
				}
			}
			user := h.clientsByUser[client.id]
			if _, ok := user[client]; ok {
				delete(user, client)
				if len(user) == 0 {
					delete(h.clientsByUser, client.id)
				}
			}
			h.releaseUserSubscription(client.id)
		case msg := <-h.broadcast:
			// Redis fan-out only happens for local messages.
			if h.redis != nil && msg.source == h.instanceID {
				h.publishRedis(msg)
			}
			h.fanout(msg)
		}
	}
}

// SendToUser broadcasts a payload to all connections for the given user id.
func (h *Hub) SendToUser(userID string, payload []byte) {
	if userID == "" {
		return
	}
	h.broadcast <- broadcastMessage{
		userID:  userID,
		payload: payload,
		source:  h.instanceID,
	}
}

// Client is a single websocket connection.
type Client struct {
	conn  *websocket.Conn
	send  chan []byte
	id    string
	group string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket upgrades the HTTP request and registers the client.
func HandleWebSocket(w http.ResponseWriter, r *http.Request, hub *Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Read the requested id + group to isolate sessions.
	id := r.URL.Query().Get("id")
	if id == "" {
		id = "default"
	}
	group := r.URL.Query().Get("group")
	if group == "" {
		group = "default"
	}

	client := &Client{conn: conn, send: make(chan []byte, 64), id: id, group: group}
	hub.register <- client

	go client.writePump()
	client.readPump(hub)
}

// readPump reads messages from the websocket and forwards them to the hub.
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(1 << 20)
	_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return
			}
			return
		}
		// Preserve echo semantics, then publish to redis for other nodes.
		hub.broadcast <- broadcastMessage{group: c.group, userID: c.id, payload: msg, source: hub.instanceID}
	}
}

// writePump sends messages from the hub to the websocket.
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// broadcastMessage keeps payloads scoped for group and user broadcasts.
type broadcastMessage struct {
	group   string
	userID  string
	payload []byte
	source  string
}

// redisEnvelope is the cross-node wire format carried by Redis pub/sub.
type redisEnvelope struct {
	Group   string `json:"group"`
	UserID  string `json:"user_id"`
	Payload string `json:"payload"`
	Source  string `json:"source"`
}

type userSubscription struct {
	sub    *redis.PubSub
	cancel context.CancelFunc
	count  int
}

func (h *Hub) fanout(msg broadcastMessage) {
	sent := make(map[*Client]struct{})
	for client := range h.clientsByGroup[msg.group] {
		if h.trySend(client, msg.payload) {
			sent[client] = struct{}{}
		}
	}
	for client := range h.clientsByUser[msg.userID] {
		if _, ok := sent[client]; ok {
			continue
		}
		h.trySend(client, msg.payload)
	}
}

// trySend attempts to enqueue a payload and cleans up stalled clients.
func (h *Hub) trySend(client *Client, payload []byte) bool {
	select {
	case client.send <- payload:
		return true
	default:
		close(client.send)
		if group := h.clientsByGroup[client.group]; group != nil {
			delete(group, client)
			if len(group) == 0 {
				delete(h.clientsByGroup, client.group)
			}
		}
		if user := h.clientsByUser[client.id]; user != nil {
			delete(user, client)
			if len(user) == 0 {
				delete(h.clientsByUser, client.id)
			}
		}
		return false
	}
}

// startRedisSubscriber fans Redis broadcasts back into the local hub.
func (h *Hub) startRedisSubscriber() {
	ctx := context.Background()
	sub := h.redis.Subscribe(ctx, h.redisChannel)
	ch := sub.Channel()

	go func() {
		for msg := range ch {
			var env redisEnvelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
				continue
			}
			if env.Source == h.instanceID {
				continue
			}
			payload, err := base64.StdEncoding.DecodeString(env.Payload)
			if err != nil {
				continue
			}
			h.broadcast <- broadcastMessage{
				group:   env.Group,
				userID:  env.UserID,
				payload: payload,
				source:  env.Source,
			}
		}
	}()
}

// ensureUserSubscription registers a per-user Redis subscription once.
func (h *Hub) ensureUserSubscription(userID string) {
	if h.redis == nil || userID == "" {
		return
	}
	sub, ok := h.userSubs[userID]
	if ok {
		sub.count++
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	pubsub := h.redis.Subscribe(ctx, userChannel(userID))
	entry := &userSubscription{cancel: cancel, sub: pubsub, count: 1}
	h.userSubs[userID] = entry

	ch := pubsub.Channel()
	go func() {
		for msg := range ch {
			// Prefix so the frontend log can identify redis messages.
			payload := []byte("[redis] " + msg.Payload)
			log.Printf("redis user message user=%s payload=%s", userID, msg.Payload)
			h.broadcast <- broadcastMessage{
				userID:  userID,
				payload: payload,
				source:  "redis-user",
			}
		}
	}()
}

// releaseUserSubscription decrements and closes the user subscription when unused.
func (h *Hub) releaseUserSubscription(userID string) {
	if h.redis == nil || userID == "" {
		return
	}
	sub, ok := h.userSubs[userID]
	if !ok {
		return
	}
	sub.count--
	if sub.count > 0 {
		return
	}
	sub.cancel()
	_ = sub.sub.Close()
	delete(h.userSubs, userID)
}

// PublishToUsers sends a payload to specific user channels via Redis.
func (h *Hub) PublishToUsers(userIDs []string, payload []byte) {
	if h.redis == nil {
		return
	}
	for _, userID := range userIDs {
		if userID == "" {
			continue
		}
		if err := h.redis.Publish(context.Background(), userChannel(userID), payload).Err(); err != nil {
			log.Printf("redis publish user=%s failed: %v", userID, err)
		}
	}
}

// publishRedis publishes the message to other nodes via Redis.
func (h *Hub) publishRedis(msg broadcastMessage) {
	payload := base64.StdEncoding.EncodeToString(msg.payload)
	env := redisEnvelope{
		Group:   msg.group,
		UserID:  msg.userID,
		Payload: payload,
		Source:  msg.source,
	}
	data, err := json.Marshal(env)
	if err != nil {
		return
	}
	if err := h.redis.Publish(context.Background(), h.redisChannel, data).Err(); err != nil {
		log.Printf("redis publish failed: %v", err)
	}
}

func newInstanceID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

func userChannel(userID string) string {
	return "ws:user:" + userID
}

// ErrNotUsed is kept to show how to add errors later without import churn.
var ErrNotUsed = errors.New("reserved")
