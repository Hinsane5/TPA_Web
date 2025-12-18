<script setup lang="ts">
import { ref, computed } from 'vue';
import { useStories } from '../composables/useStories';
import ShareUsersContainer from './ShareUsersContainer.vue';
import type { Story } from '../types/stories';

interface Props {
  story?: Story;
}

const emit = defineEmits<{
  close: [];
  send: [];
}>();

const searchQuery = ref('');
const { suggestedUsers, selectedUsers, toggleUserSelection, sendStory } = useStories();

const filteredUsers = computed(() => {
  if (!searchQuery.value.trim()) {
    return suggestedUsers.value;
  }
  return suggestedUsers.value.filter(user =>
    user.username.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
    user.fullName.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});

const closeModal = () => {
  emit('close');
};

const handleSend = () => {
  sendStory();
  emit('send');
  closeModal();
};
</script>

<template>
  <div class="share-modal-overlay" @click.self="closeModal">
    <div class="share-modal" @click.stop>
      <div class="modal-header">
        <h2 class="modal-title">Share</h2>
        <button class="close-modal-btn" @click="closeModal">✕</button>
      </div>

      <div class="modal-content">
        <div class="search-field">
          <span class="search-label">To:</span>
          <input 
            v-model="searchQuery"
            type="text" 
            placeholder="Search..."
            class="search-input"
          />
        </div>

        <div class="suggested-section">
          <h3 class="suggested-title">Suggested</h3>
          <ShareUsersContainer 
            :users="filteredUsers"
            :selected-users="selectedUsers"
            @select-user="toggleUserSelection"
          />
        </div>
      </div>

      <button class="send-btn" :disabled="selectedUsers.size === 0" @click="handleSend">
        Send
      </button>
    </div>
  </div>
</template>

<style scoped>
.share-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: flex-end;
  justify-content: center;
  z-index: 10003;
}

.share-modal {
  background: #1a1a1a;
  width: 100%;
  max-width: 500px;
  border-radius: 16px 16px 0 0;
  display: flex;
  flex-direction: column;
  max-height: 80vh;
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from {
    transform: translateY(100%);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #262626;
}

.modal-title {
  margin: 0;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
}

.close-modal-btn {
  background: none;
  border: none;
  color: #fff;
  font-size: 20px;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s ease;
}

.close-modal-btn:hover {
  opacity: 0.7;
}

.modal-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.search-field {
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-label {
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  min-width: 30px;
}

.search-input {
  flex: 1;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid #262626;
  border-radius: 6px;
  padding: 10px 12px;
  color: #fff;
  font-size: 14px;
  outline: none;
  transition: background 0.2s ease;
}

.search-input::placeholder {
  color: #808080;
}

.search-input:focus {
  background: rgba(255, 255, 255, 0.15);
  border-color: #404040;
}

.suggested-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.suggested-title {
  margin: 0;
  color: #a0a0a0;
  font-size: 13px;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.send-btn {
  background: #0095f6;
  border: none;
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  padding: 12px 20px;
  margin: 16px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s ease;
  width: calc(100% - 32px);
}

.send-btn:hover:not(:disabled) {
  background: #0081d6;
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Scrollbar styling */
.modal-content::-webkit-scrollbar {
  width: 6px;
}

.modal-content::-webkit-scrollbar-track {
  background: transparent;
}

.modal-content::-webkit-scrollbar-thumb {
  background: #404040;
  border-radius: 3px;
}

.modal-content::-webkit-scrollbar-thumb:hover {
  background: #505050;
}
</style>
