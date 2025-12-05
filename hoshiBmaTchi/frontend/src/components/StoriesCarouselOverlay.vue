<template>
  <div class="stories-overlay">
    <!-- Background overlay -->
    <div class="overlay-backdrop" @click="closeOverlay"></div>

    <!-- Main container -->
    <div class="stories-container">
      <!-- Close button -->
      <button class="close-btn" @click="closeOverlay">✕</button>

      <!-- ProgressBarsContainer component -->
      <ProgressBarsContainer 
        :stories="stories"
        :currentStoryIndex="currentStoryIndex"
      />

      <!-- Header -->
      <div class="stories-header">
        <div class="user-info">
          <div class="avatar">{{ currentStory?.userAvatar }}</div>
          <div class="user-details">
            <div class="username-row">
              <span class="username">{{ currentStory?.username }}</span>
              <span v-if="currentStory?.isVerified" class="verified-badge">✓</span>
            </div>
            <span class="time">{{ formatTime(currentStory?.timestamp) }}</span>
          </div>
        </div>
        <div class="header-actions">
          <button class="action-btn" title="Mute">🔇</button>
          <button class="action-btn" title="Play">▶️</button>
          <button class="action-btn" title="More">⋯</button>
        </div>
      </div>

      <!-- Story content -->
      <div class="story-content">
        <img 
          v-if="currentStory" 
          :src="currentStory.imageUrl" 
          :alt="`Story by ${currentStory.username}`"
          class="story-image"
        />
      </div>

      <!-- Navigation arrows -->
      <button 
        v-if="currentStoryIndex > 0"
        class="nav-btn prev-btn" 
        @click="previousStory"
        title="Previous story"
      >
        ‹
      </button>

      <button 
        v-if="currentStoryIndex < stories.length - 1"
        class="nav-btn next-btn" 
        @click="nextStory"
        title="Next story"
      >
        ›
      </button>

      <!-- Bottom actions bar -->
      <div class="bottom-actions">
        <div class="reply-section">
          <span class="reply-text">Reply to {{ currentStory?.username }}</span>
        </div>
        <div class="action-buttons">
          <button 
            class="like-btn"
            :class="{ liked: currentStory?.isLiked }"
            @click="toggleLike"
            title="Like"
          >
            ❤️
          </button>
          <button 
            class="share-btn"
            @click="showShareModal = true"
            title="Share"
          >
            ✈️
          </button>
        </div>
      </div>

      <!-- Reply input -->
      <div class="reply-input-container">
        <input 
          v-model="storyReplyText"
          type="text" 
          placeholder="Reply..."
          class="reply-input"
          @keyup.enter="addReply"
        />
      </div>

      <!-- MiniCarouselContainer component -->
      <MiniCarouselContainer 
        :stories="stories"
        :currentStoryIndex="currentStoryIndex"
        @select-story="currentStoryIndex = $event"
      />
    </div>

    <!-- Share modal -->
    <ShareStoryModal 
      v-if="showShareModal"
      :story="currentStory"
      @close="showShareModal = false"
      @send="sendStory"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useStories } from '../composables/useStories';
import ShareStoryModal from './ShareStoryModal.vue';
import ProgressBarsContainer from './ProgressBarsContainer.vue';
import MiniCarouselContainer from './MiniCarouselContainer.vue';

interface Props {
  isOpen: boolean;
  storyId?: string;
}

const emit = defineEmits<{
  close: [];
}>();

const showShareModal = ref(false);

const {
  stories,
  currentStoryIndex,
  currentStory,
  storyReplyText,
  toggleLike,
  nextStory,
  previousStory,
  addReply,
} = useStories();

const closeOverlay = () => {
  emit('close');
};

const formatTime = (date: Date | undefined) => {
  if (!date) return '';
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 60) return `${minutes}m ago`;
  if (hours < 24) return `${hours}h ago`;
  return `${days}d ago`;
};

const sendStory = () => {
  console.log('Story sent');
  showShareModal.value = false;
};
</script>

<style scoped>
.stories-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.overlay-backdrop {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  cursor: pointer;
}

.stories-container {
  position: relative;
  z-index: 10000;
  width: 100%;
  max-width: 400px;
  aspect-ratio: 9/16;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.9);
}

.close-btn {
  position: absolute;
  top: 12px;
  left: 12px;
  width: 32px;
  height: 32px;
  background: rgba(0, 0, 0, 0.6);
  border: none;
  border-radius: 50%;
  color: #fff;
  font-size: 18px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10001;
  transition: background 0.2s ease;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.8);
}

.stories-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.4);
  position: relative;
  z-index: 10001;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.username-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

.username {
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}

.verified-badge {
  color: #0095f6;
  font-size: 12px;
}

.time {
  color: #a0a0a0;
  font-size: 12px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  background: none;
  border: none;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  padding: 6px;
  transition: opacity 0.2s ease;
}

.action-btn:hover {
  opacity: 0.7;
}

.story-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
  overflow: hidden;
  position: relative;
}

.story-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0, 0, 0, 0.3);
  border: none;
  color: #fff;
  font-size: 32px;
  width: 48px;
  height: 48px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10002;
  transition: background 0.2s ease;
}

.nav-btn:hover {
  background: rgba(0, 0, 0, 0.5);
}

.prev-btn {
  left: 8px;
}

.next-btn {
  right: 8px;
}

.bottom-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.6);
  position: relative;
  z-index: 10001;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.reply-section {
  flex: 1;
}

.reply-text {
  color: #a0a0a0;
  font-size: 13px;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 12px;
}

.like-btn,
.share-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 6px;
  transition: transform 0.2s ease;
}

.like-btn:hover,
.share-btn:hover {
  transform: scale(1.1);
}

.like-btn.liked {
  transform: scale(1.15);
}

.reply-input-container {
  padding: 10px 16px;
  background: rgba(0, 0, 0, 0.6);
  position: relative;
  z-index: 10001;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.reply-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  padding: 8px 16px;
  color: #fff;
  font-size: 13px;
  outline: none;
  transition: background 0.2s ease;
}

.reply-input::placeholder {
  color: #808080;
}

.reply-input:focus {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.3);
}
</style>
