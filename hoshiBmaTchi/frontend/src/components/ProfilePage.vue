<template>
  <div class="profile-container">
    <!-- Profile Header -->
    <div class="profile-header">
      <div class="profile-info">
        <div class="profile-picture-wrapper" @click="showProfileImageModal = true">
          <img 
            :src="currentUser?.profileImage || '/placeholder.svg?height=150&width=150'" 
            :alt="currentUser?.fullName || 'Profile'"
            class="profile-picture"
          />
        </div>
        
        <div class="profile-details">
          <div class="profile-top">
            <div class="user-info">
              <h1 class="full-name">{{ currentUser?.fullName || 'Loading...' }}</h1>
              <p class="username">@{{ currentUser?.username || 'username' }}</p>
            </div>
            
            <div class="profile-actions">
              <button class="action-btn">Edit profile</button>
              <button class="action-btn">View archive</button>
              <!-- Replaced settings emoji with icon image -->
              <button class="action-btn settings-btn" title="Settings">
                <img src="/icons/setting-icon.png" alt="Settings" class="settings-icon" />
              </button>
            </div>
          </div>
          
          <p class="bio">{{ currentUser?.bio || 'No bio' }}</p>
          
          <div class="stats">
            <div class="stat">
              <span class="stat-number">{{ currentUser?.postsCount || 0 }}</span>
              <span class="stat-label">posts</span>
            </div>
            <div class="stat">
              <span class="stat-number">{{ currentUser?.followers || 0 }}</span>
              <span class="stat-label">followers</span>
            </div>
            <div class="stat">
              <span class="stat-number">{{ currentUser?.following || 0 }}</span>
              <span class="stat-label">following</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Profile Tabs -->
    <div class="profile-tabs">
      <button 
        v-for="tab in tabs"
        :key="tab"
        :class="['tab', { active: activeTab === tab }]"
        @click="activeTab = tab"
      >
        <!-- Replaced tab emoji icons with icon images -->
        <img :src="getTabIconPath(tab)" :alt="tab" class="tab-icon" />
        {{ tab.charAt(0).toUpperCase() + tab.slice(1) }}
      </button>
    </div>

    <!-- Tab Content -->
    <div class="tab-content">
      <!-- Posts Tab -->
      <div v-if="activeTab === 'posts'" class="posts-grid">
        <!-- Posts will be populated from backend via prop/API -->
        <div 
          class="grid-item" 
          v-for="post in posts" 
          :key="post.id"
          @click="openPostDetail(post)" 
        >
          <img :src="post.mediaUrl" :alt="post.caption" class="post-image" loading="lazy" />
        </div>

        <div v-if="posts.length === 0" class="empty-state">
          <p>No posts yet. Start sharing your content!</p>
        </div>
      </div>

      <!-- Reels Tab -->
      <div v-if="activeTab === 'reels'" class="reels-grid">
        <!-- Reels will be populated from backend via prop/API -->
        <div class="grid-item reel-placeholder" v-for="n in 4" :key="`reel-${n}`">
          <div class="placeholder-video">
            <span>🎬</span>
          </div>
        </div>
        <div v-if="!hasContent" class="empty-state">
          <p>No reels yet. Create your first reel!</p>
        </div>
      </div>

      <!-- Saved Tab -->
      <div v-if="activeTab === 'saved'" class="saved-grid">
        <!-- Collections will be populated from backend via prop/API -->
        <div class="grid-item collection-placeholder" v-for="n in 3" :key="`saved-${n}`">
          <div class="placeholder-collection">
            <span>📁</span>
            <p>Collection {{ n }}</p>
          </div>
        </div>
        <div v-if="!hasContent" class="empty-state">
          <p>No saved collections yet.</p>
        </div>
      </div>

      <!-- Mentions Tab -->
      <div v-if="activeTab === 'mentions'" class="mentions-grid">
        <!-- Mentions will be populated from backend via prop/API -->
        <div class="grid-item mention-placeholder" v-for="n in 6" :key="`mention-${n}`">
          <div class="placeholder-image">
            <span>💬</span>
          </div>
        </div>
        <div v-if="!hasContent" class="empty-state">
          <p>No mentions yet.</p>
        </div>
      </div>
    </div>

    <!-- Profile Image Modal -->
    <div v-if="showProfileImageModal" class="modal-overlay" @click="showProfileImageModal = false">
      <div class="modal-content" @click.stop>
        <button class="close-btn" @click="showProfileImageModal = false">✕</button>
        <img 
          :src="currentUser?.profileImage || '/placeholder.svg?height=400&width=400'" 
          :alt="currentUser?.fullName || 'Profile'"
          class="modal-image"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { User } from '../types'
import axios from 'axios'

interface Post {
  id: string
  mediaUrl: string
  caption: string
}

const currentUser = ref ({
  fullName: 'John Doe',
  username: 'XXXXXXX',
  bio: 'I love photography!',
  postsCount: 10,
  followers: 100,
  following: 50,
  profileImage: '',
})

const posts = ref<Post[]>([])
const activeTab = ref<'posts' | 'reels' | 'saved' | 'mentions'>('posts')
const showProfileImageModal = ref(false)
const hasContent = ref(false)
const tabs = ['posts', 'reels', 'saved', 'mentions'] as const

const getUserIdFromToken = (token: string): string | null => {
  try{
    const parts = token.split('.')
    
    // FIX: Check if we actually got 3 parts (Header.Payload.Signature)
    if (parts.length < 2 || !parts[1]) {
        console.error("Invalid token format")
        return null
    }

    const base64Url = parts[1]
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/')
    const jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
    }).join(''))

    const payload = JSON.parse(jsonPayload)
    return payload.user_id || payload.sub || payload.id
  } catch (e){
    console.error('Error decoding token:', e)
    return null
  }
}

const getTabIconPath = (tab: string): string => {
  const icons: Record<string, string> = {
    posts: '/icons/post-icon.png',
    reels: '/icons/reels-icon.png',
    saved: '/icons/save-icon.png',
    mentions: '/icons/mention-icon.png',
  }
  return icons[tab] || ''
}

const openPostDetail = (post: Post) => {
  console.log("Opening post detail for:", post.id)
}

const fetchUserPosts = async () => {
  try {
    const accessToken = localStorage.getItem('accessToken')
    if (!accessToken){
      console.error("No access token found")
      return
    }

    const userId = getUserIdFromToken(accessToken)

    if(!userId){
      console.error("Could not extract User ID from token")
      return
    }

    console.log("Fetching posts for User ID:", userId)

    const response = await axios.get(`/api/v1/posts/user/${userId}`, {
       headers: { Authorization: `Bearer ${accessToken}` }
    })

    posts.value = response.data || []
    currentUser.value.postsCount = posts.value.length
    hasContent.value = posts.value.length > 0
  } catch (error){
    console.error("Failed to fetch posts:", error)
  }
}

onMounted(() => {
  fetchUserPosts()
})
</script>

<style scoped>
.profile-container {
  width: 100%;
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
  color: #fff;
}

.profile-header {
  border-bottom: 1px solid #262626;
  padding-bottom: 40px;
  margin-bottom: 30px;
}

.profile-info {
  display: flex;
  gap: 40px;
  align-items: flex-start;
}

.profile-picture-wrapper {
  cursor: pointer;
  flex-shrink: 0;
}

.profile-picture {
  width: 150px;
  height: 150px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #262626;
  transition: transform 0.2s ease;
}

.profile-picture-wrapper:hover .profile-picture {
  transform: scale(1.05);
}

.profile-details {
  flex: 1;
  min-width: 0;
}

.profile-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  gap: 20px;
}

.user-info h1 {
  font-size: 28px;
  font-weight: 300;
  margin: 0 0 5px 0;
}

.username {
  color: #a0a0a0;
  font-size: 16px;
  margin: 0;
}

.profile-actions {
  display: flex;
  gap: 10px;
}

.action-btn {
  padding: 8px 16px;
  background: #262626;
  border: 1px solid #404040;
  color: #fff;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: #383838;
}

.settings-btn {
  padding: 8px 12px;
  min-width: auto;
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
  filter: brightness(0.9);
}

.bio {
  color: #e0e0e0;
  margin: 0 0 20px 0;
  font-size: 14px;
  line-height: 1.4;
}

.stats {
  display: flex;
  gap: 40px;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-number {
  font-size: 20px;
  font-weight: 600;
}

.stat-label {
  color: #a0a0a0;
  font-size: 12px;
  margin-top: 5px;
}

.profile-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid #262626;
  margin-bottom: 20px;
  overflow-x: auto;
}

.tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 15px 20px;
  background: none;
  border: none;
  color: #a0a0a0;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  border-bottom: 2px solid transparent;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.tab:hover {
  color: #fff;
}

.tab.active {
  color: #fff;
  border-bottom-color: #fff;
}

.tab-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
  filter: brightness(0.9);
}

.tab.active .tab-icon {
  filter: brightness(1.2);
}

.tab-content {
  min-height: 400px;
}

.posts-grid,
.reels-grid,
.saved-grid,
.mentions-grid {
  display: grid;
  gap: 20px;
  width: 100%;
}

.posts-grid,
.mentions-grid {
  grid-template-columns: repeat(3, 1fr);
}

.reels-grid {
  grid-template-columns: repeat(4, 1fr);
}

.saved-grid {
  grid-template-columns: repeat(3, 1fr);
}

.grid-item {
  aspect-ratio: 1;
  background: #262626;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
}

.grid-item:hover {
  transform: scale(1.02);
}

.placeholder-image,
.placeholder-video,
.placeholder-collection {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 10px;
  font-size: 48px;
  color: #808080;
}

.placeholder-collection p {
  font-size: 14px;
  color: #a0a0a0;
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
  color: #a0a0a0;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  position: relative;
  max-width: 90%;
  max-height: 90vh;
}

.close-btn {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(0, 0, 0, 0.6);
  border: none;
  color: #fff;
  font-size: 24px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
  z-index: 10;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.8);
}

.modal-image {
  width: 100%;
  max-width: 500px;
  height: auto;
  border-radius: 8px;
  object-fit: contain;
}

/* Responsive Design */
@media (max-width: 768px) {
  .profile-info {
    flex-direction: column;
    align-items: center;
    gap: 20px;
    text-align: center;
  }

  .profile-picture {
    width: 120px;
    height: 120px;
  }

  .profile-top {
    flex-direction: column;
    align-items: center;
  }

  .profile-actions {
    width: 100%;
    justify-content: center;
  }

  .action-btn {
    flex: 1;
  }

  .stats {
    justify-content: center;
  }

  .posts-grid,
  .mentions-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .reels-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .profile-container {
    padding: 15px;
  }

  .profile-picture {
    width: 100px;
    height: 100px;
  }

  .user-info h1 {
    font-size: 22px;
  }

  .stats {
    gap: 20px;
  }

  .stat-number {
    font-size: 16px;
  }

  .posts-grid,
  .mentions-grid,
  .reels-grid,
  .saved-grid {
    grid-template-columns: 1fr;
  }
}
</style>
