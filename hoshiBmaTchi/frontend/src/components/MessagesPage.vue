<template>
  <div class="messages-container">
    <div class="conversations-list">
      <div class="conversations-header">
        <h2>Messages</h2>
        <button class="compose-btn" @click="showCompose = true">✏️</button>
      </div>
      
      <div class="search-input-wrapper">
        <input 
          v-model="searchQuery"
          type="text" 
          placeholder="Search conversations..." 
          class="search-input"
        />
      </div>

      <!-- Conversations -->
      <div class="conversations-list-content">
        <!-- Conversations will be populated from backend -->
        <div class="conversation-item" v-for="n in 8" :key="`conversation-${n}`">
          <div class="conversation-avatar">👤</div>
          <div class="conversation-info">
            <p class="conversation-name">User {{ n }}</p>
            <p class="conversation-preview">Last message preview...</p>
          </div>
          <div v-if="n === 1" class="unread-badge">1</div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="!hasConversations" class="empty-state">
        <p>No conversations yet</p>
      </div>
    </div>

    <!-- Chat Area -->
    <div class="chat-area">
      <div class="chat-placeholder">
        <p>Select a conversation to start messaging</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const searchQuery = ref('')
const showCompose = ref(false)
const hasConversations = ref(false)

onMounted(() => {
  // TODO: Fetch conversations from backend
})
</script>

<style scoped>
.messages-container {
  display: grid;
  grid-template-columns: 350px 1fr;
  gap: 1px;
  height: 100%;
  background: #262626;
}

.conversations-list {
  background: #1a1a1a;
  border-right: 1px solid #262626;
  display: flex;
  flex-direction: column;
  color: #fff;
  overflow: hidden;
}

.conversations-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #262626;
  flex-shrink: 0;
}

.conversations-header h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
}

.compose-btn {
  background: none;
  border: none;
  color: #fff;
  font-size: 18px;
  cursor: pointer;
  padding: 5px 10px;
}

.search-input-wrapper {
  padding: 10px 15px;
  border-bottom: 1px solid #262626;
  flex-shrink: 0;
}

.search-input {
  width: 100%;
  background: #262626;
  border: 1px solid #404040;
  color: #fff;
  padding: 8px 12px;
  border-radius: 20px;
  font-size: 13px;
}

.search-input::placeholder {
  color: #808080;
}

.conversations-list-content {
  flex: 1;
  overflow-y: auto;
}

.conversation-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 15px;
  cursor: pointer;
  transition: background 0.2s ease;
  border-bottom: 1px solid #262626;
  position: relative;
}

.conversation-item:hover {
  background: #262626;
}

.conversation-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.conversation-info {
  flex: 1;
  min-width: 0;
}

.conversation-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-preview {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: #a0a0a0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.unread-badge {
  background: #5b5bff;
  color: #fff;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.chat-area {
  background: #1a1a1a;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #a0a0a0;
}

.chat-placeholder {
  text-align: center;
}

.empty-state {
  text-align: center;
  padding: 40px 20px;
  color: #a0a0a0;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Scrollbar styling */
.conversations-list-content::-webkit-scrollbar {
  width: 8px;
}

.conversations-list-content::-webkit-scrollbar-track {
  background: transparent;
}

.conversations-list-content::-webkit-scrollbar-thumb {
  background: #404040;
  border-radius: 4px;
}

/* Responsive */
@media (max-width: 768px) {
  .messages-container {
    grid-template-columns: 1fr;
  }

  .conversations-list {
    max-height: 50%;
    border-right: none;
    border-bottom: 1px solid #262626;
  }

  .chat-area {
    min-height: 300px;
  }
}
</style>
