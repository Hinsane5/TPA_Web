<template>
  <div class="archive-page">
    <div class="archive-header">
      <button class="back-btn" @click="goBack">
        <span class="back-arrow">←</span>
      </button>
      <div class="header-title">Archive</div>
      <div class="header-placeholder"></div> </div>

    <div class="archive-tabs">
      <div class="tab-item active">STORIES</div>
    </div>

    <div class="stories-grid">
      <div 
        v-for="(story, index) in archivedStories" 
        :key="story.id" 
        class="story-grid-item"
        @click="openStory(index)"
      >
        <div class="media-container">
          <img 
            v-if="story.mediaType === 'image'" 
            :src="getSafeImageUrl(story.mediaUrl)" 
            class="story-thumbnail" 
            loading="lazy"
          />
          <video 
            v-else 
            :src="getSafeImageUrl(story.mediaUrl)" 
            class="story-thumbnail"
            preload="metadata"
          ></video>
        </div>
        <div class="date-overlay">
          {{ formatDate(story.timestamp) }}
        </div>
      </div>

      <template v-if="isLoading">
        <div v-for="n in 8" :key="`skeleton-${n}`" class="story-grid-item skeleton"></div>
      </template>
    </div>

    <div ref="scrollTrigger" class="scroll-trigger"></div>

    <StoriesCarouselOverlay 
      v-if="showOverlay" 
      :isOpen="showOverlay"
      @close="closeOverlay"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import { storiesApi } from '../services/apiService';
import { useStories } from '../composables/useStories';
import StoriesCarouselOverlay from './StoriesCarouselOverlay.vue';
import type { Story, StoryGroup } from '../types/stories';

const router = useRouter();
const { storyGroups, currentGroupIndex, currentStoryIndex, isPlaying } = useStories();

// State
const archivedStories = ref<Story[]>([]);
const isLoading = ref(false);
const limit = 12;
const offset = ref(0);
const hasMore = ref(true);
const showOverlay = ref(false);
const scrollTrigger = ref<HTMLElement | null>(null);
let observer: IntersectionObserver | null = null;

const goBack = () => {
  router.back();
};

const getSafeImageUrl = (url: string | undefined) => {
  if (!url) return '';
  if (url.includes('minio:9000') || url.includes('backend:9000')) {
    return url.replace('minio:9000', 'localhost:9000').replace('backend:9000', 'localhost:9000');
  }
  return url;
};

const formatDate = (date: Date) => {
  return new Date(date).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short'
  });
};

const fetchArchivedStories = async () => {
  if (isLoading.value || !hasMore.value) return;
  
  isLoading.value = true;
  try {
    const response = await storiesApi.getArchivedStories(limit, offset.value);
    const newStories = response.data.stories || [];

    if (newStories.length < limit) {
      hasMore.value = false;
    }

    // Map to Story interface
    const mappedStories: Story[] = newStories.map((s: any) => {
      // 1. SAFE MEDIA TYPE CONVERSION
      let mType = 'image';
      if (typeof s.media_type === 'string') {
        mType = s.media_type.toLowerCase();
      } else if (typeof s.media_type === 'number') {
        // Handle Enum: 0 = IMAGE, 1 = VIDEO
        mType = s.media_type === 1 ? 'video' : 'image';
      }

      // 2. SAFE DATE PARSING (Fixes "Invalid Date")
      let timestamp = new Date();
      if (s.created_at) {
        if (typeof s.created_at === 'string') {
           // Handle ISO string from fixed backend
           timestamp = new Date(s.created_at);
        } else if (s.created_at.seconds) {
           // Handle raw Protobuf object from unfixed backend
           timestamp = new Date(Number(s.created_at.seconds) * 1000);
        }
      }

      return {
        id: s.id,
        mediaType: mType,
        mediaUrl: s.media_url,
        isViewed: true, 
        isLiked: s.is_liked || false,
        likes: s.likes_count || 0,
        timestamp: timestamp,
        replies: [],
        userId: 'me',
        username: 'Me',
        userAvatar: '',
        isVerified: false,
        user: {
          id: 'me',
          username: 'Me',
          fullName: '',
          userAvatar: ''
        }
      };
    });

    archivedStories.value.push(...mappedStories);
    offset.value += limit;
  } catch (error) {
    console.error('Failed to fetch archived stories:', error);
  } finally {
    isLoading.value = false;
  }
};

const openStory = (index: number) => {
  // Construct a temporary group for the Archive view
  const archiveGroup: StoryGroup = {
    userId: 'archive',
    username: 'Archive',
    userAvatar: '',
    isVerified: false,
    stories: archivedStories.value,
    hasUnseen: false
  };

  // Override global store state to show these stories
  storyGroups.value = [archiveGroup];
  currentGroupIndex.value = 0;
  currentStoryIndex.value = index;
  isPlaying.value = true;
  
  showOverlay.value = true;
};

const closeOverlay = () => {
  showOverlay.value = false;
  isPlaying.value = false;
};

onMounted(() => {
  fetchArchivedStories();

  observer = new IntersectionObserver((entries) => {
    // FIX: Safely access the first element
    const entry = entries[0];
    if (entry && entry.isIntersecting) {
      fetchArchivedStories();
    }
  }, { threshold: 0.5 });

  if (scrollTrigger.value) {
    observer.observe(scrollTrigger.value);
  }
});

onUnmounted(() => {
  if (observer) observer.disconnect();
});
</script>

<style scoped>
.archive-page {
  width: 100%;
  min-height: 100vh;
  background-color: var(--background-dark);
  color: var(--text-primary);
  display: flex;
  flex-direction: column;
}

.archive-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border-color);
  background-color: var(--background-dark);
  position: sticky;
  top: 0;
  z-index: 10;
}

.back-btn {
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 24px;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.header-title {
  font-weight: 600;
  font-size: 16px;
}

.header-placeholder {
  width: 24px;
}

.archive-tabs {
  display: flex;
  justify-content: center;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 2px;
}

.tab-item {
  padding: 12px 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--text-primary);
  letter-spacing: 1px;
}

.stories-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 2px;
  padding-bottom: 20px;
}

.story-grid-item {
  position: relative;
  aspect-ratio: 9/16;
  cursor: pointer;
  background-color: #262626;
  overflow: hidden;
}

.media-container {
  width: 100%;
  height: 100%;
}

.story-thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.2s ease;
}

.story-grid-item:hover .story-thumbnail {
  opacity: 0.9;
}

.date-overlay {
  position: absolute;
  top: 8px;
  left: 8px;
  background: rgba(255, 255, 255, 0.8);
  color: #000;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 4px;
  pointer-events: none;
}

.skeleton {
  background-color: #262626;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% { opacity: 1; }
  50% { opacity: 0.5; }
  100% { opacity: 1; }
}

.scroll-trigger {
  height: 20px;
  width: 100%;
}
</style>