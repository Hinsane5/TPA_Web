<template>
  <div class="dashboard-layout">
    <!-- Sidebar Navigation -->
    <Sidebar 
      :current-page="currentPage" 
      :notification-count="notificationCount"
      :is-search-active="isSearchOpen" 
      @toggle-search="handleSearchToggle"
      @navigate="handleNavigation" 
      @logout="handleLogout"
      @open-search="isSearchOpen = true"
      @open-notifications="isNotificationOpen = true"
      @open-create="openCreateModal"
    />
    
    <!-- Main Content Area -->
    <main class="main-content">
      <RouterView />
    </main>

    <!-- Panels & Modals -->
    <SearchPanel :is-open="isSearchOpen" @close="isSearchOpen = false" @search="handleSearch" />
    <NotificationPanel :is-open="isNotificationOpen" @close="isNotificationOpen = false" />
    <div v-if="showCreateChoice" class="choice-modal-overlay" @click.self="showCreateChoice = false">
      <div class="choice-modal-content">
        <h3>Create</h3>
        <div class="choice-list">
          <button class="choice-item" @click="initiateUpload('post')">
            <span class="choice-icon">
              <img src="/icons/post-icon.png" alt="Post" class="choice-img-icon" />
            </span>
            <span>Post</span>
          </button>
          <button class="choice-item" @click="initiateUpload('story')">
             <span class="choice-icon">
               <img src="/icons/instagram-icon.png" alt="Story" class="choice-img-icon" />
             </span> 
            <span>Story</span>
          </button>
        </div>
      </div>
    </div>
    <input 
      type="file" 
      ref="fileInput" 
      accept="image/*,video/*"
      style="display: none"
    >

    <CreatePostOverlay 
       v-if="showCreateOverlay"
       :isOpen="showCreateOverlay"
       :isStoryMode="uploadType === 'story'" 
       @close="closeCreateOverlay"
       @post-created="handlePostCreated"
       @story-created="handleStoryCreated"

    />
    
    <!-- Mini Messages Component (appears on all pages) -->
    <MiniMessagesComponent />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { DashboardPage } from '../types'
import Sidebar from './Sidebar.vue'
import SearchPanel from './SearchPanel.vue'
import NotificationPanel from './NotificationPanel.vue'
import CreatePostOverlay from './CreatePostOverlay.vue'
import MiniMessagesComponent from './MiniMessagesComponent.vue'
import { useAuth } from '../composables/useAuth'
import { useStories } from '../composables/useStories'

const router = useRouter()
const route = useRoute()

const isSearchOpen = ref(false)
const isNotificationOpen = ref(false)
const isCreateOpen = ref(false)
const notificationCount = ref(0)

const { checkLoginState } = useAuth()
const { fetchStories } = useStories()

const showCreateChoice = ref(false)
const showCreateOverlay = ref(false)
const uploadType = ref<'post' | 'story'>('post')
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

const currentPage = computed(() => {
  const pageName = route.path.split('/').pop()
  return (pageName as DashboardPage) || 'home'
})

const handleNavigation = (page: DashboardPage) => {
  router.push({ name: page })
}

const handleLogout = () => {
  router.push({ name: 'login' })
}

const handleSearch = (query: string) => {
  console.log('Search for:', query)
}

const handleSearchToggle = () => {
  isSearchOpen.value = !isSearchOpen.value;
};

const openCreateModal = () => {
  showCreateChoice.value = true
}

const initiateUpload = (type: 'post' | 'story') => {
  uploadType.value = type
  showCreateChoice.value = false
  
  showCreateOverlay.value = true;
}

const handlePostUpload = (file: File, description: string) => {
  console.log('Upload post:', file.name, 'Description:', description)
  isCreateOpen.value = false
}

const closeCreateOverlay = () => {
  showCreateOverlay.value = false
  selectedFile.value = null
  uploadType.value = 'post'
}

const handlePostCreated = () => {
  console.log('Post created successfully')
}

const handleStoryCreated = () => {
  console.log('Story created successfully')
  fetchStories()
}

onMounted(() => {
  checkLoginState()
})
</script>

<style scoped>
.dashboard-layout {
  display: flex;
  width: 100%;
  height: 100vh;
  background-color: var(--background-dark);
}

.main-content {
  flex: 1;
  margin-left: 240px;
  overflow-y: auto;
  background-color: var(--background-dark);
}

.choice-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0.65); 
  z-index: 20000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.choice-modal-content {
  background: #262626;
  border-radius: 12px;
  width: 300px;
  overflow: hidden;
  color: #fff;
  border: 1px solid #363636;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
}

.choice-modal-content h3 {
  text-align: center;
  padding: 15px;
  margin: 0;
  border-bottom: 1px solid #363636;
  font-size: 16px;
  font-weight: 600;
}

.choice-list {
  display: flex;
  flex-direction: column;
}

.choice-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 15px 20px;
  background: none;
  border: none;
  border-bottom: 1px solid #363636;
  color: #fff;
  font-size: 15px;
  cursor: pointer;
  text-align: left;
  transition: background 0.2s;
}

.choice-item:last-child {
  border-bottom: none;
}

.choice-item:hover {
  background: #303030;
}

.choice-img-icon {
  width: 24px;
  height: 24px;
  object-fit: contain;
}

/* Responsive */
@media (max-width: 1024px) {
  .main-content {
    margin-left: 64px;
  }
}

@media (max-width: 768px) {
  .main-content {
    margin-left: 0;
    margin-bottom: 60px;
  }

  .dashboard-layout {
    flex-direction: column-reverse;
  }
}
</style>
