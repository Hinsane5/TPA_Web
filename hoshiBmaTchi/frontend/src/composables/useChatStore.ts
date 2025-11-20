import { ref, computed } from "vue";
import type { User, Conversation, Message } from "../types/chat";

// Mock current user
const currentUser = ref<User>({
  id: "user-1",
  username: "wen.fu",
  fullName: "Wen Fu",
  avatar: "/placeholder.svg?height=48&width=48",
  isOnline: true,
});

// Mock conversations data
const conversations = ref<Conversation[]>([
  {
    id: "conv-1",
    type: "direct",
    participants: [
      {
        id: "user-2",
        username: "robert",
        fullName: "Robert Johnson",
        avatar: "/placeholder.svg?height=48&width=48",
        isOnline: true,
      },
    ],
    lastMessage: {
      id: "msg-1",
      senderId: "user-2",
      senderName: "Robert",
      senderAvatar: "/placeholder.svg?height=48&width=48",
      content: "Looks awesome! 👍",
      messageType: "text",
      timestamp: new Date(Date.now() - 3600000),
      status: "seen",
      isEdited: false,
      canUnsend: false,
      createdAt: new Date(Date.now() - 3600000),
    },
    unreadCount: 1,
    updatedAt: new Date(Date.now() - 3600000),
  },
  {
    id: "conv-2",
    type: "direct",
    participants: [
      {
        id: "user-3",
        username: "john",
        fullName: "John Doe",
        avatar: "/placeholder.svg?height=48&width=48",
        isOnline: false,
      },
    ],
    lastMessage: {
      id: "msg-2",
      senderId: "user-1",
      senderName: "You",
      senderAvatar: "/placeholder.svg?height=48&width=48",
      content: "See you tomorrow!",
      messageType: "text",
      timestamp: new Date(Date.now() - 7200000),
      status: "sent",
      isEdited: false,
      canUnsend: false,
      createdAt: new Date(Date.now() - 7200000),
    },
    unreadCount: 0,
    updatedAt: new Date(Date.now() - 7200000),
  },
  {
    id: "conv-3",
    type: "direct",
    participants: [
      {
        id: "user-4",
        username: "mike",
        fullName: "Mike Brown",
        avatar: "/placeholder.svg?height=48&width=48",
        isOnline: true,
      },
    ],
    lastMessage: {
      id: "msg-3",
      senderId: "user-4",
      senderName: "Mike",
      senderAvatar: "/placeholder.svg?height=48&width=48",
      content: "Thanks!",
      messageType: "text",
      timestamp: new Date(Date.now() - 10800000),
      status: "seen",
      isEdited: false,
      canUnsend: false,
      createdAt: new Date(Date.now() - 10800000),
    },
    unreadCount: 0,
    updatedAt: new Date(Date.now() - 10800000),
  },
]);

const allMessages = ref<Message[]>([
  {
    id: "msg-1",
    senderId: "user-2",
    senderName: "Robert",
    senderAvatar: "/placeholder.svg?height=48&width=48",
    content: "Looks awesome! 👍",
    messageType: "text",
    timestamp: new Date(Date.now() - 3600000),
    status: "seen",
    isEdited: false,
    canUnsend: false,
    createdAt: new Date(Date.now() - 3600000),
    conversationId: "conv-1",
  },
]);

const selectedConversationId = ref<string | null>(null);

const selectedConversation = computed<Conversation | null>(() => {
  if (!selectedConversationId.value) return null;
  return (
    conversations.value.find((c) => c.id === selectedConversationId.value) ||
    null
  );
});

// 2. Fix the messages computed property
const messages = computed<Message[]>(() => {
  if (!selectedConversationId.value) return [];
  // In a real app, you filter by conversation ID
  // For this mock, we just return messages if they matched the ID (logic added below)
  return allMessages.value.filter(
    (m) => (m as any).conversationId === selectedConversationId.value
  );
});

export function useChatStore() {
  const selectConversation = (conversationId: string) => {
    selectedConversationId.value = conversationId;
  };

  // 3. Implement Send Logic
  const sendMessage = (content: string) => {
    if (!selectedConversationId.value) return;

    // 1. Create the new message object
    const newMessage: Message = {
      id: `msg-${Date.now()}`,
      senderId: currentUser.value.id,
      senderName: currentUser.value.fullName,
      senderAvatar: currentUser.value.avatar,
      content: content,
      messageType: "text",
      timestamp: new Date(),
      status: "sent",
      isEdited: false,
      canUnsend: true,
      createdAt: new Date(),
      // Ensure your Message type includes this property (see below)
      conversationId: selectedConversationId.value, 
    };

    // 2. Add to messages list
    allMessages.value.push(newMessage);

    // 3. Update conversation (Fix applied here)
    const convIndex = conversations.value.findIndex(
      (c) => c.id === selectedConversationId.value
    );

    if (convIndex !== -1) {
      // FIX: Use strict null check or non-null assertion (!)
      // TypeScript thinks conversations.value[convIndex] might be undefined
      const conversation = conversations.value[convIndex];
      
      if (conversation) {
        conversation.lastMessage = newMessage;
        conversation.updatedAt = new Date(); // Good practice to update timestamp

        // Move conversation to top
        conversations.value.splice(convIndex, 1);
        conversations.value.unshift(conversation);
      }
    }
  };

  const unsendMessage = (messageId: string) => {
    const index = allMessages.value.findIndex((m) => m.id === messageId);
    if (index !== -1) allMessages.value.splice(index, 1);
  };

  const deleteConversation = (conversationId: string) => {
    const index = conversations.value.findIndex((c) => c.id === conversationId);
    if (index !== -1) {
      conversations.value.splice(index, 1);
      if (selectedConversationId.value === conversationId) {
        selectedConversationId.value = null;
      }
    }
  };

  return {
    currentUser,
    conversations,
    selectedConversationId,
    selectedConversation,
    messages,
    selectConversation,
    sendMessage,
    unsendMessage,
    deleteConversation,
  };
}
