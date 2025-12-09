<template>
  <Transition name="toast-slide">
    <div v-if="store.toastMessage" class="toast-notification" @click="handleClick">
      <img :src="store.toastMessage.sender_image" class="toast-avatar" />
      <div class="toast-content">
        <span class="toast-user">{{ store.toastMessage.sender_name }}</span>
        <span class="toast-text">{{ getText(store.toastMessage.type) }}</span>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { useNotificationStore } from '@/stores/notificationStore';
import { useRouter } from 'vue-router';

const store = useNotificationStore();
const router = useRouter();

const getText = (type: string) => {
  if (type === 'like') return 'liked your post';
  if (type === 'comment') return 'commented on your post';
  if (type === 'follow') return 'started following you';
  return 'sent a notification';
};

const handleClick = () => {
  const notif = store.toastMessage;
  if (!notif) return;
  
  if (notif.type === 'like' || notif.type === 'comment') {
    router.push({ name: 'PostDetail', params: { id: notif.entity_id } });
  } else {
    router.push({ name: 'Profile', params: { username: notif.sender_name } });
  }
  store.toastMessage = null; 
};
</script>

<style scoped>
.toast-notification {
  position: fixed; top: 20px; right: 20px;
  background: #262626; border-radius: 8px;
  padding: 12px; display: flex; align-items: center;
  box-shadow: 0 4px 12px rgba(0,0,0,0.5);
  cursor: pointer; z-index: 1000; border: 1px solid #333;
  color: white; min-width: 250px;
}
.toast-avatar { 
    width: 32px; 
    height: 32px; 
    border-radius: 50%; 
    margin-right: 10px; 
}

.toast-content { 
    display: flex; 
    flex-direction: column; 
    font-size: 13px; 
}

.toast-user { 
    font-weight: bold; 
}

.toast-slide-enter-active, .toast-slide-leave-active { 
    transition: all 0.3s ease; 
}

.toast-slide-enter-from, .toast-slide-leave-to { 
    opacity: 0; 
    transform: translateX(30px); 
}
</style>