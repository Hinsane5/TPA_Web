<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { formatDistanceToNow } from 'date-fns' // Import Date Fns
import type { DashboardPage } from '../types'
import Sidebar from './Sidebar.vue'
import SearchPanel from './SearchPanel.vue'

import type { Notification } from '@/types';
import CreatePostOverlay from './CreatePostOverlay.vue'
import MiniMessagesComponent from './MiniMessagesComponent.vue'
import { useAuth } from '../composables/useAuth'
import { useStories } from '../composables/useStories'
import { useNotificationStore } from '@/stores/notificationStore'

const router = useRouter()
const route = useRoute()

const notificationStore = useNotificationStore()
const isSearchOpen = ref(false)
const isNotificationOpen = ref(false)
const isCreateOpen = ref(false)
const notificationCount = computed(() => notificationStore.unreadCount)

const { checkLoginState } = useAuth()
const { fetchStories } = useStories()

const showCreateChoice = ref(false)
const showCreateOverlay = ref(false)
const uploadType = ref<'post' | 'story'>('post')
const selectedFile = ref<File | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)
const { user } = useAuth();

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

const toggleNotifications = () => {
  isNotificationOpen.value = !isNotificationOpen.value;
}

watch(isNotificationOpen, (newValue) => {
  if (newValue && notificationStore.unreadCount > 0 && user.value?.id) {
    notificationStore.markNotificationsAsRead(user.value.id);
  }
});

const formatTimeAgo = (timestamp: string | undefined) => {
  if (!timestamp) return 'just now';
  try {
    return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
  } catch (e) {
    return 'just now';
  }
};

const openCreateModal = () => {
  showCreateChoice.value = true
}

const initiateUpload = (type: 'post' | 'story') => {
  uploadType.value = type
  showCreateChoice.value = false
  showCreateOverlay.value = true;
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

const handleNotificationClick = (notif: Notification) => {
  isNotificationOpen.value = false;

  const type = notif.type ? notif.type.toLowerCase() : "";
  const targetId = notif.sender_id;

  if (['follow', 'mention'].includes(type)) {
    router.push({ 
      name: 'profile', 
      params: { id: targetId } 
    });
  } 

  else if (['like', 'comment'].includes(type)) {
    if (notif.entity_id) {
      router.push({ 
        name: 'post-detail', 
        params: { id: notif.entity_id } 
      });
    } else {
      console.warn("Notification missing entity_id, falling back to profile");
      router.push({ 
        name: 'profile', 
        params: { id: targetId } 
      });
    }
  }
  
  else {
     router.push({ name: 'profile', params: { id: targetId } });
  }
  
  if (!notif.is_read) {
    notificationStore.markAsRead(notif.ID);
  }
};

onMounted(() => {
  if (user.value?.id) {
    console.log("Initializing WebSocket for user:", user.value.id);
    notificationStore.connectWebSocket(user.value.id);
    notificationStore.fetchNotifications(user.value.id);
  } else {
    console.error("No User ID found, cannot connect to WebSocket");
  }
  checkLoginState()
})
</script>

<template>
  <div class="dashboard-layout">
    <Sidebar 
      :current-page="currentPage" 
      :notification-count="notificationCount"
      :is-search-active="isSearchOpen" 
      @toggle-search="handleSearchToggle"
      @navigate="handleNavigation" 
      @logout="handleLogout"
      @open-search="isSearchOpen = true"
      @open-notifications="toggleNotifications" 
      @open-create="openCreateModal"
    />
    
    <main class="main-content">
      <RouterView />
    </main>

    <SearchPanel :is-open="isSearchOpen" @close="isSearchOpen = false" @search="handleSearch" />

    <div v-if="isNotificationOpen" class="notification-drawer">
      <div class="notif-header">
        <h3>Notifications</h3>
        <button class="close-btn" @click="isNotificationOpen = false">×</button>
      </div>
      <div class="notif-content-area">
        <ul class="notification-list">
          <li 
            v-for="notification in notificationStore.notifications" 
            :key="notification.ID" 
            class="notification-item" 
            :class="{ 'unread': !notification.is_read }"
            style="cursor: pointer;"
            @click="handleNotificationClick(notification)"
          >
            <img src="https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png" alt="user" class="notif-avatar">
            
            <div class="notif-text-content">
              <p class="notif-message">
                <span class="username">{{ notification.sender_name || 'Unknown User' }}</span>
                {{ notification.message }}
              </p>
              <span class="notif-time">{{ formatTimeAgo(notification.CreatedAt) }}</span>
            </div>
          </li>
          <li v-if="notificationStore.notifications.length === 0" class="empty-state">
            No notifications yet
          </li>
        </ul>
      </div>
    </div>

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
      ref="fileInput" 
      type="file" 
      accept="image/*,video/*"
      style="display: none"
    >

    <CreatePostOverlay 
       v-if="showCreateOverlay"
       :is-open="showCreateOverlay"
       :is-story-mode="uploadType === 'story'" 
       @close="closeCreateOverlay"
       @post-created="handlePostCreated"
       @story-created="handleStoryCreated"
    />
    
    <MiniMessagesComponent />
  </div>
</template>

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

/* --- NEW NOTIFICATION STYLES --- */
.notification-drawer {
  position: fixed;
  left: 240px; /* Aligns next to sidebar */
  top: 0;
  bottom: 0;
  width: 350px;
  background-color: black;
  border-right: 1px solid #363636;
  z-index: 1000;
  display: flex;
  flex-direction: column;
  animation: slideIn 0.3s ease-out;
  color: white;
}

.notif-header {
  padding: 24px 20px;
  border-bottom: 1px solid #363636;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.notif-header h3 {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
}

.close-btn {
  background: none;
  border: none;
  color: #fff;
  font-size: 24px;
  cursor: pointer;
}

.notif-content-area {
  flex: 1;
  overflow-y: auto;
}

.notification-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.notification-item {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #1a1a1a;
  transition: background 0.2s;
}

.notification-item:hover {
  background-color: #121212;
}

.notif-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  margin-right: 14px;
  object-fit: cover;
}

.notif-text-content {
  display: flex;
  flex-direction: column;
}

.notif-message {
  font-size: 14px;
  margin: 0 0 4px 0;
  line-height: 1.4;
}

.notif-time {
  font-size: 12px;
  color: #a8a8a8;
}

.empty-state {
  padding: 20px;
  text-align: center;
  color: #737373;
}

@keyframes slideIn {
  from { transform: translateX(-100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
/* ------------------------------- */

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
  .notification-drawer {
    left: 64px;
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
  
  .notification-drawer {
    left: 0;
    width: 100%;
  }
}
</style>