<script setup lang="ts">
import MiniStoryItem from './MiniStoryItem.vue';
import { useStories } from '../composables/useStories';

// FIX: Remove Props. Consume global state directly.
// This prevents the "Maximum call stack size" error by removing the reactive loop.
const { storyGroups, currentGroupIndex, selectGroup } = useStories();

const emit = defineEmits<{
  'open-viewer': [];
}>();

const handleSelect = (index: number) => {
  selectGroup(index); // Update global state
  emit('open-viewer'); // Tell HomePage to open the overlay
};
</script>

<template>
  <div class="mini-carousel">
    <MiniStoryItem 
      v-for="(group, index) in storyGroups" 
      :key="group.userId"
      :avatar="group.userAvatar" 
      :username="group.username"
      :is-active="index === currentGroupIndex"
      :has-unseen="group.hasUnseen"
      @click="handleSelect(index)"
    />
  </div>
</template>

<style scoped>
.mini-carousel {
  display: flex;
  gap: 8px;
  padding: 12px;
  background: none;
  overflow-x: auto;
  position: relative;
  z-index: 1;
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.2) transparent;
}

.mini-carousel::-webkit-scrollbar {
  height: 4px;
}

.mini-carousel::-webkit-scrollbar-track {
  background: transparent;
}

.mini-carousel::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}
</style>
