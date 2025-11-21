<template>
  <div v-if="selectedConversation && chatPartner" class="chat-window">
    <div class="chat-header">
      <div class="chat-header-left">
        <div class="header-avatar">
          <img
            :src="chatPartner.avatar"
            :alt="chatPartner.username"
            class="avatar"
          />
          <div v-if="chatPartner.isOnline" class="online-indicator"></div>
        </div>
        <div class="header-info">
          <h3 class="chat-title">{{ chatPartner.fullName }}</h3>
          <p class="chat-subtitle">
            {{ chatPartner.isOnline ? 'Active now' : 'Offline' }}
          </p>
        </div>
      </div>

      <div class="chat-header-actions">
        <button class="header-action-btn icon-btn" title="Call">
          <img src="/icons/call-icon.png" alt="Call" class="action-icon" />
        </button>
        <button class="header-action-btn icon-btn" title="Video call">
          <img src="/icons/video-call-icon.png" alt="Video Call" class="action-icon" />
        </button>
        <button class="header-action-btn icon-btn" @click="handleDeleteConversation" title="Delete conversation">
          <img src="/icons/trashbin-icon.png" alt="Delete" class="action-icon" />
        </button>
        <button class="header-action-btn" title="Info">
          ⓘ
        </button>
      </div>
    </div>

    <div class="messages-area">
      <div class="messages-container">
        <MessageItem
          v-for="message in messages"
          :key="message.id"
          :message="message"
          :is-own-message="message.senderId === currentUserId"
          @unsend="handleUnsend"
        />
      </div>
    </div>

    <div class="chat-input-area">
      <div class="input-actions">
        <button class="input-action-btn icon-btn" title="Add image">
          <img src="/icons/gallery-icon.png" alt="Gallery" class="action-icon" />
        </button>
        <button class="input-action-btn icon-btn" title="Add sticker">
          <img src="/icons/sticker-icon.png" alt="Sticker" class="action-icon" />
        </button>
      </div>

      <div class="input-wrapper">
        <input
          v-model="messageInput"
          type="text"
          placeholder="Message..."
          class="message-input"
          @keydown.enter="handleSendMessage"
        />
      </div>

      <button
        v-if="messageInput.trim()"
        class="send-btn"
        @click="handleSendMessage"
        title="Send message"
      >
        ➤
      </button>
    </div>
  </div>

  <div v-else class="chat-empty-state">
    <p>Select a conversation to start messaging</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Conversation, Message } from '../types/chat'
import MessageItem from './MessageItem.vue'

interface Props {
  selectedConversation: Conversation | null
  messages: Message[]
  currentUserId: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'send-message': [content: string]
  'unsend-message': [messageId: string]
  'delete-conversation': []
}>()

// --- FIX: Computed Property for Chat Partner ---
// This safely finds the "other person" and prevents undefined errors
const chatPartner = computed(() => {
  // Safety check: if conversation or participants array is missing
  if (!props.selectedConversation || !props.selectedConversation.participants?.length) {
    return null
  }

  // Find the participant who is NOT me
  const partner = props.selectedConversation.participants.find(p => p.id !== props.currentUserId)

  // Return partner, or fallback to the first person if (e.g. self-chat)
  return partner || props.selectedConversation.participants[0]
})

const messageInput = ref('')

const handleSendMessage = () => {
  if (messageInput.value.trim()) {
    emit('send-message', messageInput.value)
    messageInput.value = ''
  }
}

const handleUnsend = (messageId: string) => {
  emit('unsend-message', messageId)
}

const handleDeleteConversation = () => {
  emit('delete-conversation')
}
</script>

<style scoped>
.chat-window {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1a1a1a;
}

.chat-empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #a0a0a0;
  font-size: 16px;
}

/* Chat Header */
.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #262626;
  flex-shrink: 0;
}

.chat-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.header-avatar {
  position: relative;
  flex-shrink: 0;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  object-fit: cover;
}

.online-indicator {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 12px;
  height: 12px;
  background: #31a24c;
  border: 2px solid #1a1a1a;
  border-radius: 50%;
}

.header-info {
  min-width: 0;
}

.chat-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.chat-subtitle {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: #a0a0a0;
}

.chat-header-actions {
  display: flex;
  gap: 12px;
}

/* Icon Buttons */
.icon-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
}

.icon-btn:hover {
  transform: scale(1.1);
}

.action-icon {
  width: 24px;
  height: 24px;
  object-fit: contain;
  /* Optional: Invert color for dark mode if icons are black */
  /* filter: invert(1); */
}

.header-action-btn {
  background: none;
  border: none;
  color: #0084ff;
  cursor: pointer;
  font-size: 18px;
  padding: 8px;
  transition: transform 0.2s ease;
}

.header-action-btn:hover {
  transform: scale(1.1);
}

/* Messages Area */
.messages-area {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.messages-container {
  display: flex;
  flex-direction: column;
  padding: 20px;
  gap: 8px;
  margin-top: auto;
}

.messages-area::-webkit-scrollbar {
  width: 8px;
}

.messages-area::-webkit-scrollbar-track {
  background: transparent;
}

.messages-area::-webkit-scrollbar-thumb {
  background: #404040;
  border-radius: 4px;
}

/* Chat Input */
.chat-input-area {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #262626;
  flex-shrink: 0;
  background: #1a1a1a;
}

.input-actions {
  display: flex;
  gap: 8px;
}

.input-action-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  transition: transform 0.2s ease;
}

.input-action-btn:hover {
  transform: scale(1.1);
}

.input-wrapper {
  flex: 1;
}

.message-input {
  width: 100%;
  background: #262626;
  border: 1px solid #404040;
  color: #fff;
  padding: 10px 16px;
  border-radius: 20px;
  font-size: 14px;
  font-family: inherit;
  transition: border-color 0.2s ease;
}

.message-input:focus {
  outline: none;
  border-color: #0084ff;
}

.message-input::placeholder {
  color: #808080;
}

.send-btn {
  background: none;
  border: none;
  color: #0084ff;
  cursor: pointer;
  font-size: 18px;
  padding: 8px;
  transition: transform 0.2s ease;
}

.send-btn:hover {
  transform: scale(1.1);
}
</style>