<template>
  <CallOverlay 
    :active="callState !== 'idle'"
    :status="callState"
    :call-type="activeCallType"
    :caller-name="incomingCaller?.name"
    :caller-avatar="incomingCaller?.avatar"
    :remote-users="remoteUsers"
    :audio-enabled="isAudioEnabled"
    :video-enabled="isVideoEnabled"
    @accept="acceptCall"
    @reject="leaveCall"
    @end="leaveCall"
    @toggle-audio="toggleAudio"
    @toggle-video="toggleVideo"
  />

  <div v-if="selectedConversation && chatPartner" class="chat-window">
    <div class="chat-header">
      <div class="chat-header-left">
        <div class="header-avatar">
          <img
            v-if="!selectedConversation.isGroup"
            :src="chatPartner?.avatar || '/placeholder.svg'"
            class="avatar"
          />
          <div v-else class="group-avatar-placeholder">👥</div>
        </div>
        <div class="header-info">
          <h3 class="chat-title">
            {{ selectedConversation.isGroup ? selectedConversation.name : chatPartner?.fullName }}
          </h3>
          <p class="chat-subtitle" v-if="!selectedConversation.isGroup">
             {{ chatPartner?.isOnline ? 'Active now' : '' }}
          </p>
          <p class="chat-subtitle" v-else>
             {{ selectedConversation.participants.length }} members
          </p>
        </div>
      </div>

      <div class="chat-header-actions">
        <button class="header-action-btn icon-btn" title="Call" @click="startCall('audio')">
          <img src="/icons/call-icon.png" alt="Call" class="action-icon" />
        </button>
        <button class="header-action-btn icon-btn" title="Video call" @click="startCall('video')">
          <img src="/icons/video-call-icon.png" alt="Video Call" class="action-icon" />
        </button>
        
        <button class="header-action-btn icon-btn" @click="handleDeleteConversation" title="Delete conversation">
          <img src="/icons/trashbin-icon.png" alt="Delete" class="action-icon" />
        </button>
        <button class="header-action-btn" title="Info" @click="showDetails = true">
          ℹ️
        </button>
      </div>
    </div>

    <div class="messages-area" ref="messagesContainer">
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

    <div v-if="showGifPicker" class="gif-picker">
      <div class="gif-header">
        <div class="gif-search-wrapper">
          <input 
            v-model="gifSearchQuery" 
            placeholder="Search GIPHY..." 
            class="gif-search-input"
          />
        </div>
        <button @click="showGifPicker = false" class="close-gif">✕</button>
      </div>

      <div class="gif-grid">
        <div v-if="isLoadingGifs" class="loading-spinner">Loading...</div>
        
        <img 
          v-for="gif in gifResults" 
          :key="gif.id" 
          :src="gif.images.fixed_height.url" 
          class="gif-option"
          @click="sendGif(gif.images.fixed_height.url)"
        />
      </div>
      
      <div class="giphy-attribution">Powered by GIPHY</div>
    </div>

    <div class="chat-input-area">
      <div class="input-actions">
        <button class="input-action-btn icon-btn" title="Add image">
          <img src="/icons/gallery-icon.png" alt="Gallery" class="action-icon" />
        </button>
        <button class="input-action-btn icon-btn" title="Add GIF" @click="showGifPicker = !showGifPicker">
          <span style="font-weight: 800; font-size: 10px; border: 1.5px solid currentColor; border-radius: 4px; padding: 2px;">GIF</span>
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

    <GroupDetailsModal
        v-if="showDetails"
        :conversation="selectedConversation"
        :current-user-id="currentUserId"
        @close="showDetails = false"
        @leave="handleLeave"
        @refresh="emit('refresh-data')" 
    />

  </div>

  <div v-else class="chat-empty-state">
    <p>Select a conversation to start messaging</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import type { Conversation, Message } from '../types/chat'
import MessageItem from './MessageItem.vue'
import GroupDetailsModal from './GroupDetailsModal.vue';
// 3. ADD THIS: Import CallOverlay and Store
import CallOverlay from './CallOverlay.vue';
import { useChatStore } from '../composables/useChatStore';

// 4. ADD THIS: Use Store logic
const chatStore = useChatStore();
const { 
  callState, 
  activeCallType, 
  incomingCaller, 
  remoteUsers,
  isAudioEnabled,
  isVideoEnabled,
  startCall,
  acceptCall,
  leaveCall,
  toggleAudio,
  toggleVideo
} = chatStore;

interface Props {
  selectedConversation: Conversation | null
  messages: Message[]
  currentUserId: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'send-message': [content: string, type?: "text" | "image" | "gif", mediaUrl?: string]
  'unsend-message': [messageId: string]
  'delete-conversation': []
  'refresh-data': [] 
}>()

// --- Refs ---
const showGifPicker = ref(false)
const gifSearchQuery = ref('')
const gifResults = ref<any[]>([])
const isLoadingGifs = ref(false)
const messagesContainer = ref<HTMLElement | null>(null) 
const messageInput = ref('')
const showDetails = ref(false);

const chatPartner = computed(() => {
  if (!props.selectedConversation || !props.selectedConversation.participants?.length) {
    return null
  }
  const partner = props.selectedConversation.participants.find(p => p.id !== props.currentUserId)
  return partner || props.selectedConversation.participants[0]
})

// --- Auto Scroll Logic ---
const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

watch(() => props.messages, () => {
  scrollToBottom()
}, { deep: true })

watch(() => props.selectedConversation, () => {
  scrollToBottom()
})

// --- GIF Logic ---
const GIPHY_API_KEY = 'YiObN9bdRusPcIp6YFM5EhpxjZeFl5rA'; 

const fetchTrendingGifs = async () => {
  if (isLoadingGifs.value) return;
  isLoadingGifs.value = true;
  try {
    const res = await fetch(`https://api.giphy.com/v1/gifs/trending?api_key=${GIPHY_API_KEY}&limit=20&rating=g`);
    const data = await res.json();
    gifResults.value = data.data;
  } catch (err) {
    console.error("Failed to fetch GIFs", err);
  } finally {
    isLoadingGifs.value = false;
  }
};

const searchGifs = async () => {
  if (!gifSearchQuery.value.trim()) {
    fetchTrendingGifs();
    return;
  }
  isLoadingGifs.value = true;
  try {
    const res = await fetch(`https://api.giphy.com/v1/gifs/search?api_key=${GIPHY_API_KEY}&q=${gifSearchQuery.value}&limit=20&rating=g`);
    const data = await res.json();
    gifResults.value = data.data;
  } catch (err) {
    console.error("Search failed", err);
  } finally {
    isLoadingGifs.value = false;
  }
};

let debounceTimer: any;
watch(gifSearchQuery, (newVal) => {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    searchGifs();
  }, 500);
});

watch(showGifPicker, (isOpen) => {
  if (isOpen && gifResults.value.length === 0) {
    fetchTrendingGifs();
  }
});

const sendGif = (url: string) => {
    emit('send-message', 'GIF', 'gif', url)
    showGifPicker.value = false
}

// --- Message Handlers ---
const handleSendMessage = () => {
  if (messageInput.value.trim()) {
    emit('send-message', messageInput.value)
    messageInput.value = ''
    scrollToBottom()
  }
}

const handleUnsend = (messageId: string) => {
  emit('unsend-message', messageId)
}

const handleDeleteConversation = () => {
  emit('delete-conversation')
}

const handleLeave = () => {
    emit('delete-conversation'); 
    showDetails.value = false;
}
</script>

<style scoped>
/* Keep existing styles */
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

.gif-picker {
    position: absolute;
    bottom: 80px;
    left: 20px;
    width: 320px;
    height: 400px;
    background: #262626;
    border: 1px solid #404040;
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.5);
    z-index: 100;
    display: flex;
    flex-direction: column;
}

.gif-header {
    padding: 12px;
    border-bottom: 1px solid #404040;
    display: flex;
    gap: 8px;
    align-items: center;
}

.gif-search-wrapper {
  flex: 1;
}

.gif-search-input {
  width: 100%;
  background: #1a1a1a;
  border: 1px solid #404040;
  color: #fff;
  padding: 6px 12px;
  border-radius: 8px;
  font-size: 13px;
}

.gif-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    padding: 12px;
    overflow-y: auto;
    flex: 1;
}

.gif-option {
    width: 100%;
    height: 100px;
    object-fit: cover;
    border-radius: 6px;
    cursor: pointer;
    background: #333;
}

.giphy-attribution {
  font-size: 10px;
  color: #808080;
  text-align: center;
  padding: 4px;
  background: #000;
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

.group-avatar-placeholder {
  width: 48px;
  height: 48px;

  display: flex;
  align-items: center;
  justify-content: center;

  background: #333;
  border-radius: 50%;
  font-size: 24px;
}
</style>