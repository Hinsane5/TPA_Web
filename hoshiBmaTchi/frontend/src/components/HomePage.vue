<template>
  <div class="home-container">
    <div class="stories-section">
      <div class="stories-carousel">
        <div class="story-item" v-for="n in 8" :key="`story-${n}`">
          <div class="story-avatar">
            <span>👤</span>
          </div>
          <p class="story-username">User {{ n }}</p>
        </div>
      </div>
    </div>

    <div class="feed-container">
      <div class="feed-section">
        <div class="post-item" v-for="n in 5" :key="`post-${n}`">
          <div class="post-header">
            <div class="post-author">
              <div class="author-avatar">👤</div>
              <div class="author-info">
                <p class="author-name">Username {{ n }}</p>
                <p class="post-time">2 hours ago</p>
              </div>
            </div>
            <button class="more-btn">⋯</button>
          </div>

          <div class="post-image">
            <span>📸</span>
          </div>

          <div class="post-actions">
            <button class="action-icon">❤️</button>
            <button class="action-icon">💬</button>
            <button class="action-icon">📤</button>
            <button class="action-icon bookmark">🔖</button>
          </div>

          <div class="post-stats">
            <p class="likes">100 likes</p>
          </div>

          <div class="post-caption">
            <p>
              <strong>username{{ n }}</strong> Amazing moment with friends!
              #nofilter
            </p>
          </div>

          <p class="view-comments">View all 20 comments</p>

          <div class="post-input">
            <input type="text" placeholder="Add a comment..." />
            <button>Post</button>
          </div>
        </div>

        <div v-if="!hasPosts" class="empty-state">
          <p>No posts yet. Follow users to see their posts!</p>
        </div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";

const hasPosts = ref(false);

onMounted(() => {
  // TODO: Fetch feed posts from backend
  // TODO: Fetch suggested users from backend
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
