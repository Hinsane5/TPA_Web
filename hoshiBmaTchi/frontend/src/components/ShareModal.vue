<template>
  <div class="share-modal-overlay" @click.self="closeModal">
    <div class="share-modal">
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
          <div v-if="loading" class="loading-state">Loading...</div>
          <div v-else class="users-list">
             <div 
                v-for="user in filteredUsers" 
                :key="user.user_id || user.id" 
                class="user-item"
                @click="toggleSelection(user.user_id || user.id)"
             >
                <div class="user-info">
                   <img :src="user.profile_picture_url || '/default-avatar.png'" class="user-avatar" />
                   <div class="user-details">
                      <span class="user-name">{{ user.username }}</span>
                      <span class="user-fullname">{{ user.name || user.full_name }}</span>
                   </div>
                </div>
                <div class="checkbox">
                   <div v-if="selectedUserId === (user.user_id || user.id)" class="checked-circle">✓</div>
                   <div v-else class="unchecked-circle"></div>
                </div>
             </div>
          </div>
        </div>
      </div>

      <button class="send-btn" @click="handleSend" :disabled="!selectedUserId">
        Send
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { reelsApi, chatApi } from '../services/apiService';

const props = defineProps<{
  contentId: string;
  type: 'post' | 'story' | 'reel';
  thumbnail?: string;
}>();

const emit = defineEmits(['close', 'sent']);

const searchQuery = ref('');
const users = ref<any[]>([]);
const loading = ref(false);
const selectedUserId = ref<string | null>(null);

onMounted(async () => {
  loading.value = true;
  try {
    // Fetch users (using "following" list as recipients)
    const res = await reelsApi.getShareRecipients();
    users.value = res.data.users || res.data || []; 
  } catch (error) {
    console.error("Failed to load share recipients", error);
  } finally {
    loading.value = false;
  }
});

const filteredUsers = computed(() => {
  if (!searchQuery.value.trim()) return users.value;
  const q = searchQuery.value.toLowerCase();
  return users.value.filter(u => 
    u.username.toLowerCase().includes(q) || 
    (u.name && u.name.toLowerCase().includes(q))
  );
});

const toggleSelection = (id: string) => {
  selectedUserId.value = selectedUserId.value === id ? null : id;
};

const closeModal = () => {
  emit('close');
};

const handleSend = async () => {
  if (!selectedUserId.value) return;

  try {
    await chatApi.shareContent({
      recipient_id: selectedUserId.value,
      content_id: props.contentId,
      type: props.type,
      thumbnail: props.thumbnail
    });
    alert("Sent!");
    emit('sent');
    closeModal();
  } catch (error) {
    console.error("Failed to share", error);
    alert("Failed to send.");
  }
};
</script>

<style scoped>
.share-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10005;
}

.share-modal {
  background: #262626;
  width: 100%;
  max-width: 400px;
  height: 60vh;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  color: white;
}

.modal-header {
  display: flex;
  justify-content: center;
  position: relative;
  padding: 12px;
  border-bottom: 1px solid #363636;
}

.modal-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.close-modal-btn {
  position: absolute;
  right: 12px;
  background: none;
  border: none;
  color: white;
  font-size: 20px;
  cursor: pointer;
}

.modal-content {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.search-field {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
  align-items: center;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  color: white;
  font-size: 14px;
  outline: none;
}

.users-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  padding: 4px 0;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.user-name { font-weight: 600; font-size: 14px; }
.user-fullname { color: #a8a8a8; font-size: 13px; }

.checked-circle {
  width: 24px;
  height: 24px;
  background: #0095f6;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
}

.unchecked-circle {
  width: 24px;
  height: 24px;
  border: 2px solid #a8a8a8;
  border-radius: 50%;
}

.send-btn {
  margin: 16px;
  padding: 12px;
  background: #0095f6;
  border: none;
  border-radius: 8px;
  color: white;
  font-weight: 600;
  cursor: pointer;
}
.send-btn:disabled { opacity: 0.5; cursor: not-allowed; }
</style>