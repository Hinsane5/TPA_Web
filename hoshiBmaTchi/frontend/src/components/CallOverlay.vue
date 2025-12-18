<script setup lang="ts">
import { computed } from 'vue';
import { useChatStore } from '../composables/useChatStore'; // Import the store

interface Props {
  active: boolean;
  status: 'idle' | 'incoming' | 'connected' | 'dialing';
  callType: 'audio' | 'video';
  callerName?: string;
  callerAvatar?: string;
  remoteUsers: any[];
  audioEnabled: boolean;
  videoEnabled: boolean;
}

const props = defineProps<Props>();
defineEmits(['accept', 'reject', 'end', 'toggle-audio', 'toggle-video']);

const chatStore = useChatStore();

// FIX: Compute the display name for the remote user
const remoteDisplayName = computed(() => {
  if (chatStore.incomingCaller.value) {
      return chatStore.incomingCaller.value.name;
  }
  
  if (chatStore.outgoingCallInfo.value) {
      return chatStore.outgoingCallInfo.value.name;
  }

  if (chatStore.selectedConversation.value) {
    return chatStore.selectedConversation.value.name;
  }
  
  // 4. Last resort
  return "Unknown User";
});

const gridClass = computed(() => {
  const count = props.remoteUsers.length + 1;
  if (count <= 2) return 'grid-1-1';
  if (count <= 4) return 'grid-2-2';
  return 'grid-auto';
});
</script>

<template>
  <div v-if="active" class="call-overlay">
    <div v-if="status === 'incoming'" class="incoming-modal">
      <div class="caller-info">
        <div class="avatar-pulse">
           <img :src="callerAvatar || '/placeholder.svg'" class="caller-avatar" />
        </div>
        <h3>{{ callerName }}</h3>
        <p>Incoming {{ callType }} Call...</p>
      </div>
      <div class="actions">
        <button class="btn-reject" @click="$emit('reject')">
          <img src="/icons/call-icon.png" alt="Decline" width="24" style="transform: rotate(135deg); filter: invert(1);"/>
        </button>
        <button class="btn-accept" @click="$emit('accept')">
          <img src="/icons/call-icon.png" alt="Accept" width="24"/>
        </button>
      </div>
    </div>

    <div v-else class="active-call">
      <div class="video-grid" :class="gridClass">
        <div id="local-player" class="video-container local">
           <p class="user-label">You</p>
        </div>
        
        <div v-for="(user, index) in remoteUsers" :id="'remote-player-' + user.uid" :key="user.uid" class="video-container">
           <p class="user-label">
             {{ remoteUsers.length === 1 ? remoteDisplayName : `User ${user.uid}` }}
           </p>
        </div>
      </div>

      <div class="controls-bar">
        <button class="control-btn" :class="{ 'off': !audioEnabled }" @click="$emit('toggle-audio')">
           Mic {{ audioEnabled ? 'On' : 'Off' }}
        </button>
        
        <button v-if="callType === 'video'" class="control-btn" :class="{ 'off': !videoEnabled }" @click="$emit('toggle-video')">
           Cam {{ videoEnabled ? 'On' : 'Off' }}
        </button>
        
        <button class="control-btn end-call" @click="$emit('end')">
           End Call
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.call-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.9);
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.incoming-modal {
  text-align: center;
  color: white;
}
/* Caller Avatar */
.caller-avatar {
  width: 100px;
  height: 100px;

  border-radius: 50%;
  object-fit: cover;
  border: 4px solid #0084ff;
}

.avatar-pulse {
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(0, 132, 255, 0.7);
  }

  70% {
    transform: scale(1.1);
    box-shadow: 0 0 0 20px rgba(0, 132, 255, 0);
  }

  100% {
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(0, 132, 255, 0);
  }
}

/* Call Actions */
.actions {
  display: flex;
  justify-content: center;
  gap: 40px;
  margin-top: 30px;
}

.btn-reject,
.btn-accept {
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 50%;
  cursor: pointer;
}

.btn-reject {
  background: #ff4444;
}

.btn-accept {
  background: #00c853;
}

/* Active Call Layout */
.active-call {
  width: 100%;
  height: 100%;

  display: flex;
  flex-direction: column;
}

.video-grid {
  flex: 1;
  display: grid;
  gap: 10px;
  padding: 20px;
}

.grid-1-1 {
  grid-template-columns: 1fr 1fr;
}

.grid-auto {
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
}

.video-container {
  background: #333;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.user-label {
  position: absolute; 
  bottom: 10px; 
  left: 10px;
  background: rgba(0,0,0,0.5); 
  color: white; 
  padding: 4px 8px; 
  border-radius: 4px;
  font-size: 0.9rem;
  z-index: 10;
}

.controls-bar {
  height: 80px;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  background: #1a1a1a;
}
.control-btn {
  padding: 10px 20px;
  border-radius: 30px;
  border: none;
  background: #404040;
  color: white;
  cursor: pointer;
  font-weight: 500;
  transition: background 0.2s;
}
.control-btn:hover {
    background: #505050;
}
.control-btn.off { 
    background: #ff4444; 
}

.control-btn.end-call { 
    background: #cc0000; 
    font-weight: bold; 
}
</style>