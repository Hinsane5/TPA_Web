<template>
  <div class="mini-carousel">
    <MiniStoryItem 
      v-for="(story, idx) in stories" 
      :key="story.id"
      :avatar="story.userAvatar"
      :username="story.username"
      :isActive="idx === currentStoryIndex"
      @click="emit('select-story', idx)"
    />
  </div>
</template>

<script setup lang="ts">
import MiniStoryItem from './MiniStoryItem.vue';
import type { Story } from '../types/stories';

interface Props {
  stories: Story[];
  currentStoryIndex: number;
}

defineProps<Props>();
const emit = defineEmits<{
  'select-story': [index: number];
}>();
</script>

<style scoped>
.mini-carousel {
  display: flex;
  gap: 8px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.6);
  overflow-x: auto;
  position: relative;
  z-index: 10001;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
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
