import { ref, computed } from "vue";
import type { Conversation, Message, User } from "../types/chat";
import { usersApi } from "../services/apiService";


const currentUser = ref<User | null>(null);
const conversations = ref<Conversation[]>([]);
const messages = ref<Message[]>([]);
const selectedConversationId = ref<string | null>(null);
const isConnected = ref(false);
let socket: WebSocket | null = null;

const API_URL = "/api";
const WS_URL = "ws://localhost:8081/ws";

export function useChatStore() {
  const getToken = () => localStorage.getItem("accessToken");

  const resolveSenderInfo = (senderId: string, conversationId: string) => {
    if (currentUser.value && senderId === currentUser.value.id) {
      return {
        name: currentUser.value.fullName || currentUser.value.username || "Me",
        avatar: currentUser.value.avatar || "/placeholder.svg",
      };
    }

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

        const mappedConversations = data.map((c: any) => ({
          ...c,
          isGroup: c.is_group,
          participants: c.participants || [],
          updatedAt:
            c.last_message?.created_at ||
            c.updated_at ||
            c.created_at ||
            new Date().toISOString(),
        }));

        const enrichedConversations = await Promise.all(
          mappedConversations.map(async (conv: any) => {
            const myId = currentUser.value?.id;

            conv.participants = await Promise.all(
              conv.participants.map(async (p: any) => {
                if (myId && p.id === myId && currentUser.value) {
                  return {
                    ...p,
                    fullName:
                      currentUser.value.fullName || currentUser.value.username,
                    username: currentUser.value.username,
                    avatar: currentUser.value.avatar || "/placeholder.svg",
                  };
                }

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

        const mappedMessages = rawData.map((msg: any) => {
          const { name, avatar } = resolveSenderInfo(
            msg.sender_id,
            conversationId
          );
          const isMedia = !!msg.media_url;

          return {
            id: msg.id,
            conversationId: msg.conversation_id,
            senderId: msg.sender_id, 
            senderName: name,
            senderAvatar: avatar,
            content: isMedia ? msg.media_url : msg.content,
            mediaUrl: msg.media_url, 
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

        if (data.type === "group_created") {
          fetchConversations();
        } else if (data.type === "new_message") {
          handleIncomingMessage(data);
        }
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
    let conversation = conversations.value.find(
      (c) => c.id === wsMsg.conversation_id
    );

    if (!conversation) {
      console.log("New conversation detected, refreshing list...");
      await fetchConversations();
      conversation = conversations.value.find(
        (c) => c.id === wsMsg.conversation_id
      );
    }

    const { name, avatar } = resolveSenderInfo(
      wsMsg.sender_id,
      wsMsg.conversation_id
    );

    const isMedia = !!wsMsg.media_url;


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
        conversation.unreadCount = (conversation.unreadCount || 0) + 1;
      }

      conversations.value = [
        conversation,
        ...conversations.value.filter((c) => c.id !== conversation!.id),
      ];
    }
  };

  const sendMessage = (
    content: string,
    type: "text" | "gif" | "image" = "text",
    mediaUrl?: string
  ) => {
    if (!socket || !selectedConversationId.value || !currentUser.value) return;

    const payload = {
      type: "chat",
      conversation_id: selectedConversationId.value,
      sender_id: currentUser.value.id,
      content: content,
      media_type: type, 
      media_url: mediaUrl, 
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
    fetchConversations,
  };
}
