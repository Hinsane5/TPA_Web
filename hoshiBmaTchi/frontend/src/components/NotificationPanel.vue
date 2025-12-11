<template>
  <div v-if="isOpen" class="notification-panel-overlay" @click="closePanel">
    <div class="notification-panel" @click.stop>
      <div class="panel-header">
        <h2>Notifications</h2>
        <button class="close-btn" @click="closePanel">✕</button>
      </div>
      
      <div class="notification-list">
        <div 
          v-for="notif in store.notifications" 
          :key="notif.ID" 
          class="notif-item" 
          @click="handleNotificationClick(notif)"
        >
          <img :src="notif.sender_image || '/icons/profile-icon.png'" class="avatar" />
          
          <div class="notif-content">
            <span class="username">{{ notif.sender_name }}</span>
            <span class="message">{{ notif.message }}</span>
          </div>
        </div>

        <div v-if="store.notifications.length === 0" class="empty-state">
          <p>No notifications yet</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useNotificationStore } from '@/stores/notificationStore';
import { useRouter } from 'vue-router';
import type { Notification } from '@/types';

defineProps<{ isOpen: boolean }>();
const emit = defineEmits<{ close: [] }>();

const store = useNotificationStore();
const router = useRouter();

const closePanel = () => emit('close');

const handleNotificationClick = (notif: Notification) => {
  // 1. Debug: Check if the click is even registering
  console.log("Notification Clicked:", notif);

  closePanel();
  
  // 2. Safety: Ensure fields exist and handle case sensitivity
  const type = notif.type ? notif.type.toLowerCase() : "";
  const targetId = notif.sender_id;

  if (!targetId) {
    console.error("Error: Notification is missing sender_id", notif);
    return;
  }

  // 3. Logic: Redirect based on type
  if (['follow', 'mention', 'like', 'comment'].includes(type)) {
    console.log(`Redirecting to profile of user: ${targetId}`);
    
    router.push({ 
      name: 'profile', 
      params: { id: targetId } 
    }).then(() => {
        // Force reload if we are already on a profile page but changing users
        // (Optional, as your ProfilePage watcher handles this, but good for safety)
    });
  } else {
    console.warn("Unknown notification type:", type);
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

/* These classes now match the HTML template */
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
  object-fit: cover;
}
.notif-content {
  flex: 1;
  font-size: 14px;
  display: flex;
  flex-direction: column;
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
