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

  const fetchConversations = async () => {
    try {
      const token = getToken();
      if (!token) return;

      const res = await fetch(`${API_URL}/chats`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (res.ok) {
        const data = await res.json();

        // 1. Map the basic structure
        const mappedConversations = data.map((c: any) => ({
          ...c,
          // FIX: Backend now sends 'participants' (lowercase)
          participants: c.participants || [],
          updatedAt: c.created_at || new Date().toISOString(),
        }));

        // 2. Fetch User Details for each chat (enrichment)
        // We need to find the "other person" in each chat and get their name/avatar
        const enrichedConversations = await Promise.all(
          mappedConversations.map(async (conv: any) => {
            // Identify the partner (the one who isn't ME)
            const myId = currentUser.value?.id;
            const partner =
              conv.participants.find((p: any) => p.id !== myId) ||
              conv.participants[0];

            if (partner && partner.id) {
              try {
                // Fetch their profile from User Service
                const userRes = await usersApi.getUserProfile(partner.id);
                const userData = userRes.data;

                // Update the participant object with real data
                partner.fullName = userData.name || userData.username;
                partner.username = userData.username;
                partner.avatar =
                  userData.profile_picture_url || "/placeholder.svg";
                partner.isOnline = false; // You can hook this up to WS later
              } catch (err) {
                console.warn("Failed to fetch user info for chat:", conv.id);
                partner.fullName = "Unknown User";
                partner.avatar = "/placeholder.svg";
              }
            }
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

  const handleIncomingMessage = (wsMsg: any) => {
    let senderName = "Unknown";
    let senderAvatar = "/placeholder.svg";

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
        senderName = currentUser.value.fullName;
        senderAvatar = currentUser.value.avatar;
      }
    }

    const isMedia = !!wsMsg.media_url;

    const newMessage: Message = {
      id: wsMsg.id || `msg-${Date.now()}`,
      conversationId: wsMsg.conversation_id,
      senderId: wsMsg.sender_id,
      senderName: senderName,
      senderAvatar: senderAvatar,
      content: isMedia ? wsMsg.media_url : wsMsg.content,
      messageType: isMedia ? "image" : "text",
      mediaUrl: wsMsg.media_url,
      createdAt: new Date().toISOString(),
      timestamp: new Date().toISOString(),
      status: "sent",
      isEdited: false,
      canUnsend: false,
    };

    if (selectedConversationId.value === wsMsg.conversation_id) {
      messages.value.push(newMessage);
    }

    if (conversation) {
      conversation.lastMessage = newMessage;
      if (selectedConversationId.value !== wsMsg.conversation_id) {
        conversation.unreadCount += 1;
      }
      conversations.value = [
        conversation,
        ...conversations.value.filter((c) => c.id !== conversation.id),
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
