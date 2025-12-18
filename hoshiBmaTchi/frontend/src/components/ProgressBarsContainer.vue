<script setup lang="ts">
import type { Story } from '../types/stories';

interface Props {
  stories: Story[];
  currentStoryIndex: number;
}

defineProps<Props>();
</script>

<template>
  <div class="progress-container">
    <div 
      v-for="(story, idx) in stories" 
      :key="story.id"
      class="progress-bar"
      :class="{ active: idx === currentStoryIndex, completed: idx < currentStoryIndex }"
    ></div>
  </div>
</template>

<style scoped>
.progress-container {
  display: flex;
  gap: 4px;
  padding: 8px;
  background: rgba(0, 0, 0, 0.4);
  position: relative;
  z-index: 10001;
}

.progress-bar {
  flex: 1;
  height: 2px;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 1px;
  transition: background 0.3s ease;
}

.progress-bar.active {
  background: rgba(255, 255, 255, 0.8);
  animation: progress 5s linear forwards;
}

.progress-bar.completed {
  background: rgba(255, 255, 255, 0.8);
}

@keyframes progress {
  from {
    width: 0;
  }
  to {
    width: 100%;
  }
}
</style>
