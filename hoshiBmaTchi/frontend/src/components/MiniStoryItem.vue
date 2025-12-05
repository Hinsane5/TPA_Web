<template>
  <button 
    class="mini-story"
    :class="{ active: isActive }"
    @click="handleClick"
    :title="`Story by ${username}`"
  >
    <div class="mini-avatar">{{ avatar }}</div>
    <span class="mini-username">{{ username }}</span>
  </button>
</template>

<script setup lang="ts">
interface Props {
  avatar: string;
  username: string;
  isActive: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  click: [];
}>();

const handleClick = () => {
  emit('click');
};
</script>

<style scoped>
.mini-story {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  background: none;
  border: none;
  cursor: pointer;
  flex-shrink: 0;
  padding: 0;
  transition: opacity 0.2s ease;
}

.mini-story:hover {
  opacity: 0.8;
}

.mini-story.active {
  opacity: 1;
}

.mini-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  border: 2px solid #0095f6;
}

.mini-story.active .mini-avatar {
  border-color: #0095f6;
}

.mini-username {
  font-size: 10px;
  color: #a0a0a0;
  max-width: 50px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
