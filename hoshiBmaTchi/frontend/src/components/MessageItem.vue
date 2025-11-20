<template>
  <div :class="['message-item', { 'is-own': isOwnMessage }]">
    <div v-if="!isOwnMessage" class="message-avatar">
      <img
        :src="message.senderAvatar"
        :alt="message.senderName"
        class="avatar"
      />
    </div>

    <div class="message-content-wrapper">
      <div class="message-bubble">
        <p v-if="message.messageType === 'text'" class="message-text">
          {{ message.content }}
        </p>
        <div v-else-if="message.messageType === 'image'" class="message-media">
          <img
            :src="message.content"
            :alt="message.senderName"
            class="media-image"
          />
        </div>
        <div v-else-if="message.messageType === 'video'" class="message-media">
          <video :src="message.content" class="media-video" controls></video>
        </div>
        <div v-else-if="message.messageType === 'gif'" class="message-media">
          <img
            :src="message.content"
            :alt="message.senderName"
            class="media-gif"
          />
        </div>

        <div v-if="message.isEdited" class="edited-indicator">(edited)</div>
      </div>

      <div class="message-footer">
        <span class="message-time">
          {{ formatTime(message.timestamp) }}
        </span>
        <span v-if="isOwnMessage" :class="['message-status', message.status]">
          {{
            message.status === "sending"
              ? "⏱️"
              : message.status === "sent"
              ? "✓"
              : "✓✓"
          }}
        </span>
      </div>

      <div v-if="isOwnMessage" class="message-actions">
        <button
          v-if="message.canUnsend"
          class="action-btn"
          @click="$emit('unsend', message.id)"
          title="Unsend message"
        >
          ⤴️
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Message } from "../types/chat";

interface Props {
  message: Message;
  isOwnMessage: boolean;
}

defineProps<Props>();
defineEmits<{
  unsend: [messageId: string];
}>();

const formatTime = (date: Date) => {
  return new Date(date).toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: true,
  });
};
</script>

<style scoped>
.message-item {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-item.is-own {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.message-content-wrapper {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.message-item.is-own .message-content-wrapper {
  align-items: flex-end;
}

.message-bubble {
  background: #262626;
  border-radius: 18px;
  padding: 12px 16px;
  max-width: 400px;
  word-wrap: break-word;
  position: relative;
}

.message-item.is-own .message-bubble {
  background: #0084ff;
  border-radius: 18px;
}

.message-text {
  margin: 0;
  color: #fff;
  font-size: 14px;
  line-height: 1.4;
}

.message-media {
  max-width: 300px;
  border-radius: 8px;
  overflow: hidden;
}

.media-image,
.media-gif {
  width: 100%;
  height: auto;
  display: block;
  border-radius: 8px;
}

.media-video {
  width: 100%;
  max-height: 400px;
  border-radius: 8px;
}

.edited-indicator {
  font-size: 11px;
  color: #a0a0a0;
  margin-top: 4px;
}

.message-footer {
  display: flex;
  gap: 6px;
  align-items: center;
  font-size: 12px;
  color: #a0a0a0;
}

.message-status {
  font-size: 11px;
}

.message-status.seen {
  color: #0084ff;
}

.message-actions {
  display: flex;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.message-item:hover .message-actions {
  opacity: 1;
}

.action-btn {
  background: none;
  border: none;
  color: #0084ff;
  cursor: pointer;
  font-size: 16px;
  padding: 4px 8px;
  transition: transform 0.2s ease;
}

.action-btn:hover {
  transform: scale(1.1);
}
</style>
