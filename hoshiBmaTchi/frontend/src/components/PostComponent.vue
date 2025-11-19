<template>
  <div class="post-container">
    <div class="post-header">
      <div class="header-left">
        <img v-if="post.profile_picture" :src="post.profile_picture" class="avatar-img" />
        <div v-else class="avatar">{{ getInitials(post.username) }}</div>
        
        <div class="user-info">
          <p class="username">{{ post.username }}</p>
          <p class="timestamp">{{ formatTime(post.created_at) }}</p>
        </div>
      </div>
      
      <button class="more-button">
        <svg class="dots-icon" fill="currentColor" viewBox="0 0 24 24">
          <circle cx="6" cy="12" r="2" />
          <circle cx="12" cy="12" r="2" />
          <circle cx="18" cy="12" r="2" />
        </svg>
      </button>
    </div>

    <div class="post-media" @click="$emit('open-detail', post)" style="cursor: pointer">
      <img 
        :src="post.media_url"
        alt="Post content"
        class="post-image"
      />
    </div>

    <div class="post-actions">
      <div class="actions-left">
        <button 
          @click="$emit('toggle-like', post)"
          class="action-button"
          :title="post.is_liked ? 'Unlike' : 'Like'"
        >
          <img 
            :src="post.is_liked ? '/icons/liked-icon.png' : '/icons/notifications-icon.png'"
            alt="Like"
            class="action-icon"
            :class="{ liked: post.is_liked }"
          />
        </button>
        <button class="action-button" title="Comment" @click="$emit('open-detail', post)">
          <img src="/icons/comment-icon.png" alt="Comment" class="action-icon" />
        </button>
        <button class="action-button" title="Share">
          <img src="/icons/share-icon.png" alt="Share" class="action-icon" />
        </button>
      </div>

      <button 
        @click="toggleSave"
        class="action-button"
        :title="isSaved ? 'Unsave' : 'Save'"
      >
        <img 
          :src="isSaved ? '/icons/saved-icon.png' : '/icons/save-icon.png'"
          alt="Save"
          class="action-icon"
          :class="{ saved: isSaved }"
        />
      </button>
    </div>

    <div class="post-footer">
      <p class="likes-count">
        {{ post.likes_count ? post.likes_count.toLocaleString() : 0 }} likes
      </p>

      <div class="caption-section">
        <p class="caption-text">
          <span class="caption-username">{{ post.username }}</span>
          <span class="caption-content" v-html="parseCaption(post.caption)"></span>
        </p>
      </div>

      <button class="view-comments" @click="$emit('open-detail', post)">
        View all {{ post.comments_count || 0 }} comments
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps } from 'vue';
import { formatDistanceToNow } from 'date-fns'; // Make sure to install: npm install date-fns
import { postsApi } from '../services/apiService';

// 1. Accept Data from Parent (HomePage.vue)
const props = defineProps({
  post: {
    type: Object,
    required: true
  }
});

const emit = defineEmits(['open-detail', 'toggle-like']);

// 2. Local State initialized from Props
const isSaved = ref(false);

const toggleSave = async () => {
  try {
    // Optimistic UI update
    isSaved.value = !isSaved.value;
    
    // Call API
    await postsApi.toggleSavePost(props.post.id);
    
    // Note: If you want to save to a specific collection, you would need to 
    // open a modal here instead of immediately calling the API. 
    // For now, this saves to the default "All Posts" collection.
    
  } catch (error) {
    // Revert on error
    isSaved.value = !isSaved.value;
    console.error("Failed to save post", error);
  }
};

// Helper: Generate initials for avatar placeholder
const getInitials = (username: string) => {
  return username ? username.substring(0, 2).toUpperCase() : 'UN';
};

// Helper: Format Date (e.g., "22h ago")
const formatTime = (dateStr: string) => {
  if (!dateStr) return '';
  try {
    return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
  } catch (e) {
    return '';
  }
};

// Helper: Rich Text (Blue Mentions & Hashtags) - Requirement Page 11
const parseCaption = (text: string) => {
  if (!text) return '';
  // Regex to find @mentions and #hashtags
  return text.replace(/([#@][\w.]+)/g, '<span style="color: rgb(0, 149, 246); cursor: pointer; font-weight: 600;">$1</span>');
};
</script>

<style scoped>
/* Main container */
.post-container {
  width: 100%;
  max-width: 480px;
  background-color: #000000;
  border: 1px solid #000;
  border-radius: 8px;
  margin-bottom: 24px;
  overflow: hidden;
  color: #ffffff;
}

/* Header Section */
.post-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #e0e0e0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #a78bfa 0%, #ec4899 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.username {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.timestamp {
  font-size: 12px;
  color: #65676b;
  margin: 0;
}

.more-button {
  background: none;
  border: none;
  padding: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background-color 0.2s ease;
}

.more-button:hover {
  background-color: #f0f0f0;
}

.dots-icon {
  width: 20px;
  height: 20px;
  color: #000000;
}

/* Media Section */
.post-media {
  width: 100%;
  aspect-ratio: 1;
  background-color: #f0f0f0;
  overflow: hidden;
}

.post-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* Action Bar Section */
.post-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}

.actions-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-button {
  background: none;
  border: none;
  padding: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.action-button:hover {
  background-color: #f0f0f0;
}

.action-button:active {
  transform: scale(0.95);
}

.action-icon {
  width: 24px;
  height: 24px;
  opacity: 1;
  /* transition: all 0.2s ease; */
}

.action-icon.liked,
.action-icon.saved {
  opacity: 1;
}

/* Footer Section */
.post-footer {
  padding: 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.likes-count {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.caption-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.caption-text {
  font-size: 14px;
  color: #fff;
  margin: 0;
  line-height: 1.4;
}

.caption-username {
  font-weight: 600;
  color: #fff;
}

.caption-content {
  color: #fff;
  margin-left: 8px;
}

.view-comments {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  font-size: 12px;
  color: #65676b;
  text-align: left;
  transition: color 0.2s ease;
}

.view-comments:hover {
  color: #fff;
}
</style>
