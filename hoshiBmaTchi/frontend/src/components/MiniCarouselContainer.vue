<template>
  <div class="mini-carousel">
    <MiniStoryItem 
      v-for="group in storyGroups" 
      :key="group.userId"
      :avatar="group.userAvatar" 
      :username="group.username"
      :isActive="isGroupActive(group.userId)"
      @click="emit('select-story', group.startIndex)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import MiniStoryItem from './MiniStoryItem.vue';
import type { Story } from '../types/stories';

interface Props {
  stories: Story[];
  currentStoryIndex: number;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  'select-story': [index: number];
}>();

const storyGroups = computed(() => {
  const groups: any[] = [];
  const seenUsers = new Set<string>();

  props.stories.forEach((story, index) => {
    if (story.userId && !seenUsers.has(story.userId)) {
      seenUsers.add(story.userId);
      
      groups.push({
        userId: story.userId,
        username: story.user?.username || story.username,
        userAvatar: story.user?.userAvatar || story.userAvatar,
        startIndex: index
      });
    }
  });

  return groups;
});


const isGroupActive = (userId: string): boolean => {
  const activeStory = props.stories[props.currentStoryIndex];
  return activeStory?.userId === userId;
};
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
