<template>
  <div class="reels-page-container">
    <div class="reels-scroll-container" ref="scrollContainer">
      <div 
        v-for="(reel, index) in reels" 
        :key="reel.id" 
        class="reel-item"
        :class="{ active: currentReelIndex === index }"
        ref="reelItems"
      >
        <div class="video-wrapper" @click="togglePlay(index)">
          <video
            ref="videoRefs"
            class="reel-video"
            :src="getDisplayUrl(reel.video_url)"
            :poster="getDisplayUrl(reel.thumbnail_url)"
            loop="false"
            muted
            playsinline
            @ended="handleVideoEnd(index)"
          ></video>
          
          <div v-if="!isPlaying[index]" class="play-overlay">
            <svg viewBox="0 0 24 24" class="play-icon" fill="white">
              <path d="M8 5v14l11-7z"/>
            </svg>
          </div>
        </div>

        <div class="reel-overlay">
          
          <div class="reel-info">
            <div class="user-profile" @click.stop="goToProfile(reel.user.id)">
              <img :src="reel.user.profile_picture || '/default-avatar.png'" class="avatar" alt="User">
              <span class="username">{{ reel.user.username }}</span>
              <button class="follow-btn" v-if="!reel.user.is_following">• Follow</button>
            </div>
            
            <div class="caption-container">
              <p class="caption">
                <span v-html="parseCaption(reel.caption)"></span>
              </p>
            </div>
            
            <div class="music-tag">
              <svg viewBox="0 0 24 24" class="music-icon" fill="white"><path d="M12 3v10.55c-.59-.34-1.27-.55-2-.55-2.21 0-4 1.79-4 4s1.79 4 4 4 4-1.79 4-4V7h4V3h-6z"/></svg>
              <span>Original Audio - {{ reel.user.username }}</span>
            </div>
          </div>

          <div class="reel-actions">
            <button class="action-btn" @click.stop="toggleLike(reel)">
              <img 
                :src="reel.is_liked ? '/icons/liked-icon.png' : '/icons/notifications-icon.png'" 
                class="action-icon"
                :class="{ 'is-liked': reel.is_liked }"
              />
              <span class="count">{{ formatCount(reel.likes_count) }}</span>
            </button>

            <button class="action-btn" @click.stop="openComments(reel)">
              <img src="/icons/comment-icon.png" class="action-icon" />
              <span class="count">{{ formatCount(reel.comments_count) }}</span>
            </button>

            <button class="action-btn" @click.stop="openShareModal(reel)">
              <img src="/icons/share-icon.png" class="action-icon" />
            </button>

            <button class="action-btn" @click.stop="toggleSave(reel)">
              <img 
                :src="reel.is_saved ? '/icons/saved-icon.png' : '/icons/save-icon.png'" 
                class="action-icon"
                :class="{ 'is-saved': reel.is_saved }"
              />
            </button>

            <button class="action-btn" @click.stop="openMoreOptions(reel)">
               <svg viewBox="0 0 24 24" class="action-icon svg-icon" fill="white">
                 <circle cx="12" cy="12" r="2"/>
                 <circle cx="12" cy="6" r="2"/>
                 <circle cx="12" cy="18" r="2"/>
               </svg>
            </button>
          </div>
        </div>
      </div>

      <div v-if="loading" class="reel-item skeleton">
        <div class="skeleton-avatar"></div>
        <div class="skeleton-text"></div>
        <div class="skeleton-text short"></div>
      </div>
    </div>

   <PostDetailOverlay 
      v-if="showCommentOverlay" 
      :is-open="showCommentOverlay"
      :post="activeReelForComments" 
      @close="closeCommentOverlay"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { reelsApi, usersApi } from '../services/apiService';
import { formatDistanceToNow } from 'date-fns';
import PostDetailOverlay from './PostDetailOverlay.vue';

const showCommentOverlay = ref(false);
const activeReelForComments = ref<any>(null);

const router = useRouter();

// --- State ---
const reels = ref<any[]>([]);
const loading = ref(false);
const currentReelIndex = ref(0);
const isPlaying = ref<boolean[]>([]);
const scrollContainer = ref<HTMLElement | null>(null);
const reelItems = ref<HTMLElement[]>([]);
const videoRefs = ref<HTMLVideoElement[]>([]);
const observer = ref<IntersectionObserver | null>(null);

// Pagination
const limit = 5;
const offset = ref(0);
const hasMore = ref(true);

// --- Lifecycle ---
onMounted(() => {
  fetchReels();
  setupIntersectionObserver();
});

onBeforeUnmount(() => {
  if (observer.value) observer.value.disconnect();
});

// --- Data Fetching ---
const fetchReels = async () => {
  if (loading.value || !hasMore.value) return;
  
  loading.value = true;
  try {
    const response = await reelsApi.getReelsFeed(limit, offset.value);
    const newReels = response.data.data || []; // Adjust based on actual API response structure
    
    if (newReels.length < limit) hasMore.value = false;
    
    // Initialize playing state for new reels
    const startIdx = reels.value.length;
    newReels.forEach(() => isPlaying.value.push(false));
    
    reels.value.push(...newReels);
    offset.value += limit;
    
    // Re-observe new elements after DOM update
    nextTick(() => {
      observeNewItems(startIdx);
    });
    
  } catch (error) {
    console.error("Failed to fetch reels", error);
  } finally {
    loading.value = false;
  }
};

// --- Intersection Observer (Auto Play/Pause) ---
const setupIntersectionObserver = () => {
  const options = {
    root: scrollContainer.value,
    threshold: 0.6 // Trigger when 60% of the reel is visible
  };

  observer.value = new IntersectionObserver((entries) => {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        // Find index of the intersecting element
        const index = reelItems.value.indexOf(entry.target as HTMLElement);
        if (index !== -1) {
          playVideo(index);
          
          // Load more if we are near the end
          if (index >= reels.value.length - 2) {
            fetchReels();
          }
        }
      } else {
        const index = reelItems.value.indexOf(entry.target as HTMLElement);
        if (index !== -1) pauseVideo(index);
      }
    });
  }, options);
};

const observeNewItems = (startIndex: number) => {
  if (!observer.value) return;
  for (let i = startIndex; i < reelItems.value.length; i++) {
    const item = reelItems.value[i];
    if (item) {
        observer.value.observe(item);
    }
  }
};

// --- Video Control ---
const playVideo = (index: number) => {
  currentReelIndex.value = index;
  const video = videoRefs.value[index];
  if (video) {
    // Pause all others first (safety check)
    videoRefs.value.forEach((v, i) => {
      if (i !== index && v) {
        v.pause();
        isPlaying.value[i] = false;
      }
    });
    
    video.muted = false; // Try to unmute, browser might block if no interaction
    video.play().then(() => {
      isPlaying.value[index] = true;
    }).catch(e => {
      console.warn("Autoplay blocked:", e);
      video.muted = true; // Fallback to muted autoplay
      video.play();
    });
  }
};

const pauseVideo = (index: number) => {
  const video = videoRefs.value[index];
  if (video) {
    video.pause();
    isPlaying.value[index] = false;
  }
};

const togglePlay = (index: number) => {
  const video = videoRefs.value[index];
  if (video) {
    if (video.paused) {
      video.play();
      isPlaying.value[index] = true;
    } else {
      video.pause();
      isPlaying.value[index] = false;
    }
  }
};

const handleVideoEnd = (index: number) => {
  if (index < reels.value.length - 1) {
    const nextReel = reelItems.value[index + 1];
    if (nextReel) {
      nextReel.scrollIntoView({ behavior: 'smooth' });
    }
  } else {
    const video = videoRefs.value[index];
    if (video) {
        video.play();
    }
  }
};

const toggleLike = async (reel: any) => {
  reel.is_liked = !reel.is_liked;
  reel.likes_count += reel.is_liked ? 1 : -1;

  try {
    if (reel.is_liked) {
      await reelsApi.likeReel(reel.id);
    } else {
      await reelsApi.unlikeReel(reel.id);
    }
  } catch (e) {
    reel.is_liked = !reel.is_liked;
    reel.likes_count += reel.is_liked ? 1 : -1;
  }
};

const toggleSave = async (reel: any) => {
  reel.is_saved = !reel.is_saved;
  try {
    if (reel.is_saved) {
      await reelsApi.saveReel(reel.id);
    } else {
      await reelsApi.unsaveReel(reel.id);
    }
  } catch (e) {
    reel.is_saved = !reel.is_saved;
  }
};

const openComments = (reel: any) => {
  activeReelForComments.value = {
    ...reel,
    user_id: reel.user.id,              
    username: reel.user.username,        
    profile_picture: reel.user.profile_picture,
    created_at: reel.created_at || new Date().toISOString()
  };
  
  showCommentOverlay.value = true;
};

const closeCommentOverlay = () => {
  showCommentOverlay.value = false;
  activeReelForComments.value = null;
};

const openShareModal = (reel: any) => {
  const recipient = prompt("Enter recipient username (Mock Share):");
  if (recipient) {
    alert(`Reel shared to ${recipient}`);
  }
};

const openMoreOptions = (reel: any) => {
  const choice = confirm("Options: \nOK: Copy Link\nCancel: Report");
  if (choice) {
    navigator.clipboard.writeText(`http://hoshibmatchi.com/reels/${reel.id}`);
    alert("Link copied!");
  } else {
    const reason = prompt("Reason for reporting:");
    if (reason) alert("Report submitted to Admin.");
  }
};

// --- Utilities ---
const getDisplayUrl = (url: string) => {
  if (!url) return "";
  // Adjust for local docker environment if needed
  return url.replace("http://minio:9000", "http://localhost:9000"); 
};

const formatCount = (count: number) => {
  if (count >= 1000000) return (count / 1000000).toFixed(1) + 'M';
  if (count >= 1000) return (count / 1000).toFixed(1) + 'K';
  return count.toString();
};

const goToProfile = (userId: string) => {
  router.push(`/dashboard/profile/${userId}`);
};

const parseCaption = (text: string) => {
  if (!text) return "";
  // Simple hashtag/mention parser
  return text.replace(/(@\w+)/g, '<span style="font-weight:bold; cursor:pointer">$1</span>')
             .replace(/(#\w+)/g, '<span style="font-weight:bold; cursor:pointer">$1</span>');
};

</script>

<style scoped>
.reels-page-container {
  background-color: #000;
  height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
}

/* Scroll Container with Snap */
.reels-scroll-container {
  height: 100%;
  width: 100%;
  max-width: 450px; /* Mobile width constraint for desktop view */
  overflow-y: scroll;
  scroll-snap-type: y mandatory;
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
  position: relative;
}

.reels-scroll-container::-webkit-scrollbar {
  display: none;
}

/* Individual Reel Item */
.reel-item {
  height: 100vh;
  width: 100%;
  scroll-snap-align: start;
  scroll-snap-stop: always;
  position: relative;
  background-color: #121212;
  border-bottom: 1px solid #262626;
}

/* Video */
.video-wrapper {
  height: 100%;
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.reel-video {
  height: 100%;
  width: 100%;
  object-fit: cover;
  cursor: pointer;
}

.play-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(0,0,0,0.4);
  border-radius: 50%;
  padding: 20px;
  pointer-events: none;
}

.play-icon {
  width: 40px;
  height: 40px;
}

/* Overlay UI */
.reel-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  padding: 20px;
  background: linear-gradient(to top, rgba(0,0,0,0.8), transparent);
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  box-sizing: border-box;
}

/* Left Info */
.reel-info {
  display: flex;
  flex-direction: column;
  gap: 10px;
  color: white;
  width: 80%;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid white;
}

.username {
  font-weight: bold;
  font-size: 14px;
}

.follow-btn {
  background: none;
  border: 1px solid white;
  color: white;
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 12px;
  cursor: pointer;
}

.caption-container {
  max-height: 100px;
  overflow-y: auto;
}

.caption {
  font-size: 14px;
  margin: 0;
}

.music-tag {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
}

.music-icon {
  width: 12px;
  height: 12px;
}

/* Right Actions */
.reel-actions {
  display: flex;
  flex-direction: column;
  gap: 20px;
  align-items: center;
  padding-bottom: 10px;
}

.action-btn {
  background: none;
  border: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  color: white;
}

.action-icon {
  width: 28px;
  height: 28px;
  filter: drop-shadow(0 0 5px rgba(0,0,0,0.5));
}

.svg-icon {
  width: 28px;
  height: 28px;
  fill: white;
}

/* Active States */
.is-liked {
  /* Assuming the liked icon is already red/filled in the image asset */
  filter: none; 
}

.is-saved {
  filter: invert(1);
}

.count {
  font-size: 12px;
  font-weight: 600;
}

/* Skeleton */
.skeleton {
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: 20px;
  background: #222;
}

.skeleton-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #333;
  margin-bottom: 10px;
  animation: pulse 1.5s infinite;
}

.skeleton-text {
  height: 15px;
  width: 70%;
  background: #333;
  margin-bottom: 5px;
  border-radius: 4px;
  animation: pulse 1.5s infinite;
}

.skeleton-text.short {
  width: 40%;
}

@keyframes pulse {
  0% { opacity: 0.5; }
  50% { opacity: 1; }
  100% { opacity: 0.5; }
}
</style>