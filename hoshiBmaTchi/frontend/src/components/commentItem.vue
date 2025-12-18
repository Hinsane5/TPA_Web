<script setup lang="ts">
import { ref, computed } from "vue";
import { formatDistanceToNow } from "date-fns";
import { useRouter } from 'vue-router';
import { usersApi } from '../services/apiService';
import { parseRichText } from "../utils/textParser";

const props = withDefaults(
  defineProps<{
    username?: string;
    profileImage?: string;
    commentText: string;
    timestamp: string;
    likes?: number;
    repliesCount?: number;
    isVerified?: boolean;
  }>(),
  {
    username: "User",
    profileImage: "",
    likes: 0,
    repliesCount: 0,
  }
);

const router = useRouter();

const emit = defineEmits(["like", "reply", "click-user"]);

const isLiked = ref(false);
const showReplies = ref(false);

const formattedTime = computed(() => {
  if (!props.timestamp) return "";
  try {
    return (
      formatDistanceToNow(new Date(props.timestamp), {
        addSuffix: false,
      }).replace("about ", "") + " ago"
    );
  } catch {
    return "";
  }
});

const toggleLike = () => {
  isLiked.value = !isLiked.value
  emit('like', isLiked.value)
}

const toggleReplies = () => {
  showReplies.value = !showReplies.value
}

const handleCommentClick = async (event: MouseEvent) => {
  const target = event.target as HTMLElement;

  if (target.classList.contains("mention-link")) {
    const rawUsername = target.dataset.username;
    
    if (rawUsername) {
      const username = rawUsername.startsWith('@') ? rawUsername.substring(1) : rawUsername;
      
      try {
        const response = await usersApi.searchUsers(username);
        const users = response.data.users || response.data.data || [];
        const foundUser = users.find((u: any) => u.username.toLowerCase() === username.toLowerCase());
        
        if (foundUser && (foundUser.user_id || foundUser.id)) {
          const userId = foundUser.user_id || foundUser.id;
          router.push(`/dashboard/profile/${userId}`);
        }
      } catch (error) {
        console.error("Failed to resolve mention:", error);
      }
    }
  }

  if (target.classList.contains("hashtag-link")) {
    const tag = target.dataset.tag;
    if (tag) {
      router.push({ path: '/dashboard/explore', query: { q: tag } });
    }
  }
};

</script>

<template>
  <div class="comment-item">
    <img 
      :src="profileImage || '/default-avatar.png'" 
      :alt="username"
      class="comment-avatar"
      @click="$emit('click-user')"
    />
    
    <div class="comment-content">
      <div class="comment-text-wrapper">
        <div class="comment-header-row">
          <span class="comment-username" @click="$emit('click-user')">{{ username }}</span>
          <img 
            v-if="isVerified" 
            src="/icons/verified-icon.png" 
            class="verified-badge-tiny" 
            alt="Verified"
          />
        </div>
        <span 
          class="comment-body" 
          @click="handleCommentClick"
          v-html="parseRichText(commentText)"
        ></span>
      </div>

      <div class="comment-metadata">
        <span class="comment-time">{{ formattedTime }}</span>
        <span v-if="likes > 0" class="comment-likes">{{ likes }} likes</span>
        <button class="reply-action" @click="$emit('reply')">Reply</button>
      </div>

      <div v-if="repliesCount > 0" class="replies-toggle">
        <div class="divider-line"></div>
        <button class="replies-button" @click="toggleReplies">
          {{ showReplies ? 'Hide' : 'View' }} {{ repliesCount }} {{ repliesCount === 1 ? 'reply' : 'replies' }}
        </button>
      </div>
    </div>

    <button class="like-button" @click="toggleLike">
      <img 
        :src="isLiked ? '/icons/liked-icon.png' : '/icons/notifications-icon.png'"
        class="action-icon"
        :class="{ active: isLiked }"
      />
    </button>
  </div>
</template>

<style scoped>
.comment-item {
  display: flex;
  gap: 12px;
  padding: 0 16px;
  align-items: flex-start;
}

.comment-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
  margin-top: 4px;
}

.comment-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.comment-text-wrapper {
  background-color: #202327;
  border-radius: 16px;
  padding: 8px 12px;
}

.comment-text {
  font-size: 13px;
  color: #fff;
  margin: 0;
  line-height: 1.4;
}

.comment-username {
  font-weight: 600;
  color: #fff;
}

.comment-body {
  color: #fff;
  margin-left: 6px;
}

.comment-metadata {
  display: flex;
  gap: 12px;
  align-items: center;
  font-size: 12px;
  color: #fff;
  padding: 0 12px;
}

.comment-time {
  cursor: default;
}

.comment-likes {
  cursor: pointer;
  transition: color 0.2s ease;
}

.comment-likes:hover {
  color: #000000;
}

.replies-button {
  background: none;
  border: none;
  padding: 0;
  color: #5b5bff;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.2s ease;
  font-family: inherit;
}

.replies-button:hover {
  color: #4949ff;
}

.replies-section {
  padding: 8px 12px;
  font-size: 12px;
  color: #65676b;
}

.replies-placeholder {
  margin: 0;
}

.comment-actions {
  display: flex;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.comment-item:hover .comment-actions {
  opacity: 1;
}

.like-button,
.reply-button {
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background-color 0.2s ease;
}

.like-button:hover,
.reply-button:hover {
  background-color: #f0f0f0;
}

.action-icon {
  width: 16px;
  height: 16px;
  opacity: 1;
  transition: opacity 0.2s ease;
}

.action-icon.active {
  opacity: 1;
}

.reply-button {
  font-size: 14px;
  color: #65676b;
}

.comment-header-row {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-right: 6px;
}

.verified-badge-tiny {
  width: 12px;
  height: 12px;
  vertical-align: middle;
}

@media (max-width: 768px) {
  .comment-item {
    padding: 0 12px;
    gap: 10px;
  }

  .comment-text-wrapper {
    padding: 6px 10px;
  }

  .comment-actions {
    opacity: 1;
  }
}
</style>
