<template>
  <div class="reels-container">
    <h2 class="reels-title">Reels</h2>
    <div class="reels-grid">
      <!-- Reels will be populated from backend -->
      <div class="reel-item" v-for="n in 8" :key="`reel-${n}`">
        <div class="reel-media">
          <span>🎬</span>
        </div>
        <div class="reel-info">
          <p class="reel-author">Author {{ n }}</p>
          <div class="reel-actions">
            <span>❤️ 5.2K</span>
            <span>💬 1.2K</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!hasContent" class="empty-state">
      <p>No reels yet. Create your first reel!</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const hasContent = ref(false)

onMounted(() => {
  // TODO: Fetch reels from backend
})
</script>

<style scoped>
.reels-container {
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
  color: #fff;
}

.reels-title {
  margin: 0 0 30px 0;
  font-size: 32px;
  font-weight: 700;
}

.reels-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
}

.reel-item {
  background: #1a1a1a;
  border: 1px solid #262626;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
}

.reel-item:hover {
  transform: scale(1.02);
}

.reel-media {
  aspect-ratio: 9 / 16;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
}

.reel-info {
  padding: 12px;
  border-top: 1px solid #262626;
}

.reel-author {
  margin: 0 0 8px 0;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.reel-actions {
  display: flex;
  gap: 15px;
  font-size: 12px;
  color: #a0a0a0;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #a0a0a0;
}

@media (max-width: 768px) {
  .reels-container {
    padding: 15px;
  }

  .reels-grid {
    grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
    gap: 15px;
  }
}
</style>
