<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useChatStore } from '../composables/useChatStore';
import type { User } from '../types/chat';

const router = useRouter();
const route = useRoute();
const chatStore = useChatStore();

const isVisible = computed(() => {
  return route.path !== '/dashboard/messages';
});

onMounted(async () => {
  await chatStore.fetchConversations();
});

const sortedConversations = computed(() => {
  return [...chatStore.conversations.value].sort((a, b) => {
    return new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime();
  });
});

const recentChatUsers = computed(() => {
  const users: Partial<User>[] = [];
  const myId = chatStore.currentUser.value?.id;

  const topConvos = sortedConversations.value.slice(0, 3);

  topConvos.forEach((conv) => {
    const other = conv.participants.find((p) => p.id !== myId);

    if (other) {
      users.push(other);
    } else if (conv.participants.length > 0 && conv.participants[0]) {
      users.push(conv.participants[0]);
    }
  });

  return users;
});

const unreadSourcesCount = computed(() => {
  return chatStore.conversations.value.filter(c => c.unreadCount > 0).length;
});

const position = ref({ x: window.innerWidth - 350, y: window.innerHeight - 100 });
const isDragging = ref(false);
const dragStart = ref({ x: 0, y: 0 });
const hasMoved = ref(false);

const startDrag = (event: MouseEvent) => {
  isDragging.value = true;
  hasMoved.value = false;
  dragStart.value = { x: event.clientX - position.value.x, y: event.clientY - position.value.y };
  
  const handleMouseMove = (moveEvent: MouseEvent) => {
    if (isDragging.value) {
      hasMoved.value = true;
      position.value = {
        x: moveEvent.clientX - dragStart.value.x,
        y: moveEvent.clientY - dragStart.value.y,
      };
    }
  };
  
  const handleMouseUp = () => {
    isDragging.value = false;
    window.removeEventListener('mousemove', handleMouseMove);
    window.removeEventListener('mouseup', handleMouseUp);

    if (!hasMoved.value) {
      navigateToMessages();
    }
  };
  
  window.addEventListener('mousemove', handleMouseMove);
  window.addEventListener('mouseup', handleMouseUp);
};

const navigateToMessages = () => {
  router.push('/dashboard/messages');
};
</script>

<template>
  <div 
    v-if="isVisible" 
    class="mini-messages-pill" 
    :style="{ left: position.x + 'px', top: position.y + 'px' }" 
    @mousedown="startDrag"
  >
    <div class="pill-header">
      <div class="icon-wrapper">
        <img 
          src="../../public/icons/messages-icon.png" 
          alt="Messages" 
          class="messenger-icon" 
        />
        
        <div v-if="unreadSourcesCount > 0" class="badge">
          {{ unreadSourcesCount }}
        </div>
      </div>
      <span class="label">Messages</span>
    </div>

    <div class="avatars-container">
      <div 
        v-for="(user, index) in recentChatUsers" 
        :key="user.id" 
        class="avatar-wrapper"
        :style="{ zIndex: 3 - index }"
      >
        <img :src="user.avatar || '/placeholder.svg'" :alt="user.username" class="avatar-img" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.mini-messages-pill {
  position: fixed;
  display: flex;
  align-items: center;
  justify-content: space-between;
  

  min-width: 180px;
  height: 56px;
  padding: 0 16px;
  border-radius: 28px;
  
  background-color: #262626; 
  border: 1px solid #363636;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  color: white;
  
  z-index: 1;
  user-select: none;
  cursor: pointer;
  transition: transform 0.1s ease;
}

.mini-messages-pill:active {
  transform: scale(0.98);
}

.pill-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-wrapper {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.messenger-icon {
  width: 24px;
  height: 24px;
  color: white;
}

.badge {
  position: absolute;
  top: -6px;
  right: -6px;
  background-color: #ff3040; /* Red notification color */
  color: white;
  font-size: 10px;
  font-weight: bold;
  min-width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #262626; /* Match background to create "cutout" effect */
}

.label {
  font-weight: 600;
  font-size: 15px;
  color: #f5f5f5;
}

.avatars-container {
  display: flex;
  align-items: center;
  margin-left: 12px;
}

.avatar-wrapper {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 2px solid #262626; /* Border matches background for separation */
  overflow: hidden;
  margin-left: -12px; /* Negative margin for overlapping effect */
  background-color: #404040;
}

.avatar-wrapper:first-child {
  margin-left: 0;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
</style>