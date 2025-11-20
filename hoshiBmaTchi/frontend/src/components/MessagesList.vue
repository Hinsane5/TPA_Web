<template>
  <div class="messages-list">
    <div class="list-header">
      <h2>Messages</h2>
      <button class="compose-btn" title="New message">✎</button>
    </div>

    <div class="search-wrapper">
      <input
        v-model="searchInput"
        type="text"
        placeholder="Search conversations..."
        class="search-input"
      />
    </div>

    <div class="conversations-container">
      <div v-if="filteredConversations.length === 0" class="empty-state">
        <p>
          {{ searchInput ? "No conversations found" : "No conversations yet" }}
        </p>
      </div>

      <ConversationItem
        v-for="conversation in filteredConversations"
        :key="conversation.id"
        :conversation="conversation"
        :is-active="selectedConversationId === conversation.id"
        @select="selectConversation(conversation.id)"
        @delete="deleteConversation(conversation.id)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import type { Conversation } from "../types/chat";
import ConversationItem from "./ConversationItem.vue";

interface Props {
  conversations: Conversation[];
  selectedConversationId: string | null;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  "select-conversation": [conversationId: string];
  "delete-conversation": [conversationId: string];
}>();

const searchInput = ref("");

const filteredConversations = computed(() => {
  if (!searchInput.value) {
    return props.conversations;
  }

  const query = searchInput.value.toLowerCase();
  return props.conversations.filter((conv) => {
    const participant = conv.participants[0];
    
    // --- FIX: Check if participant exists before accessing properties ---
    if (!participant) return false; 

    return (
      participant.username.toLowerCase().includes(query) ||
      participant.fullName.toLowerCase().includes(query)
    );
  });
});

const selectConversation = (conversationId: string) => {
  emit("select-conversation", conversationId);
};

const deleteConversation = (conversationId: string) => {
  emit("delete-conversation", conversationId);
};
</script>

<style scoped>
.messages-list {
  width: 360px;
  height: 100%;
  background: #1a1a1a;
  border-right: 1px solid #262626;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #262626;
  flex-shrink: 0;
}

.list-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: #fff;
}

.compose-btn {
  background: none;
  border: none;
  color: #0084ff;
  font-size: 20px;
  cursor: pointer;
  padding: 0;
  transition: transform 0.2s ease;
}

.compose-btn:hover {
  transform: scale(1.1);
}

.search-wrapper {
  padding: 12px 16px;
  border-bottom: 1px solid #262626;
  flex-shrink: 0;
}

.search-input {
  width: 100%;
  background: #262626;
  border: 1px solid #404040;
  color: #fff;
  padding: 10px 16px;
  border-radius: 20px;
  font-size: 13px;
  font-family: inherit;
  transition: border-color 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #0084ff;
}

.search-input::placeholder {
  color: #808080;
}

.conversations-container {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  color: #a0a0a0;
  font-size: 14px;
  text-align: center;
  padding: 40px 20px;
}

.conversations-container::-webkit-scrollbar {
  width: 8px;
}

.conversations-container::-webkit-scrollbar-track {
  background: transparent;
}

.conversations-container::-webkit-scrollbar-thumb {
  background: #404040;
  border-radius: 4px;
}

.conversations-container::-webkit-scrollbar-thumb:hover {
  background: #505050;
}

@media (max-width: 768px) {
  .messages-list {
    width: 100%;
    height: auto;
    border-right: none;
    border-bottom: 1px solid #262626;
  }

  .conversations-container {
    max-height: 300px;
  }
}
</style>