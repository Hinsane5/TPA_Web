<template>
  <div class="dashboard-layout">
    <!-- Sidebar Navigation -->
    <Sidebar 
      :current-page="currentPage" 
      :notification-count="notificationCount"
      @navigate="handleNavigation" 
      @logout="handleLogout"
      @open-search="isSearchOpen = true"
      @open-notifications="isNotificationOpen = true"
      @open-create="isCreateOpen = true"
    />
    
    <!-- Main Content Area -->
    <main class="main-content">
      <RouterView />
    </main>

    <!-- Panels & Modals -->
    <SearchPanel :is-open="isSearchOpen" @close="isSearchOpen = false" @search="handleSearch" />
    <NotificationPanel :is-open="isNotificationOpen" @close="isNotificationOpen = false" />
    <CreatePostOverlay :is-open="isCreateOpen" @close="isCreateOpen = false" @upload="handlePostUpload" />
    
    <!-- Mini Messages Component (appears on all pages) -->
    <MiniMessagesComponent />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { DashboardPage } from '../types'
import Sidebar from './Sidebar.vue'
import SearchPanel from './SearchPanel.vue'
import NotificationPanel from './NotificationPanel.vue'
import CreatePostOverlay from './CreatePostOverlay.vue'
import MiniMessagesComponent from './MiniMessagesComponent.vue'

const router = useRouter()
const route = useRoute()

const isSearchOpen = ref(false)
const isNotificationOpen = ref(false)
const isCreateOpen = ref(false)
const notificationCount = ref(0)

const currentPage = computed(() => {
  const pageName = route.path.split('/').pop()
  return (pageName as DashboardPage) || 'home'
})

const handleNavigation = (page: DashboardPage) => {
  router.push({ name: page })
}

const handleLogout = () => {
  // TODO: Clear auth token/session from backend
  router.push({ name: 'login' })
}

const handleSearch = (query: string) => {
  console.log('Search for:', query)
  // TODO: Implement search backend call
}

const handlePostUpload = (file: File, description: string) => {
  console.log('Upload post:', file.name, 'Description:', description)
  // TODO: Implement post upload to backend with description
  isCreateOpen.value = false
}
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
