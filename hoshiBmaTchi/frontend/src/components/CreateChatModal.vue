<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>New Message</h3>
        <button class="close-btn" @click="$emit('close')">✕</button>
      </div>

      <div class="modal-body">
        <div v-if="selectedUsers.length > 1" class="input-group">
          <input 
            v-model="groupName" 
            placeholder="Name your group..." 
            class="text-input"
          />
        </div>

        <div class="search-section">
          <div class="to-label">To:</div>
          <div class="selected-tags">
            <span v-for="user in selectedUsers" :key="user.id" class="user-tag">
              {{ user.username }}
              <button @click="toggleUser(user)">✕</button>
            </span>
            <input 
              v-model="searchQuery" 
              placeholder="Search..." 
              class="search-input"
            />
          </div>
        </div>

        <div class="results-list">
          <div v-if="loading" class="loading">Loading...</div>
          <div 
            v-for="user in searchResults" 
            :key="user.id" 
            class="user-item"
            @click="toggleUser(user)"
          >
            <img :src="user.profile_picture_url || '/placeholder.svg'" class="avatar" />
            <div class="user-info">
              <span class="username">{{ user.username }}</span>
              <span class="fullname">{{ user.name }}</span>
            </div>
            <div class="checkbox">
              <div v-if="isSelected(user.id)" class="checked-indicator">✓</div>
            </div>
          </div>
        </div>

        <button 
          class="create-btn" 
          :disabled="selectedUsers.length === 0"
          @click="createChat"
        >
          Chat
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { usersApi } from '../services/apiService';
import { useDebounce } from '../composables/useDebounce';

const emit = defineEmits(['close', 'create']);

const groupName = ref('');
const searchResults = ref<any[]>([]);
const selectedUsers = ref<any[]>([]);
const loading = ref(false);

const { value: searchQuery, debouncedValue: debouncedSearchQuery } = useDebounce('', 300);

const handleSearch = useDebounce(async () => {
  if (!searchQuery.value) return;
  loading.value = true;
  try {
    const res = await usersApi.searchUsers(searchQuery.value);
    // Filter out already selected users from results if desired, or keep them
    searchResults.value = res.data; 
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
}, 300);

const isSelected = (id: string) => selectedUsers.value.some(u => u.id === id);

const toggleUser = (user: any) => {
  if (isSelected(user.id)) {
    selectedUsers.value = selectedUsers.value.filter(u => u.id !== user.id);
  } else {
    selectedUsers.value.push(user);
  }
  searchQuery.value = ''; // Clear search after selection
  searchResults.value = [];
};

const createChat = async () => {
  const payload = {
    name: groupName.value,
    user_ids: selectedUsers.value.map(u => u.id)
  };
  
  // Call API to create chat
  const token = localStorage.getItem('accessToken');
  try {
    const res = await fetch('/api/chats', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(payload)
    });
    
    if (res.ok) {
      const data = await res.json();
      emit('create', data.conversation_id);
      emit('close');
    }
  } catch (err) {
    console.error("Failed to create chat", err);
  }
};

watch(debouncedSearchQuery, async (newQuery) => {
  if (!newQuery.trim()) {
      searchResults.value = [];
      return;
  }
  
  loading.value = true;
  try {
    const res = await usersApi.searchUsers(newQuery);
    searchResults.value = res.data; 
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: #262626;
  width: 400px;
  height: 500px;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 12px 16px;
  border-bottom: 1px solid #363636;
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: white;
}

.close-btn {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
  font-size: 20px;
}

.modal-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.input-group {
    padding: 10px 16px;
    border-bottom: 1px solid #363636;
}

.text-input {
    width: 100%;
    background: transparent;
    border: none;
    color: white;
    font-size: 14px;
}

.search-section {
  padding: 10px 16px;
  border-bottom: 1px solid #363636;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.to-label {
  color: white;
  font-weight: 600;
  padding-top: 5px;
}

.selected-tags {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.user-tag {
  background: #0095f6;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.user-tag button {
  background: none;
  border: none;
  color: white;
  cursor: pointer;
}

.search-input {
  background: transparent;
  border: none;
  color: white;
  outline: none;
  flex: 1;
  min-width: 100px;
}

.results-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.user-item {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  cursor: pointer;
}

.user-item:hover {
  background: #363636;
}

.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  margin-right: 12px;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.username {
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.fullname {
  color: #a8a8a8;
  font-size: 13px;
}

.checkbox {
  width: 24px;
  height: 24px;
  border: 1px solid #555;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.checked-indicator {
  width: 16px;
  height: 16px;
  background: #0095f6;
  border-radius: 50%;
  color: white;
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.create-btn {
  margin: 16px;
  padding: 12px;
  background: #0095f6;
  border: none;
  border-radius: 8px;
  color: white;
  font-weight: 600;
  cursor: pointer;
}

.create-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>