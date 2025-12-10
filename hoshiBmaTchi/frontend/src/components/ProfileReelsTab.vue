<template>
  <div class="reels-tab-container">
    <div v-if="loading" class="loading-skeleton">
      <div v-for="n in 4" :key="n" class="skeleton-item"></div>
    </div>

    <div v-else-if="reels.length === 0" class="empty-state">
      <div class="empty-icon-wrapper">
        <img src="/icons/reels-icon.png" alt="Reels" class="empty-icon" />
      </div>
      <h2>No Reels Yet</h2>
      <p>Capture and edit short videos.</p>
    </div>

    <div v-else class="reels-grid">
      <div 
        v-for="reel in reels" 
        :key="reel.id" 
        class="reel-item"
        @click="openReel(reel)"
      >
        <div class="media-container">
          <img 
            :src="getThumbnail(reel)" 
            class="reel-thumbnail" 
            loading="lazy" 
            alt="Reel thumbnail" 
          />
          
          <div class="video-indicator">
            <img src="/icons/reels-icon.png" alt="Reel" />
          </div>

          <div class="hover-overlay">
            <div class="stat">
              <img src="/icons/liked-icon.png" alt="Likes" />
              <span>{{ formatCount(reel.likes_count) }}</span>
            </div>
            <div class="stat">
              <img src="/icons/comment-icon.png" alt="Comments" />
              <span>{{ formatCount(reel.comments_count) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { postsApi } from '@/services/apiService';

const props = defineProps<{
  userId: string;
}>();

const emit = defineEmits<{
  (e: 'open-post', post: any): void;
}>();

const reels = ref<any[]>([]);
const loading = ref(false);

const formatCount = (count: number) => {
  if (!count) return '0';
  if (count >= 1000) {
    return (count / 1000).toFixed(1) + 'k';
  }
  return count.toString();
};

const getThumbnail = (post: any) => {
  if (post.media && Array.isArray(post.media) && post.media.length > 0) {
    return post.media[0].media_url.replace("http://minio:9000", "http://localhost:9000");
  }
  return "/placeholder.png";
};

const fetchReels = async () => {
  if (!props.userId) return;
  
  loading.value = true;
  try {
    const response = await postsApi.getUserReels(props.userId);
    reels.value = Array.isArray(response.data) ? response.data : (response.data.data || []);
  } catch (error) {
    console.error("Failed to fetch user reels:", error);
  } finally {
    loading.value = false;
  }
};

const openReel = (reel: any) => {
  emit('open-post', reel);
};

onMounted(() => {
  fetchReels();
});

watch(() => props.userId, () => {
  reels.value = [];
  fetchReels();
});
</script>

<style scoped>
.reels-tab-container {
  width: 100%;
  padding-top: 10px;
}

.reels-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr); 
  gap: 4px;
}

.reel-item {
  aspect-ratio: 9 / 16; 
  background-color: #262626;
  position: relative;
  cursor: pointer;
  overflow: hidden;
}

.media-container {
  width: 100%;
  height: 100%;
  position: relative;
}

.reel-thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-indicator {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 20px;
  height: 20px;
  filter: drop-shadow(0 2px 4px rgba(0,0,0,0.5));
}

.video-indicator img {
  width: 100%;
  height: 100%;
  filter: invert(1);
}

.hover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 15px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.reel-item:hover .hover-overlay {
  opacity: 1;
}

.stat {
  display: flex;
  align-items: center;
  gap: 5px;
  color: white;
  font-weight: 600;
  font-size: 14px;
}

.stat img {
  width: 18px;
  height: 18px;
  filter: invert(1); 
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: #a8a8a8;
}

.empty-icon-wrapper {
  width: 60px;
  height: 60px;
  border: 2px solid #262626;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 15px;
}

.empty-icon {
  width: 30px;
  height: 30px;
  filter: invert(1);
}

h2 {
  font-size: 24px;
  font-weight: 700;
  color: white;
  margin-bottom: 10px;
}

@media (max-width: 768px) {
  .reels-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 480px) {
  .reels-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 1px;
  }
}
</style>