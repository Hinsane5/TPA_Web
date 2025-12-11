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
        <p 
          v-if="message.messageType === 'text'" 
          class="message-text"
          @click="handleMessageClick"
          v-html="parseMessage(message.content)"
        ></p>
        <div v-else-if="message.messageType === 'image'" class="message-media">
          <img
            :src="getDisplayUrl(message.content)"
            :alt="message.senderName"
            class="media-image"
          />
        </div>
        <div v-else-if="message.messageType === 'video'" class="message-media">
          <video 
            :src="getDisplayUrl(message.content)" 
            class="media-video" 
            controls
          ></video>
        </div>
        <div v-else-if="message.messageType === 'gif'" class="message-media">
          <img
            :src="message.content"
            :alt="message.senderName"
            class="media-gif"
          />
        </div>

        <div v-else-if="isSharedContent" class="shared-content-card" @click="viewSharedContent">
          <div class="shared-media-preview">
            <video
                v-if="isSharedVideo"
                :src="getDisplayUrl(message.mediaUrl)"
                class="shared-thumbnail"
                muted
                autoplay
                loop
                playsinline
              ></video>
            <img 
              v-if="message.mediaUrl" 
              :src="getDisplayUrl(message.mediaUrl)" 
              class="shared-thumbnail" 
            />
            <div v-else class="shared-placeholder">
              <span class="placeholder-icon">{{ getShareIcon }}</span>
            </div>
            
            <div class="type-badge">
              <img :src="`/icons/${getShareTypeIcon}-icon.png`" class="badge-icon" />
            </div>
          </div>
          
          <div class="shared-info">
            <span class="shared-title">Shared {{ getShareTypeName }}</span>
            <span class="view-link">Tap to view</span>
          </div>
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
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { usersApi } from "../services/apiService";
import router from "@/router";

interface Props {
  message: Message;
  isOwnMessage: boolean;
}

const props = defineProps<Props>();
defineEmits<{
  unsend: [messageId: string];
}>();

const formatTime = (date: string | Date) => {
  return new Date(date).toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: true,
  });
};

const parseMessage = (text: string) => {
  if (!text) return "";
  return text.replace(
    /(@[a-zA-Z0-9._]+)/g,
    '<span class="mention-link" data-username="$1" style="color: rgb(0, 149, 246); cursor: pointer; font-weight: 600;">$1</span>'
  );
};

const isSharedContent = computed(() => {
  return ['post_share', 'reel_share', 'story_share'].includes(props.message.messageType);
});

const handleMessageClick = async (event: MouseEvent) => {
  const target = event.target as HTMLElement;

  if (target.classList.contains("mention-link")) {
    const rawUsername = target.dataset.username;
    
    if (rawUsername) {
      const username = rawUsername.substring(1);
      
      try {
        const response = await usersApi.searchUsers(username);
        const users = response.data.users || [];
        const foundUser = users.find((u: any) => u.username === username);
        
        if (foundUser && foundUser.user_id) {
          router.push(`/dashboard/profile/${foundUser.user_id}`);
        } else {
          console.warn("User not found for mention:", username);
        }
      } catch (error) {
        console.error("Failed to resolve mention in chat:", error);
      }
    }
  }
};

const getShareTypeName = computed(() => {
  if (props.message.messageType === 'post_share') return 'Post';
  if (props.message.messageType === 'reel_share') return 'Reel';
  if (props.message.messageType === 'story_share') return 'Story';
  return 'Content';
});

const getShareTypeIcon = computed(() => {
  if (props.message.messageType === 'reel_share') return 'reels';
  return 'post';
});

const getShareIcon = computed(() => {
  if (props.message.messageType === 'post_share') return '📸';
  if (props.message.messageType === 'reel_share') return '🎬';
  if (props.message.messageType === 'story_share') return '⭕';
  return '🔗';
});

const getDisplayUrl = (url: string | undefined) => {
  if (!url) return "";
  return url
    .replace("http://minio:9000", "http://localhost:9000")
    .replace("http://backend:9000", "http://localhost:9000");
};

const viewSharedContent = () => {
  const contentId = props.message.content; // Content ID is stored in message content for shares
  
  if (props.message.messageType === 'post_share') {
    // Navigate to post (or open overlay logic if implemented)
    // router.push(`/p/${contentId}`); 
    alert("Navigating to Post: " + contentId);
  } else if (props.message.messageType === 'reel_share') {
    // router.push(`/reels/${contentId}`);
    alert("Navigating to Reel: " + contentId);
  } else if (props.message.messageType === 'story_share') {
     alert("View Story: " + contentId);
  }
};

const isSharedVideo = computed(() => {
  if (props.message.messageType === 'reel_share') return true;
  
  if (props.message.mediaUrl) {
    const url = props.message.mediaUrl.toLowerCase();
    return url.endsWith('.mp4') || url.endsWith('.mov') || url.endsWith('.webm');
  }
  return false;
});

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
