<template>
  <div class="progress-bar"
    :class="{ active: isActive, completed: isCompleted }"
  ></div>
</template>

<script setup lang="ts">
interface Props {
  isActive: boolean;
  isCompleted: boolean;
}

defineProps<Props>();
</script>

<style scoped>
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
