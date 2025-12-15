<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{{ conversation.isGroup ? 'Group Details' : 'Details' }}</h3>
        <button class="close-btn" @click="$emit('close')">✕</button>
      </div>
      
      <div class="modal-body">
        <div class="participants-section">
          <h4>Members</h4>
          <div class="member-list">
             <div v-for="member in conversation.participants" :key="member.id" class="member-item">
                <img :src="member.avatar || '/placeholder.svg'" class="avatar" />
                <div class="member-info">
                    <span class="username">{{ member.username }}</span>
                </div>
                <button 
                  v-if="conversation.isGroup && member.id !== currentUserId" 
                  class="action-btn remove"
                  @click="removeMember(member.id)"
                >
                  Remove
                </button>
             </div>
          </div>
        </div>

        <div v-if="conversation.isGroup" class="add-section">
            <h4>Add People</h4>
             <div class="search-box">
                <input v-model="addQuery" placeholder="Search user to add..."/>
             </div>
             <div v-if="searchResults.length" class="mini-results">
                 <div v-for="user in searchResults" :key="user.id" class="result-item" @click="addMember(user.id)">
                     <span>{{ user.username }}</span>
                     <span class="plus">+</span>
                 </div>
             </div>
        </div>

        <div class="footer-actions">
             <button class="danger-btn" @click="$emit('leave')">
                {{ conversation.isGroup ? 'Leave Group' : 'Delete Chat' }}
             </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import type { Conversation } from '../types/chat';
import { usersApi } from '../services/apiService';
import { useDebounce } from '../composables/useDebounce';

const props = defineProps<{
    conversation: Conversation;
    currentUserId: string;
}>();

const emit = defineEmits(['close', 'leave', 'refresh']);
const { value: addQuery, debouncedValue: debouncedAddQuery } = useDebounce('', 300);
const searchResults = ref<any[]>([]);

const handleSearch = useDebounce(async () => {
  if (!addQuery.value) return;
  const res = await usersApi.searchUsers(addQuery.value);
  // Filter out existing members
  searchResults.value = res.data.filter((u: any) => 
    !props.conversation.participants.some(p => p.id === u.id)
  );
}, 300);

const token = localStorage.getItem('accessToken');

const addMember = async (userId: string) => {
    try {
        await fetch(`/api/chats/${props.conversation.id}/participants`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
            body: JSON.stringify({ user_id: userId })
        });
        addQuery.value = '';
        searchResults.value = [];
        emit('refresh'); // Refresh conversation data
    } catch (e) { console.error(e); }
};

const removeMember = async (userId: string) => {
    try {
        await fetch(`/api/chats/${props.conversation.id}/participants`, {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
            body: JSON.stringify({ user_id: userId })
        });
        emit('refresh');
    } catch (e) { console.error(e); }
};

watch(debouncedAddQuery, async (newQuery) => {
  if (!newQuery.trim()) return;
  
  try {
      const res = await usersApi.searchUsers(newQuery);
      searchResults.value = res.data.filter((u: any) => 
        !props.conversation.participants.some(p => p.id === u.id)
      );
  } catch (e) { console.error(e); }
});
</script>

<style scoped>
/* Reuse styles from CreateChatModal or similar */
.modal-overlay { 
    position: fixed; 
    inset: 0; 
    background: rgba(0,0,0,0.7); 
    display:flex; 
    justify-content:center; 
    align-items:center; 
    z-index:1001; 
}

.modal-content { 
    background: #262626; 
    width: 350px; 
    border-radius: 12px; 
    color: white; 
    padding-bottom: 20px; 
}

.modal-header { 
    padding: 15px; 
    border-bottom: 1px solid #363636; 
    display:flex; 
    justify-content:space-between; 
}

.close-btn { 
    background:none; 
    border:none; 
    color:white; 
    font-size:18px; 
    cursor:pointer; 
}
/* Modal */
.modal-body {
  padding: 15px;
}

/* Member List */
.member-list {
  margin-top: 10px;
  max-height: 200px;
  overflow-y: auto;
}

.member-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 0;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
}

.member-info {
  flex: 1;
}

/* Buttons */
.action-btn {
  background: #363636;
  border: none;
  color: #fff;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

.danger-btn {
  width: 100%;
  margin-top: 20px;
  padding: 10px;

  background: transparent;
  color: #ed4956;

  border: 1px solid #363636;
  border-radius: 8px;
  cursor: pointer;
}

/* Search */
.search-box input {
  width: 100%;
  margin-top: 5px;
  padding: 8px;

  background: #1a1a1a;
  color: #fff;

  border: none;
  border-radius: 6px;
}

/* Search Results */
.mini-results {
  margin-top: 5px;
  background: #1a1a1a;
  border-radius: 6px;
}

.result-item {
  padding: 8px;
  display: flex;
  justify-content: space-between;
  cursor: pointer;
}

.result-item:hover {
  background: #333;
}

</style>