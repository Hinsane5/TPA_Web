<template>
  <div class="mini-messages" :style="{ left: position.x + 'px', top: position.y + 'px' }" @mousedown="startDrag">
    <div class="messages-header" @mousemove.stop>
      <span class="messages-title">💬 Messages</span>
      <button class="minimize-btn" @click="isMinimized = !isMinimized">{{ isMinimized ? '▲' : '▼' }}</button>
    </div>
    <div v-if="!isMinimized" class="messages-content">
      <!-- Mini messages content will be populated from backend -->
      <div class="empty-state">
        <p>No recent messages</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const isMinimized = ref(false)
const position = ref({ x: 1570, y: 850 })
const isDragging = ref(false)
const dragStart = ref({ x: 0, y: 0 })

const startDrag = (event: MouseEvent) => {
  isDragging.value = true
  dragStart.value = { x: event.clientX - position.value.x, y: event.clientY - position.value.y }
  
  const handleMouseMove = (moveEvent: MouseEvent) => {
    if (isDragging.value) {
      position.value = {
        x: moveEvent.clientX - dragStart.value.x,
        y: moveEvent.clientY - dragStart.value.y,
      }
    }
  }
  
  const handleMouseUp = () => {
    isDragging.value = false
    window.removeEventListener('mousemove', handleMouseMove)
    window.removeEventListener('mouseup', handleMouseUp)
  }
  
  window.addEventListener('mousemove', handleMouseMove)
  window.addEventListener('mouseup', handleMouseUp)
}
</script>

<style scoped>
.mini-messages {
  position: fixed;
  width: 300px;
  background: var(--surface-light);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  z-index: 400;
  user-select: none;
}

.messages-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 15px;
  border-bottom: 1px solid var(--border-color);
  cursor: move;
  background: var(--background-dark);
}

.messages-title {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 14px;
}

.minimize-btn {
  background: none;
  border: none;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 12px;
  padding: 4px 8px;
}

.messages-content {
  max-height: 400px;
  overflow-y: auto;
}

.empty-state {
  padding: 20px;
  text-align: center;
  color: var(--text-secondary);
  font-size: 14px;
}
</style>
