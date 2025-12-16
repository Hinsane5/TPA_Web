<template>
  <aside class="sidebar">
    <div class="sidebar-content">
      <div class="logo">
        <span class="logo-text">Instagram</span>
      </div>

      <nav class="nav-menu">
        <a
          v-for="item in navItems"
          :key="item.id"
          href="#"
          :class="['nav-item', { active: currentPage === item.id }]"
          @click.prevent="navigateTo(item.id)"
        >
          <img 
            v-if="item.id === 'profile' && userProfile?.profile_picture_url"
            :src="getDisplayUrl(userProfile.profile_picture_url)"
            :alt="item.label"
            class="nav-icon nav-avatar"
          />
          <img 
            v-else
            :src="item.iconPath" 
            :alt="item.label" 
            class="nav-icon" 
          />
          
          <span class="nav-label">{{ item.label }}</span>
        </a>

        <button 
          class="nav-item search-btn"
          @click="$emit('openSearch')"
        >
          <img src="/icons/search-icon.png" alt="Search" class="nav-icon" />
          <span class="nav-label">Search</span>
        </button>

        <button 
          class="nav-item notifications-btn"
          @click="$emit('openNotifications')"
        >
          <img src="/icons/notifications-icon.png" alt="Notifications" class="nav-icon" />
          <span class="nav-label">Notifications</span>
          <span v-if="notificationCount > 0" class="notification-badge">{{ notificationCount }}</span>
        </button>

        <button 
          class="nav-item create-btn"
          @click="$emit('openCreate')"
        >
          <img src="/icons/create-icon.jpg" alt="Create" class="nav-icon" />
          <span class="nav-label">Create</span>
        </button>
      </nav>

      <div class="sidebar-bottom">
        <button class="nav-item more-button" @click="toggleMoreMenu">
          <img src="/icons/hamburger-more-icon.png" alt="More" class="nav-icon more-icon" />
          <span class="nav-label">More</span>
        </button>

        <div v-if="isMoreMenuOpen" class="more-menu">
          <a 
            v-if="user && user.role === 'admin'" 
            href="#" 
            class="more-menu-item" 
            @click.prevent="navigateTo('admin')"
          >
            <img src="/icons/lock-icon.png" alt="Admin" class="menu-icon" />
            <span>Admin Page</span>
          </a>

          <a href="#" class="more-menu-item" @click.prevent="navigateTo('settings')">
            <img src="/icons/setting-icon.png" alt="Settings" class="menu-icon" />
            <span>Settings</span>
          </a>
          <a href="#" class="more-menu-item" @click.prevent="navigateTo('saved')">
            <img src="/icons/save-icon.png" alt="Saved" class="menu-icon" />
            <span>Saved</span>
          </a>
          <button class="more-menu-item" @click="toggleTheme">
            <img src="/icons/theme-icon.png" alt="Theme" class="menu-icon" />
            <span>Theme</span>
          </button>
          <a href="#" class="more-menu-item logout" @click.prevent="handleLogout">
            <span>Logout</span>
          </a>
        </div>
      </div>

      <div class="meta-section">
        <a href="#" class="meta-link">Also from Meta</a>
      </div>
    </div>

    <div class="user-profile" v-if="userProfile">
      <img 
        :src="getDisplayUrl(userProfile.profile_picture_url)" 
        :alt="userProfile.username" 
        class="user-avatar" 
      />
      <div class="user-info">
        <div class="user-name">{{ userProfile.username }}</div>
        <div class="user-fullname">{{ userProfile.name }}</div>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import type { DashboardPage } from '../types'
import { useAuth } from '../composables/useAuth'
import { usersApi } from '../services/apiService'

interface NavItem {
  id: DashboardPage
  label: string
  iconPath: string
}

const props = defineProps<{
  currentPage: DashboardPage
  notificationCount?: number
}>()

const notificationCount = computed(() => props.notificationCount ?? 0)

const emit = defineEmits<{
  navigate: [page: DashboardPage]
  logout: []
  openSearch: []
  openNotifications: []
  openCreate: []
}>()

const isMoreMenuOpen = ref(false)
const userProfile = ref<any>(null) 

const { user, logout } = useAuth()

const navItems: NavItem[] = [
  { id: 'home', label: 'Home', iconPath: '/icons/home-icon.png' },
  { id: 'explore', label: 'Explore', iconPath: '/icons/explore-icon.png' },
  { id: 'reels', label: 'Reels', iconPath: '/icons/reels-icon.png' },
  { id: 'messages', label: 'Messages', iconPath: '/icons/messages-icon.png' },
  { id: 'profile', label: 'Profile', iconPath: '/icons/profile-icon.png' },
]

// HELPER: Convert minio:9000 to localhost:9000
const getDisplayUrl = (url: string) => {
  if (!url) return 'https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_1280.png';
  return url.replace("http://minio:9000", "http://localhost:9000");
};

const navigateTo = (page: DashboardPage | 'settings' | 'saved' | 'admin') => {
  if (page === 'settings') {
    emit('navigate', 'settings' as any)
  } else if (page === 'saved') {
    emit('navigate', 'profile')
  } else if (page === 'admin') {
    emit('navigate', 'admin' as any)
  } else {
    emit('navigate', page)
  }
  isMoreMenuOpen.value = false
}

const toggleMoreMenu = () => {
  isMoreMenuOpen.value = !isMoreMenuOpen.value
}

const toggleTheme = () => {
  console.log('Toggle theme')
}

const handleLogout = () => {
  logout()
}

const fetchUserProfile = async () => {
  try {
    const res = await usersApi.getMe();
    userProfile.value = res.data;
  } catch (error) {
    console.error("Failed to load user profile for sidebar", error);
  }
}

// Watch for user changes (e.g. after login) to re-fetch profile
watch(user, (newUser) => {
  if (newUser) {
    fetchUserProfile();
  }
}, { immediate: true });

onMounted(() => {
  fetchUserProfile();
})
</script>

<style scoped>
.sidebar {
  width: 240px;
  height: 100vh;
  background-color: var(--background-dark);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 0;
  overflow-y: auto;
  z-index: 100;
}

.sidebar-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px 16px;
  gap: 20px;
}

.logo {
  padding: 16px 8px;
  margin-bottom: 10px;
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -1px;
}

.logo-text {
  background: linear-gradient(135deg, #5b5bff 0%, #ff006e 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  text-decoration: none;
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 500;
  transition: all 0.3s ease;
  cursor: pointer;
  border: none;
  background: none;
  text-align: left;
  font-family: inherit;
  position: relative;
}

.nav-item:hover {
  background-color: var(--surface-light);
}

.nav-item.active {
  background-color: var(--surface-light);
  color: var(--primary-color);
}

.nav-icon {
  width: 24px;
  height: 24px;
  object-fit: contain;
  filter: brightness(0.9);
}

.nav-item.active .nav-icon {
  filter: brightness(1.2);
}

.nav-item.search-btn,
.nav-item.notifications-btn,
.nav-item.create-btn {
  background: none;
  border: none;
}

.notification-badge {
  position: absolute;
  top: 8px;
  right: 8px;
  background: #ff0000;
  color: white;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
}

.nav-label {
  white-space: nowrap;
}

.more-icon {
  width: 20px;
  height: 20px;
}

.sidebar-bottom {
  position: relative;
}

.more-button {
  width: 100%;
}

.more-menu {
  position: absolute;
  bottom: 100%;
  left: 0;
  right: 0;
  background-color: var(--surface-light);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 8px 0;
  margin-bottom: 8px;
  display: flex;
  flex-direction: column;
}

.more-menu-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  color: var(--text-primary);
  text-decoration: none;
  cursor: pointer;
  transition: background-color 0.2s ease;
  border: none;
  background: none;
  font-size: 14px;
  text-align: left;
  font-family: inherit;
}

.more-menu-item:hover {
  background-color: var(--surface-dark);
}

.more-menu-item.logout {
  border-top: 1px solid var(--border-color);
  color: var(--error-color);
}

.menu-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
  filter: brightness(0.9);
}

.more-menu-item:hover .menu-icon {
  filter: brightness(1.1);
}

.meta-section {
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.meta-link {
  display: block;
  padding: 12px 16px;
  color: var(--text-secondary);
  font-size: 14px;
  text-decoration: none;
  transition: color 0.2s ease;
}

.meta-link:hover {
  color: var(--text-primary);
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-fullname {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.loading .skeleton-text {
  width: 100%;
  height: 16px;
  background: linear-gradient(90deg, #e0e0e0 25%, #f0f0f0 50%, #e0e0e0 75%);
  background-size: 200px 100%;
  animation: skeleton-animation 1.5s infinite ease-in-out;
}

.loading .skeleton-text.small {
  height: 12px;
}

@keyframes skeleton-animation {
  0% {
    background-position: -200px 0;
  }
  100% {
    background-position: 200px 0;
  }
}

/* Scrollbar styling */
.sidebar::-webkit-scrollbar {
  width: 8px;
}

.sidebar::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar::-webkit-scrollbar-thumb {
  background: var(--border-color);
  border-radius: 4px;
}

.sidebar::-webkit-scrollbar-thumb:hover {
  background: var(--text-tertiary);
}

/* Responsive - Hide on mobile, collapse on tablet */
@media (max-width: 1024px) {
  .sidebar {
    width: 64px;
  }

  .nav-label,
  .meta-section,
  .user-info {
    display: none;
  }

  .sidebar-content {
    padding: 12px 8px;
  }

  .nav-item {
    justify-content: center;
    padding: 12px 8px;
  }

  .user-profile {
    flex-direction: column;
    padding: 8px;
  }

  .user-avatar {
    width: 36px;
    height: 36px;
  }

  .more-menu {
    right: 0;
    left: auto;
    width: 200px;
  }
}

@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    height: auto;
    border-right: none;
    border-bottom: 1px solid var(--border-color);
    flex-direction: row;
    position: static;
    padding: 8px 12px;
  }

  .sidebar-content {
    flex-direction: row;
    gap: 4px;
    padding: 8px 0;
    align-items: center;
  }

  .logo {
    display: none;
  }

  .nav-menu {
    flex-direction: row;
    gap: 4px;
  }

  .nav-item {
    padding: 8px 12px;
    font-size: 14px;
    gap: 8px;
  }

  .user-profile {
    display: none;
  }

  .meta-section {
    display: none;
  }
}
</style>