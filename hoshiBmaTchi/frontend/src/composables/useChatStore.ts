import { ref, computed } from "vue";
import type { Conversation, Message, User } from "../types/chat";

// --- STATE ---
const currentUser = ref<User | null>(null);
const conversations = ref<Conversation[]>([]);
const messages = ref<Message[]>([]); // Messages for the ACTIVE conversation
const selectedConversationId = ref<string | null>(null);
const isConnected = ref(false);
let socket: WebSocket | null = null;

// --- CONFIG ---
// Point this to your Gateway URL
const API_URL = "http://localhost:8000/api";
const WS_URL = "ws://localhost:8000/ws";

export function useChatStore() {
  // 1. GET AUTH TOKEN (Helper)
  const getToken = () => localStorage.getItem("token"); // Or use your useAuth composable

  // 2. INITIALIZE (Call this when App mounts or User logs in)
  const initialize = async (user: User) => {
    currentUser.value = user;
    await fetchConversations();
    connectWebSocket();
  };

  // 3. REST API: Fetch Conversations
  const fetchConversations = async () => {
    try {
      const token = getToken();
      if (!token) return;

      const res = await fetch(`${API_URL}/chats`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (res.ok) {
        const data = await res.json();
        // Map backend data to frontend shape if necessary
        conversations.value = data.map((c: any) => ({
          ...c,
          participants: c.Participants || [], // Handle capitalization differences
          updatedAt: c.CreatedAt, // Or LastMessageAt if you have it
        }));
      }
    } catch (error) {
      console.error("Failed to fetch chats:", error);
    }
  };

  // 4. REST API: Fetch Messages for specific chat
  const selectConversation = async (conversationId: string) => {
    selectedConversationId.value = conversationId;
    messages.value = []; // Clear old messages instantly

    try {
      const token = getToken();
      const res = await fetch(
        `${API_URL}/chats/${conversationId}/messages?limit=50`,
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );

      if (res.ok) {
        const data = await res.json();
        messages.value = data.reverse(); // Backend usually sends newest first, UI often needs oldest first
      }
    } catch (error) {
      console.error("Failed to fetch history:", error);
    }
  };

  // 5. WEBSOCKET CONNECTION
  const connectWebSocket = () => {
    if (socket) return; // Already connected

    const token = getToken();
    if (!token) return;

    // Connect via Gateway
    socket = new WebSocket(`${WS_URL}?token=${token}`);

    socket.onopen = () => {
      console.log("WS Connected");
      isConnected.value = true;
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleIncomingMessage(data);
      } catch (e) {
        console.error("WS Parse Error", e);
      }
    };

    socket.onclose = () => {
      isConnected.value = false;
      socket = null;
      // Optional: Implement reconnect logic here
      setTimeout(connectWebSocket, 3000);
    };
  };

  // 6. HANDLE INCOMING MESSAGES
  // ... inside useChatStore.ts ...

  // 6. HANDLE INCOMING MESSAGES
  const handleIncomingMessage = (wsMsg: any) => {
    // 1. Find Sender Info (Backend WS doesn't send name/avatar, so look it up locally)
    let senderName = "Unknown";
    let senderAvatar = "/placeholder.svg"; // Default avatar

    const conversation = conversations.value.find(
      (c) => c.id === wsMsg.conversation_id
    );
    if (conversation) {
      const sender = conversation.participants.find(
        (p) => p.id === wsMsg.sender_id
      );
      if (sender) {
        senderName = sender.fullName;
        senderAvatar = sender.avatar;
      } else if (
        currentUser.value &&
        wsMsg.sender_id === currentUser.value.id
      ) {
        // It's me
        senderName = currentUser.value.fullName;
        senderAvatar = currentUser.value.avatar;
      }
    }

    // 2. Determine Type
    const isMedia = !!wsMsg.media_url;

    // 3. Create Message Object
    const newMessage: Message = {
      id: wsMsg.id || `msg-${Date.now()}`, // Prefer ID from backend if available
      conversationId: wsMsg.conversation_id,
      senderId: wsMsg.sender_id,
      senderName: senderName, // Filled from local lookup
      senderAvatar: senderAvatar, // Filled from local lookup

      // UI expects the Image URL to be in 'content' if type is 'image'
      content: isMedia ? wsMsg.media_url : wsMsg.content,

      // FIX: Correct property name 'messageType'
      messageType: isMedia ? "image" : "text",

      mediaUrl: wsMsg.media_url,
      createdAt: new Date().toISOString(),
      timestamp: new Date().toISOString(),
      status: "sent",
      isEdited: false,
      canUnsend: false,
    };

    // 4. Add to List (if chat is open)
    if (selectedConversationId.value === wsMsg.conversation_id) {
      messages.value.push(newMessage);
    }

    // 5. Update Conversation List Preview
    if (conversation) {
      conversation.lastMessage = newMessage;
      if (selectedConversationId.value !== wsMsg.conversation_id) {
        conversation.unreadCount += 1;
      }
      // Move to top
      conversations.value = [
        conversation,
        ...conversations.value.filter((c) => c.id !== conversation.id),
      ];
    }
  };

  // 7. SEND MESSAGE
  const sendMessage = (content: string) => {
    if (!socket || !selectedConversationId.value || !currentUser.value) return;

    const payload = {
      type: "chat", // Defined in your backend Hub
      conversation_id: selectedConversationId.value,
      sender_id: currentUser.value.id,
      content: content,
    };

    socket.send(JSON.stringify(payload));

    // Optimistic UI update (optional: show immediately before server confirms)
    // logic similar to handleIncomingMessage...
  };

  const deleteConversation = async (conversationId: string) => {
    // Call DELETE endpoint
    const token = getToken();
    await fetch(`${API_URL}/chats/${conversationId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    // Update local state
    conversations.value = conversations.value.filter(
      (c) => c.id !== conversationId
    );
    if (selectedConversationId.value === conversationId) {
      selectedConversationId.value = null;
    }
  };

  const unsendMessage = async (messageId: string) => {
    // 1. Optimistic Update (Update UI immediately)
    const msg = messages.value.find((m) => m.id === messageId);
    if (msg) {
      msg.content = "This message was unsent";
      msg.isUnsent = true;
      msg.mediaUrl = undefined;
      msg.messageType = "text";
    }

    // 2. Call Backend API
    // (Assuming your backend has a DELETE or PUT endpoint for this)
    try {
      const token = localStorage.getItem("token");
      await fetch(`http://localhost:8080/api/chats/messages/${messageId}`, {
        method: "DELETE", // or PUT depending on your backend implementation
        headers: { Authorization: `Bearer ${token}` },
      });
    } catch (err) {
      console.error("Failed to unsend:", err);
    }
  };

  const selectedConversation = computed(
    () =>
      conversations.value.find((c) => c.id === selectedConversationId.value) ||
      null
  );

  return {
    currentUser,
    conversations,
    messages,
    selectedConversationId,
    selectedConversation,
    isConnected,
    initialize,
    selectConversation,
    sendMessage,
    deleteConversation,
    unsendMessage,
  };
}
