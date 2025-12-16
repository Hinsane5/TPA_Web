<template>
  <button 
    class="mini-story"
    :class="{ active: isActive }"
    @click="handleClick"
    :title="`Story by ${username}`"
  >
    <div class="mini-avatar-container">
        <img 
          v-if="avatar" 
          :src="avatar" 
          alt="User" 
          class="mini-avatar-img"
        />
        <div v-else class="mini-avatar-placeholder">
            {{ username.charAt(0).toUpperCase() }}
        </div>
    </div>
    
    <span class="mini-username">{{ username }}</span>
  </button>
</template>

<script setup lang="ts">
interface Props {
  avatar?: string; // Made optional
  username?: string; // Made optional
  isActive: boolean;
}

// Default values to prevent crashes
withDefaults(defineProps<Props>(), {
    avatar: '',
    username: 'User'
});

const emit = defineEmits<{
  click: [];
}>();

const handleClick = () => {
  emit('click');
};
</script>

<style scoped>
.mini-story {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  cursor: pointer;
  flex-shrink: 0;
  padding: 0;
  transition: opacity 0.2s ease;
}

.mini-story:hover {
  opacity: 0.8;
}

.mini-story.active {
  opacity: 1;
}

/* FIX: Renamed from .mini-avatar to .mini-avatar-container */
.mini-avatar-container {
  width: 48px;  /* Increased slightly for better visibility */
  height: 48px;
  border-radius: 50%;
  padding: 2px; /* Space between image and border */
  border: 2px solid #262626; /* Default border color */
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden; /* Ensures image stays inside circle */
}

/* Active state (gradient border look) */
.mini-story.active .mini-avatar-container {
  border-color: #0095f6;
}

/* FIX: Added style for the image */
.mini-avatar-img {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

/* Placeholder style */
.mini-avatar-placeholder {
  width: 100%;
  height: 100%;
  background: #333;
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 18px;
}

.mini-username {
  font-size: 11px;
  color: #fff;
  max-width: 60px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
