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
        <button class="close-btn" @click="closeSearch">✕</button>
      </div>
      <div class="search-results">
        <!-- Results will be populated from backend -->
        <div class="empty-state" v-if="!searchQuery">
          <p>Start typing to search for users</p>
        </div>
        <div class="empty-state" v-else>
          <p>Search results for "{{ searchQuery }}"</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  close: []
  search: [query: string]
}>()

const searchQuery = ref('')

const closeSearch = () => {
  searchQuery.value = ''
  emit('close')
}
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
  background: var(--background-dark);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  animation: slideInRight 0.3s ease;
}

@keyframes slideInRight {
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
  border-bottom: 1px solid var(--border-color);
  align-items: center;
}

.search-input {
  flex: 1;
  background: var(--surface-light);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
  padding: 10px 15px;
  border-radius: 20px;
  font-size: 14px;
}

.search-input::placeholder {
  color: var(--text-secondary);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 20px;
  cursor: pointer;
  padding: 5px 10px;
}

.search-results {
  flex: 1;
  overflow-y: auto;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .search-panel {
    width: 100%;
  }
}
</style>
