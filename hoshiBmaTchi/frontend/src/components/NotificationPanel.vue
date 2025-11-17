<template>
  <div v-if="isOpen" class="notification-panel-overlay" @click="closePanel">
    <div class="notification-panel" @click.stop>
      <div class="panel-header">
        <h2>Notifications</h2>
        <button class="close-btn" @click="closePanel">✕</button>
      </div>
      <div class="notification-list">
        <!-- Notifications will be populated from backend -->
        <div class="empty-state">
          <p>No notifications yet</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  close: []
}>()

const closePanel = () => {
  emit('close')
}
</script>

<style scoped>
.notification-panel-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 500;
  display: flex;
  justify-content: flex-end;
}

.notification-panel {
  width: 400px;
  height: 100vh;
  background: var(--background-dark);
  border-left: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  animation: slideInLeft 0.3s ease;
}

@keyframes slideInLeft {
  from {
    transform: translateX(100%);
  }
  to {
    transform: translateX(0);
  }
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid var(--border-color);
}

.panel-header h2 {
  margin: 0;
  font-size: 20px;
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-primary);
  font-size: 20px;
  cursor: pointer;
  padding: 5px 10px;
}

.notification-list {
  flex: 1;
  overflow-y: auto;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
  color: var(--text-secondary);
}

@media (max-width: 768px) {
  .notification-panel {
    width: 100%;
  }
}
</style>
