<template>
  <div v-if="isOpen" class="notification-panel-overlay" @click="closePanel">
    <div class="notification-panel" @click.stop>
      <div class="panel-header">
        <h2>Notifications</h2>
        <button class="close-btn" @click="closePanel">✕</button>
      </div>

      <div class="notification-list">
        <div v-if="notifications.length === 0" class="empty-state">
          <p>No notifications yet</p>
        </div>

        <div 
          v-for="notif in notifications" 
          :key="notif.ID" 
          class="notif-item"
          @click="handleNotifClick(notif)"
        >
          <img :src="notif.sender_image" alt="User Avatar" class="avatar" />
          <div class="notif-content">
            <span class="username">{{ notif.sender_name }}</span>
            <span class="message">{{ notif.message }}</span>
            <span class="time">{{ formatTime(notif.created_at) }}</span>
          </div>
          <div v-if="['like', 'comment'].includes(notif.type)" class="interaction-icon">
            <span v-if="notif.type === 'like'">❤️</span>
             <span v-if="notif.type === 'comment'">💬</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useNotifications } from '@/composables/useNotifications';
import { formatDistanceToNow } from 'date-fns'; 

defineProps<{ isOpen: boolean }>();
const emit = defineEmits<{ close: [] }>();
const router = useRouter();

const { notifications, fetchHistory, connect } = useNotifications();

onMounted(() => {
  connect();     
  fetchHistory();
});

const closePanel = () => emit('close');

const formatTime = (dateStr: string) => {
  return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
};

const handleNotifClick = (notif: any) => {
  closePanel();
  if (notif.type === 'follow') {
    router.push(`/profile/${notif.sender_name}`);
  } else if (['like', 'comment', 'mention'].includes(notif.type)) {
    console.log("Open post:", notif.entity_id);
  }
};
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

.notif-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.2s;
}
.notif-item:hover {
  background: rgba(255, 255, 255, 0.05);
}
.avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  margin-right: 12px;
}
.notif-content {
  flex: 1;
  font-size: 14px;
}
.username {
  font-weight: 600;
  margin-right: 4px;
}
.time {
  color: var(--text-secondary);
  font-size: 12px;
  margin-left: 4px;
}

@media (max-width: 768px) {
  .notification-panel {
    width: 100%;
  }
}
</style>
