<template>
  <div class="user-item" @click="handleClick">
    <div class="user-item-left">
      <div class="user-avatar">{{ user.userAvatar }}</div>
      <div class="user-text">
        <p class="user-name">{{ user.fullName }}</p>
        <p class="user-handle">{{ user.username }}</p>
      </div>
    </div>
    <button 
      class="user-checkbox"
      :class="{ checked: isSelected }"
      @click.stop="handleClick"
    ></button>
  </div>
</template>

<script setup lang="ts">
import type { User } from '../types/stories';

interface Props {
  user: User;
  isSelected: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  select: [];
}>();

const handleClick = () => {
  emit('select');
};
</script>

<style scoped>
.user-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 8px;
  border-radius: 6px;
  transition: background 0.2s ease;
  cursor: pointer;
}

.user-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.user-item-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #262626;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  flex-shrink: 0;
}

.user-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.user-name {
  margin: 0;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-handle {
  margin: 0;
  color: #a0a0a0;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-checkbox {
  width: 24px;
  height: 24px;
  border: 2px solid #404040;
  border-radius: 50%;
  background: none;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.user-checkbox:hover {
  border-color: #606060;
}

.user-checkbox.checked {
  background: #0095f6;
  border-color: #0095f6;
}

.user-checkbox.checked::after {
  content: '✓';
  color: #fff;
  font-size: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
