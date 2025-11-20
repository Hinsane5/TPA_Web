<template>
  <div class="messages-page">
    <MessagesList
      :conversations="conversations"
      :selected-conversation-id="selectedConversationId"
      @select-conversation="selectConversation"
      @delete-conversation="deleteConversation"
    />

    <ChatWindow
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
} = useChatStore();

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
