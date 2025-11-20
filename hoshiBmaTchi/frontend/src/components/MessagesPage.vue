<template>
  <div class="messages-page">
    <MessagesList
      :conversations="conversations"
      :selected-conversation-id="selectedConversationId"
      @select-conversation="selectConversation"
      @delete-conversation="deleteConversation"
    />

    <ChatWindow
      v-if="currentUser"
      :selected-conversation="selectedConversation"
      :messages="messages"
      :current-user-id="currentUser.id"
      @send-message="sendMessage"
      @unsend-message="unsendMessage"
      @delete-conversation="deleteConversationFromChat"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'; // Import lifecycle hooks
import { useRoute } from 'vue-router';   // Import router
import MessagesList from "./MessagesList.vue";
import ChatWindow from "./ChatWindow.vue";
import { useChatStore } from "../composables/useChatStore";

const {
  currentUser,
  conversations,
  selectedConversationId,
  selectedConversation,
  messages,
  selectConversation,
  sendMessage,
  unsendMessage,
  deleteConversation,
  // initialize, 
} = useChatStore();

const route = useRoute();

const handleDeepLink = async () => {
  // 1. Get ID from URL query (e.g., /messages?conversationId=123)
  const targetId = route.query.conversationId as string;

  if (targetId) {
    // 2. Select it (this opens the chat window)
    selectConversation(targetId);
  }
};

onMounted(async () => {
  // If your store needs initialization, await it here:
  // await initialize(); 
  
  handleDeepLink();
});

// Run if the URL changes while staying on the page
watch(() => route.query.conversationId, () => {
  handleDeepLink();
});

const deleteConversationFromChat = () => {
  if (selectedConversationId.value) {
    deleteConversation(selectedConversationId.value);
  }
};

</script>

<style scoped>
.messages-page {
  display: grid;
  grid-template-columns: 360px 1fr;
  gap: 0;
  height: 100vh;
  background: #1a1a1a;
}

@media (max-width: 1024px) {
  .messages-page {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .messages-page {
    grid-template-columns: 1fr;
    height: auto;
  }
}
</style>
