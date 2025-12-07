<template>
  <div class="post-container">
    <div class="post-header">
      <div class="header-left">
        <img
          v-if="post.profile_picture"
          :src="post.profile_picture"
          class="avatar-img"
          alt="Profile"
        />
        <div v-else class="avatar">{{ getInitials(post.username) }}</div>

        <div class="user-info">
          <p class="username">{{ post.username }}</p>
          <p class="timestamp">{{ formatTime(post.created_at) }}</p>
        </div>
      </div>

      <button class="more-button">
        <svg class="dots-icon" fill="currentColor" viewBox="0 0 24 24">
          <circle cx="6" cy="12" r="2" />
          <circle cx="12" cy="12" r="2" />
          <circle cx="18" cy="12" r="2" />
        </svg>
      </button>
    </div>

    <div class="post-media" @dblclick="$emit('toggle-like', post)">
      <div class="media-item-wrapper" v-if="currentMedia">
        <video
          v-if="currentMedia.media_type.startsWith('video/')"
          :src="getDisplayUrl(currentMedia.media_url)"
          controls
          loop
          muted
          class="post-image"
        ></video>
        <img
          v-else
          :src="getDisplayUrl(currentMedia.media_url)"
          alt="Post content"
          class="post-image"
        />
      </div>

      <button
        v-if="hasMultiple && currentIndex > 0"
        class="nav-btn left"
        @click.stop="currentIndex--"
      >
        ❮
      </button>

      <button
        v-if="hasMultiple && currentIndex < mediaList.length - 1"
        class="nav-btn right"
        @click.stop="currentIndex++"
      >
        ❯
      </button>

      <div v-if="hasMultiple" class="dots-container">
        <div
          v-for="(_, idx) in mediaList"
          :key="idx"
          class="dot"
          :class="{ active: idx === currentIndex }"
        ></div>
      </div>
    </div>

    <div class="post-actions">
      <div class="actions-left">
        <button
          @click="$emit('toggle-like', post)"
          class="action-button"
          :title="post.is_liked ? 'Unlike' : 'Like'"
        >
          <img
            :src="
              post.is_liked
                ? '/icons/liked-icon.png'
                : '/icons/notifications-icon.png'
            "
            alt="Like"
            class="action-icon"
            :class="{ liked: post.is_liked }"
          />
        </button>
        <button
          class="action-button"
          title="Comment"
          @click="$emit('open-detail', post)"
        >
          <img
            src="/icons/comment-icon.png"
            alt="Comment"
            class="action-icon"
          />
        </button>
        <button class="action-button" title="Share">
          <img src="/icons/share-icon.png" alt="Share" class="action-icon" />
        </button>
      </div>

      <button
        @click="toggleSave"
        class="action-button"
        :title="isSaved ? 'Unsave' : 'Save'"
      >
        <img
          :src="isSaved ? '/icons/saved-icon.png' : '/icons/save-icon.png'"
          alt="Save"
          class="action-icon"
          :class="{ saved: isSaved }"
        />
      </button>
    </div>

    <div class="post-footer">
      <p class="likes-count">
        {{ post.likes_count ? post.likes_count.toLocaleString() : 0 }} likes
      </p>

      <div class="caption-section">
        <p class="caption-text" @click="handleCaptionClick">
          <span class="caption-username">
            {{ post.username }}
          </span>
          <span
            class="caption-content"
            v-html="parseCaption(post.caption)"
          ></span>
        </p>
      </div>

      <button
        v-if="post.comments_count > 0"
        class="view-comments"
        @click="$emit('open-detail', post)"
      >
        View all {{ post.comments_count }} comments
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps, computed } from "vue";
import { formatDistanceToNow } from "date-fns";
import { postsApi, usersApi } from "../services/apiService";
import { useRouter } from 'vue-router';
const router = useRouter();

const props = defineProps({
  post: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["open-detail", "toggle-like"]);


const isSaved = ref(false);
const currentIndex = ref(0);


const mediaList = computed(() => {

  if (
    props.post.media &&
    Array.isArray(props.post.media) &&
    props.post.media.length > 0
  ) {
    return props.post.media;
  }

  if (props.post.media_url) {
    return [
      {
        media_url: props.post.media_url,
        media_type: props.post.media_type || "image/jpeg",
      },
    ];
  }
  return [];
});

const currentMedia = computed(() => {
  if (mediaList.value.length === 0) return null;
  return mediaList.value[currentIndex.value];
});

const hasMultiple = computed(() => mediaList.value.length > 1);

const toggleSave = async () => {
  try {
    isSaved.value = !isSaved.value;
    await postsApi.toggleSavePost(props.post.id);
  } catch (error) {
    isSaved.value = !isSaved.value;
    console.error("Failed to save post", error);
  }
};

const getDisplayUrl = (url: string) => {
  if (!url) return "/placeholder.png";


  return url
    .replace("http://minio:9000", "http://localhost:9000")
    .replace("http://host.docker.internal:9000", "http://localhost:9000")
    .replace("http://backend:9000", "http://localhost:9000");
};

const getInitials = (username: string) => {
  return username ? username.substring(0, 2).toUpperCase() : "UN";
};

const formatTime = (dateStr: string) => {
  if (!dateStr) return "";
  try {
    return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
  } catch (e) {
    return "";
  }
};

const handleCaptionClick = async (event: MouseEvent) => {
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

const parseCaption = (text: string) => {
  if (!text) return "";
  return text.replace(
    /(@[a-zA-Z0-9._]+)/g,
    '<span class="mention-link" data-username="$1" style="color: rgb(0, 149, 246); cursor: pointer; font-weight: 600;">$1</span>'
  );
};
</script>

<style scoped>
/* Main container */
.post-container {
  width: 100%;
  max-width: 480px;
  background-color: #000000;
  border: 1px solid #262626; /* Darker border for dark mode */
  border-radius: 8px;
  margin-bottom: 24px;
  overflow: hidden;
  color: #ffffff;
}

/* Header Section */
.post-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #262626;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #a78bfa 0%, #ec4899 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
}

.avatar-img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.username {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.timestamp {
  font-size: 12px;
  color: #a8a8a8;
  margin: 0;
}

.more-button {
  background: none;
  border: none;
  padding: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.dots-icon {
  width: 20px;
  height: 20px;
}

/* Media Section */
.post-media {
  position: relative; /* Needed for absolute positioning of arrows/dots */
  width: 100%;
  aspect-ratio: 1; /* Square post */
  background-color: #121212;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.media-item-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.post-image {
  width: 100%;
  height: 100%;
  object-fit: cover; /* Maintains aspect ratio while filling square */
}

/* --- CAROUSEL CONTROLS --- */
.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(26, 26, 26, 0.8);
  color: white;
  border: none;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  z-index: 10;
  transition: opacity 0.2s;
}

.nav-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.left {
  left: 10px;
}
.right {
  right: 10px;
}

.dots-container {
  position: absolute;
  bottom: 15px;
  width: 100%;
  display: flex;
  justify-content: center;
  gap: 6px;
  pointer-events: none; /* Let clicks pass through */
}

.dot {
  width: 6px;
  height: 6px;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 50%;
  transition: background 0.2s;
}

.dot.active {
  background: #3badf8; /* Blue active dot like Instagram */
}

/* Action Bar Section */
.post-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}

.actions-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-button {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.1s ease;
}

.action-button:active {
  transform: scale(0.9);
}

.action-icon {
  width: 24px;
  height: 24px;
}

/* Specific fixes for icons that might already be colored/white */
.action-icon.liked {
  filter: none; /* Don't invert the red heart */
}
.action-icon.saved {
  filter: invert(1); /* Keep white */
}

/* Footer Section */
.post-footer {
  padding: 0 16px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.likes-count {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin: 0;
}

.caption-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.caption-text {
  font-size: 14px;
  color: #fff;
  margin: 0;
  line-height: 1.4;
}

.caption-username {
  font-weight: 600;
  color: #fff;
  margin-right: 6px;
}

.caption-content {
  color: #fff;
}

.view-comments {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  font-size: 14px;
  color: #a8a8a8;
  text-align: left;
  margin-top: 4px;
}

.view-comments:hover {
  color: #fff;
}
</style>
