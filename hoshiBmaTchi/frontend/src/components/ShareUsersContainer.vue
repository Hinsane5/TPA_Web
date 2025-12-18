<script setup lang="ts">
import ShareUserItem from './ShareUserItem.vue';
import type { User } from '../types/stories';

interface Props {
  users: User[];
  selectedUsers: Set<string>;
}

defineProps<Props>();
const emit = defineEmits<{
  'select-user': [userId: string];
}>();
</script>

<template>
  <ShareUserItem 
    v-for="user in users" 
    :key="user.id"
    :user="user"
    :is-selected="selectedUsers.has(user.id)"
    @select="emit('select-user', user.id)"
  />
</template>
