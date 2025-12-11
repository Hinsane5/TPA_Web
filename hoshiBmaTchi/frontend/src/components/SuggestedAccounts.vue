<template>
  <div class="suggested-container" v-if="suggestedUsers.length > 0">
    <div class="suggested-header">
      <h3>Suggested for you</h3>
      <a href="#" class="see-all">See all</a>
    </div>

    <div class="suggested-list">
      <div 
        v-for="user in suggestedUsers" 
        :key="user.user_id" 
        class="suggested-user-item"
      >
        <div class="user-info-section" @click="goToProfile(user.user_id)">
          <div class="avatar-wrapper">
             <img 
              v-if="user.profile_picture_url" 
              :src="user.profile_picture_url" 
              alt="Profile" 
              class="avatar-img"
            />
            <div v-else class="avatar-placeholder">
              {{ user.username.charAt(0).toUpperCase() }}
            </div>
          </div>
          
          <div class="text-info">
            <span class="username">{{ user.username }}</span>
            <span class="subtext">Suggested for you</span>
          </div>
        </div>

        <button 
          class="follow-btn" 
          @click="handleFollow(user)"
          :disabled="isProcessing === user.user_id"
        >
          {{ isProcessing === user.user_id ? 'Loading...' : 'Follow' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { usersApi } from '../services/apiService';

interface SuggestedUser {
  user_id: string;
  username: string;
  name: string;
  profile_picture_url?: string;
}

const suggestedUsers = ref<SuggestedUser[]>([]);
const isProcessing = ref<string | null>(null);
const router = useRouter();

const fetchSuggestions = async () => {
  try {
    const response = await usersApi.getSuggestedUsers();
    if (response.data && response.data.users) {
      suggestedUsers.value = response.data.users;
    }
  } catch (error) {
    console.error("Failed to fetch suggested users", error);
  }
};

const handleFollow = async (user: SuggestedUser) => {
  isProcessing.value = user.user_id;
  try {
    await usersApi.followUser(user.user_id);
    // Remove user from the list after following
    suggestedUsers.value = suggestedUsers.value.filter(u => u.user_id !== user.user_id);
  } catch (error) {
    console.error("Failed to follow user", error);
  } finally {
    isProcessing.value = null;
  }
};

const goToProfile = (userId: string) => {
  router.push(`/profile/${userId}`);
};

onMounted(() => {
  fetchSuggestions();
});
</script>

<style scoped>
.suggested-container {
  background: #1a1a1a;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #262626;
}

.suggested-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.suggested-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #a8a8a8;
}

.see-all {
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  text-decoration: none;
}

.suggested-user-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.user-info-section {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  flex: 1;
}

.avatar-wrapper {
  width: 44px;
  height: 44px;
  flex-shrink: 0;
}

.avatar-img, .avatar-placeholder {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

.avatar-placeholder {
  background: #262626;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 18px;
}

.text-info {
  display: flex;
  flex-direction: column;
}

.username {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.subtext {
  font-size: 12px;
  color: #a8a8a8;
}

.follow-btn {
  background: transparent;
  border: none;
  color: #0095f6;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  padding: 8px 0;
}

.follow-btn:hover {
  color: #e0f1ff;
}

.follow-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

</style>