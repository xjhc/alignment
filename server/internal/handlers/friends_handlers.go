package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/actors"
	"github.com/xjhc/alignment/server/internal/store"
)

// NotificationSender interface for sending real-time notifications
type NotificationSender interface {
	SendToPlayer(gameID, playerID string, event core.Event) error
	GetPlayerActor(playerID string) (*actors.PlayerActor, bool)
}

// FriendsHandlers handles friends system related HTTP requests
type FriendsHandlers struct {
	postgresStore      *store.PostgresStore
	notificationSender NotificationSender
}

// NewFriendsHandlers creates a new friends handlers instance
func NewFriendsHandlers(postgresStore *store.PostgresStore) *FriendsHandlers {
	return &FriendsHandlers{
		postgresStore: postgresStore,
	}
}

// SetNotificationSender sets the notification sender for real-time notifications
func (fh *FriendsHandlers) SetNotificationSender(sender NotificationSender) {
	fh.notificationSender = sender
}

// sendFriendNotification sends a real-time notification to a player if they are online
func (fh *FriendsHandlers) sendFriendNotification(playerID, notificationType string, payload map[string]interface{}) {
	if fh.notificationSender == nil {
		return // No notification sender available
	}

	// Check if player is online
	if _, isOnline := fh.notificationSender.GetPlayerActor(playerID); !isOnline {
		return // Player is not online, they'll see the update when they fetch their friends list
	}

	// Create the notification event
	event := core.Event{
		ID:        uuid.New().String(),
		Type:      core.EventPrivateNotification,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"type": notificationType,
		},
	}

	// Add additional payload data
	for key, value := range payload {
		event.Payload[key] = value
	}

	// Send the notification (gameID can be empty for friend notifications)
	if err := fh.notificationSender.SendToPlayer("", playerID, event); err != nil {
		// Log error but don't fail the request - the user will see updates when they refresh
		fmt.Printf("Failed to send friend notification to player %s: %v\n", playerID, err)
	}
}

// getFriendRequestByID is a helper method to get friend request details by ID
func (fh *FriendsHandlers) getFriendRequestByID(requestID string) (*store.FriendRequest, error) {
	return fh.postgresStore.GetFriendRequestByID(requestID)
}

// SendFriendRequestRequest represents the request to send a friend request
type SendFriendRequestRequest struct {
	RecipientID string `json:"recipientId"`
}

// FriendRequestResponse represents a friend request with user details
type FriendRequestResponse struct {
	ID          string `json:"id"`
	RequesterID string `json:"requesterId"`
	RecipientID string `json:"recipientId"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	// Include requester/recipient details for display
	RequesterUsername    string `json:"requesterUsername,omitempty"`
	RequesterDisplayName string `json:"requesterDisplayName,omitempty"`
	RecipientUsername    string `json:"recipientUsername,omitempty"`
	RecipientDisplayName string `json:"recipientDisplayName,omitempty"`
}

// FriendsListResponse represents the friends list with requests
type FriendsListResponse struct {
	Friends           []store.FriendWithStatus    `json:"friends"`
	IncomingRequests  []FriendRequestResponse     `json:"incomingRequests"`
	OutgoingRequests  []FriendRequestResponse     `json:"outgoingRequests"`
}

// SendFriendRequest handles sending a friend request
func (fh *FriendsHandlers) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract requester ID from URL path or query parameter
	requesterID := r.URL.Query().Get("id")
	if requesterID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req SendFriendRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RecipientID == "" {
		http.Error(w, "RecipientID is required", http.StatusBadRequest)
		return
	}

	// Prevent self-friend requests
	if requesterID == req.RecipientID {
		http.Error(w, "Cannot send friend request to yourself", http.StatusBadRequest)
		return
	}

	// Check if recipient exists
	_, err := fh.postgresStore.GetPlayerByID(req.RecipientID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Recipient not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to verify recipient: %v", err), http.StatusInternalServerError)
		return
	}

	err = fh.postgresStore.SendFriendRequest(requesterID, req.RecipientID)
	if err != nil {
		if strings.Contains(err.Error(), "already friends") {
			http.Error(w, "Players are already friends", http.StatusConflict)
			return
		}
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, "Friend request already sent", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to send friend request: %v", err), http.StatusInternalServerError)
		return
	}

	// Get requester details for the notification
	requester, err := fh.postgresStore.GetPlayerByID(requesterID)
	if err == nil {
		// Send real-time notification to recipient
		fh.sendFriendNotification(req.RecipientID, "FRIEND_REQUEST_RECEIVED", map[string]interface{}{
			"requesterId":        requesterID,
			"requesterUsername":  requester.Username,
			"requesterDisplayName": requester.DisplayName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend request sent successfully",
	})
}

// AcceptFriendRequest handles accepting a friend request
func (fh *FriendsHandlers) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract request ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/friends/requests/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "accept" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	requestID := parts[0]

	if requestID == "" {
		http.Error(w, "Request ID is required", http.StatusBadRequest)
		return
	}

	// Get the friend request details before accepting
	friendRequestDetails, err := fh.getFriendRequestByID(requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Friend request not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get friend request: %v", err), http.StatusInternalServerError)
		return
	}

	err = fh.postgresStore.AcceptFriendRequest(requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Friend request not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to accept friend request: %v", err), http.StatusInternalServerError)
		return
	}

	// Get recipient details for the notification
	recipient, err := fh.postgresStore.GetPlayerByID(friendRequestDetails.RecipientID)
	if err == nil {
		// Send real-time notification to requester that their request was accepted
		fh.sendFriendNotification(friendRequestDetails.RequesterID, "FRIEND_REQUEST_ACCEPTED", map[string]interface{}{
			"accepterId":         friendRequestDetails.RecipientID,
			"accepterUsername":   recipient.Username,
			"accepterDisplayName": recipient.DisplayName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend request accepted successfully",
	})
}

// DeclineFriendRequest handles declining a friend request
func (fh *FriendsHandlers) DeclineFriendRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract request ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/friends/requests/")
	requestID := path

	if requestID == "" {
		http.Error(w, "Request ID is required", http.StatusBadRequest)
		return
	}

	// Get the friend request details before declining
	friendRequestDetails, err := fh.getFriendRequestByID(requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Friend request not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get friend request: %v", err), http.StatusInternalServerError)
		return
	}

	err = fh.postgresStore.DeclineFriendRequest(requestID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Friend request not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to decline friend request: %v", err), http.StatusInternalServerError)
		return
	}

	// Get recipient details for the notification
	recipient, err := fh.postgresStore.GetPlayerByID(friendRequestDetails.RecipientID)
	if err == nil {
		// Send real-time notification to requester that their request was declined
		fh.sendFriendNotification(friendRequestDetails.RequesterID, "FRIEND_REQUEST_DECLINED", map[string]interface{}{
			"declinerId":         friendRequestDetails.RecipientID,
			"declinerUsername":   recipient.Username,
			"declinerDisplayName": recipient.DisplayName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend request declined successfully",
	})
}

// GetFriendsList handles retrieving friends list with requests
func (fh *FriendsHandlers) GetFriendsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get friends with status
	friends, err := fh.postgresStore.GetFriends(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get friends: %v", err), http.StatusInternalServerError)
		return
	}

	// Get incoming friend requests
	incomingRequests, err := fh.postgresStore.GetIncomingFriendRequests(userID)
	if err != nil {
		incomingRequests = []store.FriendRequest{} // Continue with empty list if fetch fails
	}

	// Get outgoing friend requests
	outgoingRequests, err := fh.postgresStore.GetOutgoingFriendRequests(userID)
	if err != nil {
		outgoingRequests = []store.FriendRequest{} // Continue with empty list if fetch fails
	}

	// Convert incoming requests to response format with requester details
	incomingResponses := make([]FriendRequestResponse, 0, len(incomingRequests))
	for _, req := range incomingRequests {
		response := FriendRequestResponse{
			ID:          req.ID,
			RequesterID: req.RequesterID,
			RecipientID: req.RecipientID,
			Status:      req.Status,
			CreatedAt:   req.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		
		// Get requester details
		if requester, err := fh.postgresStore.GetPlayerByID(req.RequesterID); err == nil {
			response.RequesterUsername = requester.Username
			response.RequesterDisplayName = requester.DisplayName
		}
		
		incomingResponses = append(incomingResponses, response)
	}

	// Convert outgoing requests to response format with recipient details
	outgoingResponses := make([]FriendRequestResponse, 0, len(outgoingRequests))
	for _, req := range outgoingRequests {
		response := FriendRequestResponse{
			ID:          req.ID,
			RequesterID: req.RequesterID,
			RecipientID: req.RecipientID,
			Status:      req.Status,
			CreatedAt:   req.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		
		// Get recipient details
		if recipient, err := fh.postgresStore.GetPlayerByID(req.RecipientID); err == nil {
			response.RecipientUsername = recipient.Username
			response.RecipientDisplayName = recipient.DisplayName
		}
		
		outgoingResponses = append(outgoingResponses, response)
	}

	response := FriendsListResponse{
		Friends:          friends,
		IncomingRequests: incomingResponses,
		OutgoingRequests: outgoingResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveFriendRequest represents the request to remove a friend
type RemoveFriendRequest struct {
	FriendID string `json:"friendId"`
}

// RemoveFriend handles removing a friend
func (fh *FriendsHandlers) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req RemoveFriendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FriendID == "" {
		http.Error(w, "FriendID is required", http.StatusBadRequest)
		return
	}

	// Prevent removing yourself
	if userID == req.FriendID {
		http.Error(w, "Cannot remove yourself", http.StatusBadRequest)
		return
	}

	err := fh.postgresStore.RemoveFriend(userID, req.FriendID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Friendship not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to remove friend: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Friend removed successfully",
	})
}

// GetFriendsStatus handles retrieving presence status for multiple players
func (fh *FriendsHandlers) GetFriendsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract friend IDs from query parameter (comma-separated)
	friendIDsParam := r.URL.Query().Get("ids")
	if friendIDsParam == "" {
		http.Error(w, "Friend IDs are required", http.StatusBadRequest)
		return
	}

	friendIDs := strings.Split(friendIDsParam, ",")
	if len(friendIDs) == 0 {
		http.Error(w, "At least one friend ID is required", http.StatusBadRequest)
		return
	}

	// Get presence status for all friends
	presenceMap, err := fh.postgresStore.GetMultiplePlayerPresence(friendIDs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get friends status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presenceMap)
}