<template>
  <div class="explore-container">
    <div class="explore-header">
      <h2>Explore</h2>
      <input type="text" placeholder="Search..." class="search-input" />
    </div>

    <div class="explore-grid">
      <div class="explore-item" v-for="n in 12" :key="`explore-${n}`">
        <div class="item-image">
          <span>📸</span>
        </div>
        <div class="item-overlay">
          <div class="item-stats">
            <span>❤️ 1.2K</span>
            <span>💬 450</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!hasContent" class="empty-state">
      <p>Explore posts coming soon</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const hasContent = ref(false)

onMounted(() => {
  // TODO: Fetch explore posts from backend
})
</script>

<style scoped>
.explore-container {
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
  color: #fff;
}

.explore-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  gap: 20px;
}

.explore-header h2 {
  margin: 0;
  font-size: 32px;
  font-weight: 700;
}

.search-input {
  background: #1a1a1a;
  border: 1px solid #262626;
  color: #fff;
  padding: 10px 15px;
  border-radius: 20px;
  width: 200px;
  font-size: 14px;
}

.search-input::placeholder {
  color: #808080;
}

.explore-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
}

.explore-item {
  aspect-ratio: 1;
  background: #262626;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  position: relative;
  transition: all 0.2s ease;
}

.explore-item:hover {
  transform: scale(1.02);
}

.item-image {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  background: #1a1a1a;
}

.item-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.explore-item:hover .item-overlay {
  opacity: 1;
}

.item-stats {
  display: flex;
  gap: 20px;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #a0a0a0;
}

@media (max-width: 768px) {
  .explore-container {
    padding: 15px;
  }

  .explore-header {
    flex-direction: column;
  }

  .search-input {
    width: 100%;
  }

  .explore-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 15px;
  }
}
</style>
