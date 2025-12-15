<template>
  <div
    :class="['conversation-item', { active: isActive }]"
    @click="$emit('select')"
  >
    <div class="conversation-avatar">
      <img
        v-if="participant"
        :src="participant.avatar"
        :alt="participant.username"
        class="avatar"
      />
      <div
        v-if="participant?.isOnline"
        class="online-status"
      ></div>
    </div>

    <div class="conversation-info">
      <div class="info-header">
        <h4 class="conversation-name">
          {{ participant?.fullName || 'Unknown' }}
        </h4>
        <span class="message-time">{{
          formatTime(conversation.updatedAt)
        }}</span>
      </div>
      <p class="message-preview">{{ getMessagePreview() }}</p>
    </div>

    <div v-if="conversation.unreadCount > 0" class="unread-badge">
      {{ conversation.unreadCount }}
    </div>

    <button
      class="delete-btn"
      @click.stop="$emit('delete')"
      title="Delete conversation"
    >
      ✕
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Conversation } from "../types/chat";
import { useChatStore } from "../composables/useChatStore";

interface Props {
  conversation: Conversation;
  isActive: boolean;
  currentUserId: string;
}

const props = defineProps<Props>();
const store = useChatStore();

defineEmits<{
  select: [];
  delete: [];
}>();

const participant = computed(() => {
  const parts = props.conversation.participants;
  if (!parts || parts.length === 0) return null;

  const other = parts.find(p => p.id !== props.currentUserId);
  return other || parts[0];
});

const formatTime = (date: Date | string) => {
  const now = new Date();
  const messageDate = new Date(date);
  const diffMs = now.getTime() - messageDate.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return "Now";
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return messageDate.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
  });
};

const getMessagePreview = () => {
  const lastMessage = props.conversation.lastMessage;
  
  if (!lastMessage) return "No messages yet";

  const isOwn = lastMessage.senderId === props.currentUserId;
  const prefix = isOwn ? "You: " : "";

  const content = lastMessage.content || ""; 

  if (!content && lastMessage.mediaUrl) {
      return prefix + "[Media]";
  }

  return prefix + (content.substring(0, 50) || "Sent a message");
};
</script>

<style scoped>
.conversation-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 12px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  position: relative;
  border-radius: 8px;
  margin: 0 8px;
}

.conversation-item:hover {
  background: #262626;
}

.conversation-item.active {
  background: #262626;
}

.conversation-avatar {
  position: relative;
  flex-shrink: 0;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
  background-color: #333; /* Fallback color if image fails to load */
}

.online-status {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 12px;
  height: 12px;
  background: #31a24c;
  border: 2px solid #1a1a1a;
  border-radius: 50%;
}

.conversation-info {
  flex: 1;
  min-width: 0;
}

.info-header {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 8px;
}

.conversation-name {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-time {
  font-size: 12px;
  color: #a0a0a0;
  flex-shrink: 0;
}

.message-preview {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: #a0a0a0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.unread-badge {
  background: #0084ff;
  color: #fff;
  border-radius: 50%;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.delete-btn {
  opacity: 0;
  background: none;
  border: none;
  color: #a0a0a0;
  cursor: pointer;
  font-size: 14px;
  padding: 4px 8px;
  transition: opacity 0.2s ease, color 0.2s ease;
}

.conversation-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  color: #ff4458;
}
</style>