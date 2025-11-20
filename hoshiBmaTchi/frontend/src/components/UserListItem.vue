<template>
  <div class="user-list-item" @click="$emit('click')">
    <div class="user-avatar-wrapper">
      <img 
        :src="user.profile_picture_url || '/default-avatar.png'" 
        :alt="user.username"
        class="user-avatar"
        @error="handleImageError"
      />
    </div>

    <div class="user-info-wrapper">
      <div class="username-section">
        <span class="username-text">{{ user.username }}</span>
        <img 
          v-if="user.is_verified"
          src="/icons/verified-icon.png"
          alt="Verified"
          class="verified-badge"
        />
      </div>

      <p class="user-details">
        {{ user.name }}
        <span v-if="formattedFollowers"> • {{ formattedFollowers }} followers</span>
      </p>
    </div>

    <div class="action-area" @click.stop>
      <button 
        v-if="showClose" 
        class="close-btn"
        @click="$emit('remove')"
        title="Remove from history"
      >
        ×
      </button>

      <button 
        v-else-if="showFollow"
        class="follow-button"
        @click="handleFollowClick"
        :class="{ following: isFollowingLocal }"
      >
        {{ isFollowingLocal ? 'Following' : 'Follow' }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';

// Define the Props to match your Backend Data Structure
interface Props {
  user: {
    user_id: string;
    username: string;
    name: string; // Backend sends 'name', not 'fullName'
    profile_picture_url?: string; // Backend sends snake_case
    followers?: number; // Optional: Search results might not have this
    is_verified?: boolean;
    is_following?: boolean;
  };
  showClose?: boolean;
  showFollow?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showClose: false,
  showFollow: false
});

const emit = defineEmits(['click', 'remove', 'follow', 'unfollow']);

// Local state for optimistic UI updates
const isFollowingLocal = ref(props.user.is_following || false);

// Watch for prop updates (e.g., if data is refreshed)
watch(() => props.user.is_following, (newVal) => {
  if (newVal !== undefined) isFollowingLocal.value = newVal;
});

// COMPUTED PROPERTY: Safely format followers (Prevents the crash!)
const formattedFollowers = computed(() => {
  if (props.user.followers !== undefined && props.user.followers !== null) {
    return props.user.followers.toLocaleString();
  }
  return null; // Returns null if no follower data (e.g., in Search results)
});

const handleFollowClick = () => {
  if (isFollowingLocal.value) {
    isFollowingLocal.value = false;
    emit('unfollow', props.user.user_id);
  } else {
    isFollowingLocal.value = true;
    emit('follow', props.user.user_id);
  }
};

const handleImageError = (e: Event) => {
  (e.target as HTMLImageElement).src = '/default-avatar.png';
};
</script>

<style scoped>
.user-list-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  /* Changed background to dark theme (#1a1a1a instead of white) */
  background-color: #121212;
  border-bottom: 1px solid #262626;
  transition: background-color 0.2s ease;
}

.user-list-item:hover {
  /* Changed hover background to dark theme */
  background-color: #262626;
}

/* User Avatar */
.user-avatar-wrapper {
  flex-shrink: 0;
}

.user-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

/* User Info */
.user-info-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.username-section {
  display: flex;
  align-items: center;
  gap: 6px;
}

.username-text {
  font-size: 14px;
  font-weight: 600;
  /* Changed text color from black to white for dark theme */
  color: #ffffff;
  word-break: break-word;
}

.verified-badge {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

.user-details {
  font-size: 13px;
  /* Changed text color to light gray for dark theme */
  color: #a0a0a0;
  margin: 0;
  line-height: 1.4;
  word-break: break-word;
}

/* Action Area */
.action-area {
  flex-shrink: 0;
  display: flex;
  align-items: center;
}

/* Styled close button with dark theme (×) */
.close-btn {
  background: none;
  border: none;
  color: #808080;
  font-size: 20px;
  cursor: pointer;
  padding: 4px 8px;
  transition: color 0.2s ease;
  font-family: inherit;
}

.close-btn:hover {
  color: #ffffff;
}

/* Follow Button */
.follow-button {
  flex-shrink: 0;
  padding: 8px 20px;
  background-color: #0095f6;
  color: #ffffff;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.follow-button:hover {
  background-color: #007bd2;
}

.follow-button:active {
  transform: scale(0.98);
}

.follow-button.following {
  background-color: transparent;
  /* Changed border color for dark theme */
  color: #ffffff;
  border: 1px solid #404040;
}

.follow-button.following:hover {
  background-color: #262626;
  border-color: #505050;
}

/* Responsive Design */
@media (max-width: 480px) {
  .user-list-item {
    padding: 10px 12px;
    gap: 10px;
  }

  .user-avatar {
    width: 40px;
    height: 40px;
  }

  .follow-button {
    padding: 6px 16px;
    font-size: 12px;
  }
}
</style>
