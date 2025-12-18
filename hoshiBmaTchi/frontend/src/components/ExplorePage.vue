<template>
  <div class="explore-container">
    <div class="explore-header" v-if="searchQuery">
      <h2>{{ searchQuery.startsWith('#') ? searchQuery : '#' + searchQuery }}</h2>
    </div>

    <div class="explore-grid" ref="scrollContainer">
      <div 
        v-for="(post, index) in posts" 
        :key="post.id" 
        class="explore-item"
        :class="{ 'big-item': isBigItem(index) }"
        @click="openPostDetail(post)"
      >
        <div v-if="post.media && post.media.length > 0 && post.media[0].media_type.startsWith('image')" class="media-container">
          <img 
            :src="post.media[0].media_url" 
            loading="lazy"
            class="media-content"
            @error="handleImageError"
          />
        </div>

        <div v-else-if="post.media && post.media.length > 0" class="media-container">
          <video 
            :src="post.media[0].media_url" 
            class="media-content"
            autoplay
            muted
            loop
            playsinline
          ></video>
          <div class="reel-icon">
            <img src="/icons/reels-icon.png" alt="Reel" />
          </div>
        </div>

        <div v-else class="media-container text-only">
          <p class="caption-preview">{{ post.caption }}</p>
        </div>

        <div v-if="post.media && post.media.length > 1" class="gallery-icon">
          <img src="/icons/gallery-icon.png" alt="Gallery" />
        </div>

        <div class="item-overlay">
          <div class="item-stats">
            <div class="stat">
               <img src="/icons/liked-icon.png" class="stat-icon" />
               <span>{{ formatNumber(post.likes_count) }}</span>
            </div>
            <div class="stat">
               <img src="/icons/comment-icon.png" class="stat-icon" />
               <span>{{ formatNumber(post.comments_count) }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <template v-if="loading">
        <div v-for="n in 6" :key="`skeleton-${n}`" class="explore-item skeleton"></div>
      </template>

      <div v-if="!loading && posts.length === 0" class="empty-state">
        <p>No posts found.</p>
      </div>
    </div>

    <div ref="observerTarget" class="observer-target"></div>

    <PostDetailOverlay 
      v-if="selectedPost" 
      :is-open="true"
      :post="selectedPost" 
      @close="closePostDetail" 
      @toggle-like="handleLike"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRoute } from 'vue-router'
import { postsApi } from '@/services/apiService'
import PostDetailOverlay from '@/components/PostDetailOverlay.vue'

const route = useRoute()
const posts = ref<any[]>([])
const loading = ref(false)
const offset = ref(0)
const limit = 15
const hasMore = ref(true)
const selectedPost = ref<any>(null)
const observerTarget = ref<HTMLElement | null>(null)

const searchQuery = computed(() => route.query.q as string || '')

const fetchPosts = async (reset = false) => {
  if (loading.value || (!hasMore.value && !reset)) return
  
  loading.value = true
  if (reset) {
    posts.value = []
    offset.value = 0
    hasMore.value = true
  }

  try {
    const rawQuery = searchQuery.value || '';
    const cleanQuery = rawQuery.startsWith('#') ? rawQuery.slice(1) : rawQuery;

    const res = await postsApi.getExplorePosts(limit, offset.value, cleanQuery)
    
    const newPosts = res.data?.data || []

    if (newPosts.length > 0) {
      posts.value.push(...newPosts)
      offset.value += limit
    } else {
      hasMore.value = false
    }
  } catch (error) {
    console.error('Failed to fetch explore posts:', error)
  } finally {
    loading.value = false
  }
}

const isBigItem = (index: number) => {
  const patternIndex = index % 10
  return patternIndex === 2 || patternIndex === 9 
}

const openPostDetail = (post: any) => {
  selectedPost.value = post
}

const closePostDetail = () => {
  selectedPost.value = null
}

const handleLike = async (post: any) => {
  try {
    if (post.is_liked) {
      await postsApi.unlikePost(post.id);
      post.likes_count--;
      post.is_liked = false;
    } else {
      await postsApi.likePost(post.id);
      post.likes_count++;
      post.is_liked = true;
    }
  } catch (error) {
    console.error("Failed to toggle like", error);
  }
}

const formatNumber = (num: number) => {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}

const handleImageError = (e: Event) => {
  (e.target as HTMLImageElement).style.display = 'none';
}

watch(() => route.query.q, () => {
  fetchPosts(true)
})

onMounted(() => {
  fetchPosts(true)
  
  const observer = new IntersectionObserver((entries) => {
    if (entries[0] && entries[0].isIntersecting) {
      fetchPosts()
    }
  }, { rootMargin: '200px' })
  
  if (observerTarget.value) {
    observer.observe(observerTarget.value)
  }
})
</script>

<style scoped>
.explore-container {
  width: 100%;
  max-width: 935px;
  margin: 0 auto;
  padding: 20px 0;
}

.explore-header {
  padding: 0 20px 20px;
  color: white;
}

.explore-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
  grid-auto-flow: dense;
}

.explore-item {
  position: relative;
  aspect-ratio: 1 / 1;
  background: #262626;
  cursor: pointer;
  overflow: hidden;
  border: 1px solid #363636;
}

.explore-item.big-item {
  grid-column: span 2;
  grid-row: span 2;
}

.media-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.media-content {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.text-only {
  background: #000;
  padding: 10px;
  text-align: center;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
}

.caption-preview {
  font-size: 14px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 4; /* Limit lines */
  -webkit-box-orient: vertical;
}

.gallery-icon, .reel-icon {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 20px;
  height: 20px;
  filter: drop-shadow(0 0 2px rgba(0,0,0,0.5));
  z-index: 2;
}

.item-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transition: opacity 0.2s ease;
  z-index: 3;
}

.explore-item:hover .item-overlay {
  opacity: 1;
}

.item-stats {
  display: flex;
  gap: 20px;
  color: white;
  font-weight: bold;
}

.stat {
  display: flex;
  align-items: center;
  gap: 5px;
}

.stat-icon {
  width: 20px;
  height: 20px;
  filter: invert(1);
}

.skeleton {
  background: linear-gradient(90deg, #262626 25%, #333 50%, #262626 75%);
  background-size: 200% 100%;
  animation: loading 1.5s infinite;
}

.empty-state {
    color: #8e8e8e;
    text-align: center;
    grid-column: 1 / -1;
    padding: 40px;
}

.observer-target {
  height: 20px;
  margin-top: 20px;
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

@media (max-width: 768px) {
  .explore-grid {
    gap: 1px;
  }
}
</style>