<script setup lang="ts">
import { ref, watch, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import UserListItem from './UserListItem.vue'
import { usersApi, postsApi } from '../services/apiService'

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
const isHashtagSearch = ref(false)

let searchTimeout: ReturnType<typeof setTimeout> | null = null 

watch(searchQuery, (newQuery) => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }

  if (!newQuery.trim()) {
    results.value = []
    isHashtagSearch.value = false
    return
  }

  if (newQuery.startsWith('#')) {
    isHashtagSearch.value = true
  } else {
    isHashtagSearch.value = false
  }

  searchTimeout = setTimeout(async () => {
    try {
      if (isHashtagSearch.value) {
        const cleanQuery = newQuery.substring(1);
        if (!cleanQuery) return;
        
        const response = await postsApi.searchHashtags(cleanQuery)
        
        results.value = response.data.hashtags || [];
      
      } else {
        const response = await usersApi.searchUsers(newQuery)
        results.value = response.data.users || [] 
      }
    } catch (error) {
      console.error('Search failed:', error)
      results.value = []
    }
  }, 300)
})

watch(() => props.isOpen, (isOpen) => {
  if (isOpen) {
    loadHistory()
  }
})

const closeSearch = () => {
  searchQuery.value = ''
  emit('close')
}

const goToProfile = (user: any) => {
  addToHistory(user)
  router.push(`/dashboard/profile/${user.user_id}`)
  emit('close')
}

const goToHashtag = (tagName: string) => {
  router.push({ path: '/dashboard/explore', query: { q: tagName } })
  emit('close')
}

const addToHistory = (user: any) => {
  if (!storageKey.value) return 
  const filtered = recentSearches.value.filter(u => u.user_id !== user.user_id)
  const newHistory = [user, ...filtered].slice(0, 10)
  recentSearches.value = newHistory
  localStorage.setItem(storageKey.value, JSON.stringify(newHistory))
}

const clearHistory = () => {
  if (!storageKey.value) return
  recentSearches.value = []
  localStorage.removeItem(storageKey.value)
}

const removeFromHistory = (userId: string) => {
  if (!storageKey.value) return
  recentSearches.value = recentSearches.value.filter(u => u.user_id !== userId)
  localStorage.setItem(storageKey.value, JSON.stringify(recentSearches.value))
}

const getCurrentUserId = (): string | null => {
  const token = localStorage.getItem('accessToken');
  if (!token) return null;
  try {
    const parts = token.split('.');
    if (parts.length < 2) return null;
    const payloadPart = parts[1];
    if (!payloadPart) return null;
    const base64 = payloadPart.replace(/-/g, '+').replace(/_/g, '/');
    const payload = JSON.parse(atob(base64));
    const userId = payload.user_id || payload.sub || payload.id;
    return typeof userId === 'string' ? userId : null;
  } catch(e){ return null; }
}

const storageKey = computed(() => {
  const userId = getCurrentUserId()
  return userId ? `searchHistory_${userId}` : null
})

const loadHistory = () => {
  if (!storageKey.value) {
    recentSearches.value = []
    return
  }
  const history = localStorage.getItem(storageKey.value)
  if (history) {
    try {
      recentSearches.value = JSON.parse(history)
    } catch (e) {
      console.error('Failed to parse search history', e)
      recentSearches.value = []
    }
  } else {
    recentSearches.value = []
  }
}

onMounted(() => {
})
</script>

<template>
  <div v-if="isOpen" class="search-panel-overlay" @click="closeSearch">
    <div class="search-panel" @click.stop>
      <div class="search-header">
        <input 
          v-model="searchQuery"
          type="text"
          placeholder="Search"
          class="search-input"
          autofocus
        />
        <button class="close-btn" @click="closeSearch">×</button>
      </div>

      <div class="search-results">
        <div v-if="!searchQuery && recentSearches.length > 0" class="recent-section">
          <div class="section-header">
            <span>Recent</span>
            <button class="clear-all-btn" @click="clearHistory">Clear all</button>
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

        <div v-else-if="searchQuery" class="results-list">
          
          <div v-if="isHashtagSearch">
             <div 
                v-for="tag in results" 
                :key="tag.name" 
                class="hashtag-item"
                @click="goToHashtag(tag.name)"
             >
                <div class="hashtag-icon-circle">#</div>
                <div class="hashtag-info">
                   <p class="hashtag-name">#{{ tag.name }}</p>
                   <p class="hashtag-count">{{ tag.count }} posts</p>
                </div>
             </div>
          </div>

          <div v-else>
            <UserListItem 
              v-for="user in results" 
              :key="user.user_id"
              :user="user"
              @click="goToProfile(user)"
            />
          </div>

        </div>

        <div v-else class="empty-state">
          <p>Start typing to search</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ... (Keep your existing styles for search-panel-overlay, etc.) ... */
.search-panel-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
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
  from { transform: translateX(-100%); }
  to { transform: translateX(0); }
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

.search-input::placeholder { color: #808080; }
.search-input:focus { outline: none; border-color: #606060; }

.close-btn {
  background: none; border: none;
  color: #ffffff; font-size: 20px;
  cursor: pointer; padding: 5px 10px;
}

.search-results { flex: 1; overflow-y: auto; }
.recent-section { display: flex; flex-direction: column; }
.section-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 15px; border-bottom: 1px solid #262626;
}
.section-header span { color: #ffffff; font-size: 14px; font-weight: 600; }
.clear-all-btn {
  background: none; border: none; color: #0095f6; font-size: 12px; font-weight: 600; cursor: pointer;
}
.results-list { display: flex; flex-direction: column; }
.empty-state { padding: 40px 20px; text-align: center; color: #808080; font-size: 14px; }

/* --- NEW STYLES FOR HASHTAGS --- */
.hashtag-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  color: white;
}
.hashtag-item:hover {
  background-color: #262626;
}
.hashtag-icon-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1px solid #363636;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  font-size: 20px;
  font-weight: 300;
}
.hashtag-info {
  display: flex;
  flex-direction: column;
}
.hashtag-name {
  font-weight: 600;
  font-size: 14px;
  margin: 0;
}
.hashtag-count {
  font-size: 12px;
  color: #a8a8a8;
  margin: 0;
}

@media (max-width: 768px) {
  .search-panel { width: 100%; }
}
</style>