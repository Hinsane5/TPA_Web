import { ref, computed } from "vue";
import type { Conversation, Message, User } from "../types/chat";
import { usersApi } from "../services/apiService";

// --- STATE ---
const currentUser = ref<User | null>(null);
const conversations = ref<Conversation[]>([]);
const messages = ref<Message[]>([]);
const selectedConversationId = ref<string | null>(null);
const isConnected = ref(false);
let socket: WebSocket | null = null;

// --- CONFIG ---
// FIX 1: Use relative path '/api' to use the Vite Proxy (points to 8081)
const API_URL = "/api";
// FIX 2: Point WebSocket to the Gateway port (8081)
const WS_URL = "ws://localhost:8081/ws";

export function useChatStore() {
  // FIX 3: Use 'accessToken' to match apiService.ts
  const getToken = () => localStorage.getItem("accessToken");

  const resolveSenderInfo = (senderId: string, conversationId: string) => {
    // 1. Is it Me?
    if (currentUser.value && senderId === currentUser.value.id) {
      return {
        name: currentUser.value.fullName || currentUser.value.username || "Me",
        avatar: currentUser.value.avatar || "/placeholder.svg",
      };
    }

    // 2. Is it a Participant in the conversation?
    const conversation = conversations.value.find(
      (c) => c.id === conversationId
    );
    if (conversation) {
      const participant = conversation.participants.find(
        (p) => p.id === senderId
      );
      if (participant && participant.fullName) {
        return {
          name: participant.fullName,
          avatar: participant.avatar,
        };
      }
    }

    // 3. Fallback
    return { name: "Unknown", avatar: "/placeholder.svg" };
  };

  const initialize = async (user: User) => {
    currentUser.value = user;
    await fetchConversations();
    connectWebSocket();
  };

  // 3. REST API: Fetch Conversations (ROBUST VERSION)
  const fetchConversations = async () => {
    try {
      const token = getToken();
      if (!token) return;

      const res = await fetch(`${API_URL}/chats`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (res.ok) {
        const data = await res.json();

        const mappedConversations = data.map((c: any) => ({
          ...c,
          participants: c.participants || [],
          updatedAt: c.created_at || new Date().toISOString(),
        }));

        // FIX: Iterate ALL participants to ensure everyone has a name
        const enrichedConversations = await Promise.all(
          mappedConversations.map(async (conv: any) => {
            const myId = currentUser.value?.id;

            // Map over the participants array and enrich EACH one
            conv.participants = await Promise.all(
              conv.participants.map(async (p: any) => {
                // 1. If it's Me, use local data (Optimization)
                if (myId && p.id === myId && currentUser.value) {
                  return {
                    ...p,
                    fullName:
                      currentUser.value.fullName || currentUser.value.username,
                    username: currentUser.value.username,
                    avatar: currentUser.value.avatar || "/placeholder.svg",
                  };
                }

                // 2. If it's someone else, fetch their profile from API
                try {
                  const userRes = await usersApi.getUserProfile(p.id);
                  const userData = userRes.data;
                  return {
                    ...p,
                    fullName: userData.name || userData.username,
                    username: userData.username,
                    avatar: userData.profile_picture_url || "/placeholder.svg",
                  };
                } catch (err) {
                  console.warn(`Failed to fetch user ${p.id}`, err);
                  return {
                    ...p,
                    fullName: "Unknown User",
                    avatar: "/placeholder.svg",
                  };
                }
              })
            );

            return conv;
          })
        );

        conversations.value = enrichedConversations;
      }
    } catch (error) {
      console.error("Failed to fetch chats:", error);
    }
  };

  const selectConversation = async (conversationId: string) => {
    selectedConversationId.value = conversationId;
    messages.value = [];

    try {
      const token = getToken();
      if (!token) return;

      const res = await fetch(
        `${API_URL}/chats/${conversationId}/messages?limit=50`,
        { headers: { Authorization: `Bearer ${token}` } }
      );

      if (res.ok) {
        const rawData = await res.json();

        // FIX: Map Snake_Case -> CamelCase & Enrich Sender
        const mappedMessages = rawData.map((msg: any) => {
          const { name, avatar } = resolveSenderInfo(
            msg.sender_id,
            conversationId
          );
          const isMedia = !!msg.media_url;

          return {
            id: msg.id,
            conversationId: msg.conversation_id,
            senderId: msg.sender_id, // Crucial Fix: sender_id -> senderId
            senderName: name,
            senderAvatar: avatar,
            content: isMedia ? msg.media_url : msg.content,
            mediaUrl: msg.media_url, // Crucial Fix: media_url -> mediaUrl
            messageType: isMedia ? "image" : "text",
            createdAt: msg.created_at,
            timestamp: msg.created_at,
            isUnsent: msg.is_unsent,
          };
        });

        messages.value = mappedMessages.reverse();
      }
    } catch (error) {
      console.error("Failed to fetch history:", error);
    }
  };

  const connectWebSocket = () => {
    if (socket) return;

    const token = getToken();
    if (!token) return;

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
      setTimeout(connectWebSocket, 3000);
    };
  };

  const handleIncomingMessage = async (wsMsg: any) => {
    // 1. Check if we have this conversation locally
    let conversation = conversations.value.find(
      (c) => c.id === wsMsg.conversation_id
    );

    // --- FIX: If conversation is missing (New Chat), fetch list immediately ---
    if (!conversation) {
      console.log("New conversation detected, refreshing list...");
      await fetchConversations();
      // Try finding it again after fetch
      conversation = conversations.value.find(
        (c) => c.id === wsMsg.conversation_id
      );
    }

    // 2. Resolve Sender Info (Now works because we fetched the list + participants)
    const { name, avatar } = resolveSenderInfo(
      wsMsg.sender_id,
      wsMsg.conversation_id
    );

    const isMedia = !!wsMsg.media_url;

    // 3. Create Message Object
    const newMessage: Message = {
      id: wsMsg.id || `msg-${Date.now()}`,
      conversationId: wsMsg.conversation_id,
      senderId: wsMsg.sender_id,
      senderName: name,
      senderAvatar: avatar,
      content: isMedia ? wsMsg.media_url : wsMsg.content,
      messageType: isMedia ? "image" : "text",
      mediaUrl: wsMsg.media_url,
      createdAt: new Date().toISOString(),
      timestamp: new Date().toISOString(),
      status: "sent", // You received it, so it's effectively sent/delivered
      isEdited: false,
      canUnsend: false,
    };

    // 4. Add to Message List (only if viewing THIS conversation)
    if (selectedConversationId.value === wsMsg.conversation_id) {
      messages.value.push(newMessage);

      // Optional: Scroll to bottom logic usually goes here or in the component watcher
    }

    // 5. Update Sidebar Preview (Last Message & Unread Count)
    if (conversation) {
      conversation.lastMessage = newMessage;

      // Increment unread count if we aren't looking at this chat
      if (selectedConversationId.value !== wsMsg.conversation_id) {
        conversation.unreadCount = (conversation.unreadCount || 0) + 1;
      }

      // Move this conversation to the top of the list
      conversations.value = [
        conversation,
        ...conversations.value.filter((c) => c.id !== conversation!.id),
      ];
    }
  };

  const sendMessage = (content: string) => {
    if (!socket || !selectedConversationId.value || !currentUser.value) return;

    const payload = {
      type: "chat",
      conversation_id: selectedConversationId.value,
      sender_id: currentUser.value.id,
      content: content,
    };

    socket.send(JSON.stringify(payload));
  };

  const deleteConversation = async (conversationId: string) => {
    const token = getToken();
    await fetch(`${API_URL}/chats/${conversationId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });

    conversations.value = conversations.value.filter(
      (c) => c.id !== conversationId
    );
    if (selectedConversationId.value === conversationId) {
      selectedConversationId.value = null;
    }
  };

  const unsendMessage = async (messageId: string) => {
    const msg = messages.value.find((m) => m.id === messageId);
    if (msg) {
      msg.content = "This message was unsent";
      msg.isUnsent = true;
      msg.mediaUrl = undefined;
      msg.messageType = "text";
    }

    try {
      const token = getToken();
      await fetch(`${API_URL}/chats/messages/${messageId}`, {
        method: "DELETE",
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
