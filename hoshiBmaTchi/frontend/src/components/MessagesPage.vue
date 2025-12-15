<template>
  <div class="messages-page">
    <MessagesList
      v-if="currentUser" 
      :conversations="conversations"
      :selected-conversation-id="selectedConversationId"
      :current-user-id="currentUser.id" 
      :current-user="currentUser" 
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
      @refresh-data="initialize(currentUser)" 
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'; 
import { useRoute } from 'vue-router';  
import MessagesList from "./MessagesList.vue";
import ChatWindow from "./ChatWindow.vue";
import { useChatStore } from "../composables/useChatStore";
import { usersApi } from "../services/apiService";

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
  initialize, 
} = useChatStore();

const route = useRoute();

const handleDeepLink = async () => {
  const targetId = route.query.conversationId as string;

  if (targetId) {
    selectConversation(targetId);
  }
};

onMounted(async () => {
  try {
    const { data: user } = await usersApi.getMe();

    await initialize(user);

    await handleDeepLink();
  } catch (error) {
    console.error("Failed to initialize messages page:", error);
  }
});

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
