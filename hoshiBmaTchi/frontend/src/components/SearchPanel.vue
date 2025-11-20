<template>
  <div v-if="isOpen" class="search-panel-overlay" @click="closeSearch">
    <div class="search-panel" @click.stop>
      <div class="search-header">
        <input 
          v-model="searchQuery"
          type="text"
          placeholder="Search username..."
          class="search-input"
          autofocus
        />
        <button class="close-btn" @click="closeSearch">×</button>
      </div>

      <div class="search-results">
        <!-- Recent searches section -->
        <div v-if="!searchQuery && recentSearches.length > 0" class="recent-section">
          <div class="section-header">
            <span>Recent</span>
            <button @click="clearHistory" class="clear-all-btn">Clear all</button>
          </div>
          
          <UserListItem 
            v-for="user in recentSearches" 
            :key="user.user_id"
            :user="user"
            :show-close="true"
            @click="goToProfile(user)"
            @remove="removeFromHistory(user.user_id)"
          />
        </div>

        <!-- Search results -->
        <div v-else-if="searchQuery" class="results-list">
          <UserListItem 
            v-for="user in results" 
            :key="user.user_id"
            :user="user"
            @click="goToProfile(user)"
          />
        </div>

        <!-- Empty state -->
        <div v-else class="empty-state">
          <p>Start typing to search for users</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import UserListItem from './UserListItem.vue'
import { usersApi } from '../services/apiService'

interface Props {
  isOpen: boolean
}

const props = defineProps<Props>()

const emit = defineEmits<{
  close: []
}>()

const router = useRouter()
const searchQuery = ref('')
const results = ref<any[]>([])
const recentSearches = ref<any[]>([])
let searchTimeout: ReturnType<typeof setTimeout> | null = null // Variable for debounce timer

watch(searchQuery, (newQuery) => {
  // 1. Clear the previous timer if the user keeps typing
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }

  // 2. If query is empty, clear results and stop
  if (!newQuery.trim()) {
    results.value = []
    return
  }

  // 3. Set a new timer to wait 300ms before calling API
  searchTimeout = setTimeout(async () => {
    try {
      const response = await usersApi.searchUsers(newQuery)
      // Based on your backend handler, the data is wrapped in "users"
      results.value = response.data.users || [] 
    } catch (error) {
      console.error('Search failed:', error)
    }
  }, 300)
})

const closeSearch = () => {
  searchQuery.value = ''
  emit('close')
}

const goToProfile = (user: any) => {
  addToHistory(user)
  router.push(`/dashboard/profile/${user.user_id}`)
}

const addToHistory = (user: any) => {
  const filtered = recentSearches.value.filter(u => u.user_id !== user.user_id)
  const newHistory = [user, ...filtered].slice(0, 10)
  recentSearches.value = newHistory
  localStorage.setItem('searchHistory', JSON.stringify(newHistory))
}

const clearHistory = () => {
  recentSearches.value = []
  localStorage.removeItem('searchHistory')
}

const removeFromHistory = (userId: string) => {
  recentSearches.value = recentSearches.value.filter(u => u.user_id !== userId)
  localStorage.setItem('searchHistory', JSON.stringify(recentSearches.value))
}

onMounted(() => {
  const history = localStorage.getItem('searchHistory')
  if (history) recentSearches.value = JSON.parse(history)
})
</script>

<style scoped>
.search-panel-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 500;
  display: flex;
}

.search-panel {
  width: 350px;
  height: 100vh;
  background: #121212;
  border-right: 1px solid #262626;
  display: flex;
  flex-direction: column;
  animation: slideInLeft 0.3s ease;
}

@keyframes slideInLeft {
  from {
    transform: translateX(-100%);
  }
  to {
    transform: translateX(0);
  }
}

.search-header {
  display: flex;
  gap: 10px;
  padding: 15px;
  border-bottom: 1px solid #262626;
  align-items: center;
}

.search-input {
  flex: 1;
  background: #262626;
  border: 1px solid #404040;
  color: #ffffff;
  padding: 10px 15px;
  border-radius: 20px;
  font-size: 14px;
  font-family: inherit;
}

.search-input::placeholder {
  color: #808080;
}

.search-input:focus {
  outline: none;
  border-color: #606060;
}

.close-btn {
  background: none;
  border: none;
  color: #ffffff;
  font-size: 20px;
  cursor: pointer;
  padding: 5px 10px;
  transition: color 0.2s ease;
}

.close-btn:hover {
  color: #b0b0b0;
}

.search-results {
  flex: 1;
  overflow-y: auto;
}

.recent-section {
  display: flex;
  flex-direction: column;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  border-bottom: 1px solid #262626;
  gap: 10px;
}

.section-header span {
  color: #ffffff;
  font-size: 14px;
  font-weight: 600;
}

.clear-all-btn {
  background: none;
  border: none;
  color: #0095f6;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  padding: 4px 8px;
  transition: color 0.2s ease;
  font-family: inherit;
}

.clear-all-btn:hover {
  color: #00a3ff;
}

.results-list {
  display: flex;
  flex-direction: column;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: #808080;
  font-size: 14px;
}

@media (max-width: 768px) {
  .search-panel {
    width: 100%;
  }
}
</style>
