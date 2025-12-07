<template>
  <div class="home-container">
    <div class="stories-section">
      <MiniCarouselContainer 
        v-if="!isLoadingStories && storyGroups.length > 0"
        @open-viewer="isStoriesOverlayOpen = true"
      />
      
      <div v-if="isLoadingStories && storyGroups.length === 0" class="story-loading">
        Loading...
      </div>
      
      <div v-else-if="storyGroups.length === 0" class="story-placeholder">
        <p>No stories yet</p>
      </div>
    </div>

    <StoriesCarouselOverlay 
      v-if="isStoriesOverlayOpen"
      :isOpen="isStoriesOverlayOpen"
      @close="closeStoriesOverlay"
    />

    <div class="feed-container">
      <div class="feed-section">
        
        <div v-if="posts.length > 0" class="posts-list">
          <PostComponent 
            v-for="post in posts" 
            :key="post.id" 
            :post="post"
            @open-detail="openOverlay"
            @toggle-like="handleToggleLike"
          />
        </div>

        <div v-else-if="!isLoading && posts.length === 0" class="empty-state">
          <p>No posts yet. Follow users to see their posts!</p>
        </div>

        <div v-if="isLoading" class="skeleton-loader">
          <div class="skeleton-card" v-for="n in 2" :key="n">
            <div class="skeleton-header">
              <div class="skeleton-avatar"></div>
              <div class="skeleton-text"></div>
            </div>
            <div class="skeleton-media"></div>
          </div>
        </div>

        <div ref="scrollTrigger" class="scroll-trigger"></div>
      </div>

      <div class="suggested-section">
        <div class="suggested-header">
          <h3>Suggested for you</h3>
          <a href="#">See all</a>
        </div>

        <div class="suggested-user" v-for="n in 5" :key="`suggested-${n}`">
          <div class="suggested-avatar">👤</div>
          <div class="suggested-info">
            <p class="suggested-name">user{{ n }}</p>
            <p class="suggested-label">Suggested for you</p>
          </div>
          <button class="follow-btn">Follow</button>
        </div>
      </div>
    </div>

    <PostDetailOverlay 
      v-if="selectedPost"
      :isOpen="!!selectedPost" 
      :post="selectedPost" 
      @close="closeOverlay" 
      @comment-added="handleCommentAdded"
      @toggle-like="handleToggleLike"
    />

  </div>
</template>


<script setup lang="ts">
import { ref, onMounted } from "vue";
import PostComponent from './PostComponent.vue';
import { postsApi } from '../services/apiService';
import PostDetailOverlay from './PostDetailOverlay.vue';
import StoriesCarouselOverlay from './StoriesCarouselOverlay.vue'; 
import MiniCarouselContainer from './MiniCarouselContainer.vue';
import { useStories } from '../composables/useStories';

interface Post {
  id: string;
  username: string;
  profile_picture?: string;
  media_url: string;
  media_type: string;
  caption: string;
  likes_count: number;
  comments_count: number;
  created_at: string;
  is_liked: boolean;
}

const posts = ref<Post[]>([]);
const isLoading = ref(false)
const page = ref(0)
const limit = 5
const scrollTrigger = ref<HTMLElement | null>(null);
const isLoadingStories = ref(false);
const selectedPost = ref<Post | null>(null);
const isStoriesOverlayOpen = ref(false);

const { storyGroups, fetchStories } = useStories();

const fetchFeed = async () => {
  if (isLoading.value) return;
  isLoading.value = true;
  try {
    const response = await postsApi.getHomeFeed(5, 0);
    if (response.data.data) {
      posts.value = response.data.data;
    }
  } catch(error){
    console.error("Failed to fetch feed:", error);
  } finally {
    isLoading.value = false;
  }
};

const openOverlay = (post: any) => {
  selectedPost.value = post;
  window.history.pushState({}, '', `/p/${post.id}`);
};

const closeOverlay = () => {
  selectedPost.value = null;
  window.history.pushState({}, '', '/'); 
}; 

// const openStoriesCarousel = (index: number) => {
//   selectGroup(index); 
//   isStoriesOverlayOpen.value = true;
// };

const closeStoriesOverlay = () => {
  isStoriesOverlayOpen.value = false;
  fetchStories(); 
};

const handleCommentAdded = () => {
  if (selectedPost.value) {
    selectedPost.value.comments_count = (selectedPost.value.comments_count || 0) + 1;
    
    const originalPost = posts.value.find(p => p.id === selectedPost.value?.id);
    if (originalPost) {
       originalPost.comments_count = selectedPost.value.comments_count;
    }
  }
};

const handleToggleLike = async (post: Post) => {
  const targetPost = posts.value.find(p => p.id === post.id);
  if (!targetPost) return;

  const wasLiked = targetPost.is_liked;
  targetPost.is_liked = !targetPost.is_liked;
  targetPost.likes_count += (targetPost.is_liked ? 1 : -1);

  try {
    if (targetPost.is_liked) {
      await postsApi.likePost(post.id);
    } else {
      await postsApi.unlikePost(post.id);
    }
  } catch (error) {
    targetPost.is_liked = wasLiked;
    targetPost.likes_count += (targetPost.is_liked ? 1 : -1);
    console.error("Like failed", error);
  }
};

onMounted(() => {
  fetchFeed();

  fetchStories();

  const observer = new IntersectionObserver((entries) => {
    if (entries[0]?.isIntersecting){
      fetchFeed();
    }
  }, {threshold: 1.0});

  if(scrollTrigger.value){
    observer.observe(scrollTrigger.value);
  }

});
</script>

<style scoped>
.home-container {
  width: 100%;
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  color: #fff;
}

.stories-section {
  background: #1a1a1a;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #262626;
}

.stories-carousel {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding: 5px 0;
}

.story-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.story-item:hover {
  opacity: 0.8;
}

.story-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  border: 2px solid #404040;
}

.story-username {
  font-size: 12px;
  color: #a0a0a0;
  margin: 0;
  max-width: 56px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.feed-container {
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 20px;
}

.feed-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.post-item {
  background: #1a1a1a;
  border: 1px solid #262626;
  border-radius: 8px;
  overflow: hidden;
}

.post-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid #262626;
}

.post-author {
  display: flex;
  align-items: center;
  gap: 10px;
}

.author-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.author-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.post-time {
  margin: 0;
  font-size: 12px;
  color: #a0a0a0;
}

.more-btn {
  background: none;
  border: none;
  color: #a0a0a0;
  font-size: 18px;
  cursor: pointer;
}

.post-image {
  aspect-ratio: 1;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
}

.post-actions {
  display: flex;
  gap: 10px;
  padding: 10px 15px;
  border-bottom: 1px solid #262626;
}

.action-icon {
  background: none;
  border: none;
  color: #fff;
  font-size: 20px;
  cursor: pointer;
  transition: opacity 0.2s ease;
  padding: 5px;
}

.action-icon:hover {
  opacity: 0.6;
}

.action-icon.bookmark {
  margin-left: auto;
}

.post-stats {
  padding: 10px 15px;
}

.likes {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.post-caption {
  padding: 0 15px;
}

.post-caption p {
  margin: 0;
  font-size: 14px;
  color: #fff;
}

.view-comments {
  padding: 5px 15px;
  margin: 0;
  font-size: 13px;
  color: #a0a0a0;
  cursor: pointer;
}

.post-input {
  display: flex;
  gap: 10px;
  padding: 10px 15px;
  border-top: 1px solid #262626;
}

.post-input input {
  flex: 1;
  background: transparent;
  border: none;
  color: #fff;
  font-size: 14px;
  outline: none;
}

.post-input input::placeholder {
  color: #808080;
}

.post-input button {
  background: none;
  border: none;
  color: #5b5bff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.suggested-section {
  background: #1a1a1a;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #262626;
  max-height: 600px;
  overflow-y: auto;
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
}

.suggested-header a {
  color: #5b5bff;
  text-decoration: none;
  font-size: 13px;
}

.suggested-user {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 15px;
}

.suggested-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
}

.suggested-info {
  flex: 1;
  min-width: 0;
}

.suggested-name {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
}

.suggested-label {
  margin: 2px 0 0 0;
  font-size: 12px;
  color: #a0a0a0;
}

.follow-btn {
  background: none;
  border: none;
  color: #5b5bff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 100px 20px;
  color: #a0a0a0;
}

/* Scrollbar styling */
.suggested-section::-webkit-scrollbar {
  width: 8px;
}

.suggested-section::-webkit-scrollbar-track {
  background: transparent;
}

.suggested-section::-webkit-scrollbar-thumb {
  background: #404040;
  border-radius: 4px;
}

/* Responsive */
@media (max-width: 1024px) {
  .feed-container {
    grid-template-columns: 1fr;
  }

  .suggested-section {
    display: none;
  }
}

@media (max-width: 768px) {
  .home-container {
    padding: 15px;
  }

  .stories-carousel {
    gap: 8px;
  }

  .story-avatar {
    width: 48px;
    height: 48px;
  }

  .story-username {
    font-size: 11px;
  }
}
</style>
