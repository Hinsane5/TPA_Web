<template>
  <div class="blocked-users-container">
    <h2>Blocked Accounts</h2>
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else-if="blockedUsers.length === 0" class="empty-state">
      You haven't blocked anyone.
    </div>

    <div v-else class="user-list">
      <div v-for="user in blockedUsers" :key="user.user_id" class="user-item">
        <div class="user-info">
          <img :src="user.profile_picture_url || '/placeholder.png'" class="avatar" />
          <span class="username">{{ user.username }}</span>
        </div>
        <button class="unblock-btn" @click="unblock(user.user_id)">Unblock</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { usersApi } from '@/services/apiService';

const blockedUsers = ref<any[]>([]);
const loading = ref(true);

const fetchBlockedUsers = async () => {
  try {
    const res = await usersApi.getBlockedUsers();
    blockedUsers.value = res.data.users || [];
  } catch (error) {
    console.error(error);
  } finally {
    loading.value = false;
  }
};

const unblock = async (userId: string) => {
  try {
    await usersApi.unblockUser(userId);
    // Remove from list immediately
    blockedUsers.value = blockedUsers.value.filter(u => u.user_id !== userId);
  } catch (error) {
    alert("Failed to unblock user");
  }
};

onMounted(() => {
  fetchBlockedUsers();
});
</script>

<style scoped>
.blocked-users-container {
  color: white;
  padding: 20px;
}
.user-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #333;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}
.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}
.unblock-btn {
  background: #0095f6;
  color: white;
  border: none;
  padding: 5px 15px;
  border-radius: 5px;
  cursor: pointer;
}
</style>