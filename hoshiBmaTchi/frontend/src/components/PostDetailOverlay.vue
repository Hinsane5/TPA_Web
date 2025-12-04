<template>
  <div v-if="isOpen" class="overlay-backdrop" @click="$emit('close')">
    <div class="overlay-container" @click.stop>
      <div class="overlay-left">
        <div class="media-container" @dblclick="$emit('toggle-like', post)">
          <div v-if="currentMedia" class="media-wrapper">
            <video
              v-if="
                currentMedia.media_type &&
                currentMedia.media_type.startsWith('video/')
              "
              :src="getDisplayUrl(currentMedia.media_url)"
              controls
              autoplay
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
      </div>

      <div class="overlay-right">
        <div class="overlay-header">
          <div class="header-content">
            <img
              :src="post.profile_picture || '/placeholder.png'"
              class="author-avatar"
            />
            <div class="author-info">
              <p class="author-username">{{ post.username }}</p>
              <p
                v-if="!isOwnPost && !isFollowing"
                class="follow-text"
                @click="handleFollow"
              >
                • Follow
              </p>
              <p
                v-if="!isOwnPost && isFollowing"
                class="following-text"
                @click="handleUnfollow"
              >
                • Following
              </p>
            </div>
          </div>
          <button class="close-button" @click="$emit('close')">✕</button>
        </div>

        <div class="comments-section">
          <CommentItem
            v-if="post.caption"
            :username="post.username"
            :profile-image="post.profile_picture"
            :comment-text="post.caption"
            :timestamp="post.created_at"
            :likes="post.likes_count"
          />

          <div v-if="loadingComments" class="loading-spinner">
            <div class="spinner"></div>
          </div>

          <CommentItem
            v-for="comment in enrichedComments"
            :key="comment.id"
            :username="comment.username"
            :profile-image="comment.profile_picture"
            :comment-text="comment.content"
            :timestamp="comment.created_at"
            @reply="handleReply(comment.username)"
          />

          <div
            v-if="
              !loadingComments && enrichedComments.length === 0 && !post.caption
            "
            class="no-comments"
          >
            <p>No comments yet.</p>
          </div>
        </div>

        <div class="divider"></div>

        <div class="action-section-wrapper">
          <div class="action-icons">
            <div class="icons-left">
              <button class="icon-button" @click="$emit('toggle-like', post)">
                <img
                  :src="
                    post.is_liked
                      ? '/icons/liked-icon.png'
                      : '/icons/notifications-icon.png'
                  "
                  class="icon"
                  :class="{ active: post.is_liked }"
                />
              </button>
              <button class="icon-button" @click="focusInput">
                <img src="/icons/comment-icon.png" class="icon" />
              </button>
              <button class="icon-button">
                <img src="/icons/share-icon.png" class="icon" />
              </button>
            </div>
            <button class="icon-button" @click="toggleSave">
              <img
                :src="
                  isSaved ? '/icons/saved-icon.png' : '/icons/save-icon.png'
                "
                class="icon"
              />
            </button>
          </div>
          <p class="likes-text">{{ post.likes_count }} likes</p>
          <p class="date-text">{{ formatFullDate(post.created_at) }}</p>
        </div>

        <div class="divider"></div>

        <div class="comment-input-section">
          <input
            ref="commentInputRef"
            v-model="commentText"
            placeholder="Add a comment..."
            class="comment-input"
            @keyup.enter="submitComment"
          />
          <button
            v-if="commentText.trim()"
            @click="submitComment"
            class="post-button"
            :disabled="isSubmitting"
          >
            Post
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { format } from "date-fns";
import { postsApi, usersApi } from "@/services/apiService";
import CommentItem from "./commentItem.vue";

const props = defineProps(["isOpen", "post"]);
const emit = defineEmits(["close", "comment-added", "toggle-like"]);

const rawComments = ref<any[]>([]);
const userCache = ref(new Map());
const commentText = ref("");
const loadingComments = ref(true);
const isSubmitting = ref(false);
const commentInputRef = ref<HTMLInputElement | null>(null);
const isSaved = ref(false);
const isFollowing = ref(false);

const currentIndex = ref(0);

watch(
  () => props.post.id,
  () => {
    currentIndex.value = 0;
  }
);

const currentUserId = localStorage.getItem("userID");
const currentUsername = localStorage.getItem("username") || "me";
const currentUserPic = localStorage.getItem("profilePicture") || "";
const isOwnPost = props.post.user_id === currentUserId;

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

const getDisplayUrl = (url: string) => {
  if (!url) return "/placeholder.png";
  return url.replace("http://minio:9000", "http://localhost:9000");
};

const enrichedComments = computed(() => {
  return rawComments.value.map((comment) => {
    const user = userCache.value.get(comment.user_id);
    return {
      ...comment,
      username: user ? user.username : "Loading...",
      profile_picture: user ? user.profile_picture_url : null,
    };
  });
});

onMounted(async () => {
  document.body.style.overflow = "hidden";

  if (!isOwnPost) {
    try {
      const profileRes = await usersApi.getUserProfile(props.post.user_id);
      isFollowing.value = profileRes.data.is_following;
    } catch (e) {}
  }

  try {
    const res = await postsApi.getCommentForPost(props.post.id);
    rawComments.value = res.data || [];

    const userIdsToFetch = new Set(rawComments.value.map((c) => c.user_id));
    const fetchPromises = Array.from(userIdsToFetch).map(
      async (uid: string) => {
        if (userCache.value.has(uid)) return;
        try {
          const userRes = await usersApi.getUserProfile(uid);
          userCache.value.set(uid, userRes.data);
        } catch (e) {
          userCache.value.set(uid, {
            username: "Unknown",
            profile_picture_url: "",
          });
        }
      }
    );

    await Promise.all(fetchPromises);
    userCache.value = new Map(userCache.value);
  } catch (e) {
    console.error(e);
  } finally {
    loadingComments.value = false;
  }
});

const submitComment = async () => {
  if (!commentText.value.trim()) return;
  const tempText = commentText.value;
  commentText.value = "";
  isSubmitting.value = true;

  const fakeId = Date.now().toString();
  rawComments.value.push({
    id: fakeId,
    user_id: currentUserId,
    content: tempText,
    created_at: new Date().toISOString(),
  });

  if (currentUserId && !userCache.value.has(currentUserId)) {
    userCache.value.set(currentUserId, {
      username: currentUsername,
      profile_picture_url: currentUserPic,
    });
  }

  try {
    await postsApi.createComment(props.post.id, tempText);
    emit("comment-added");
  } catch (e) {
    rawComments.value = rawComments.value.filter((c) => c.id !== fakeId);
    commentText.value = tempText;
    alert("Failed to post comment");
  } finally {
    isSubmitting.value = false;
  }
};

const handleReply = (username: string) => {
  commentText.value = `@${username} `;
  focusInput();
};

const focusInput = () => commentInputRef.value?.focus();

const handleFollow = async () => {
  isFollowing.value = true;
  try {
    await usersApi.followUser(props.post.user_id);
  } catch {
    isFollowing.value = false;
  }
};

const handleUnfollow = async () => {
  isFollowing.value = false;
  try {
    await usersApi.unfollowUser(props.post.user_id);
  } catch {
    isFollowing.value = true;
  }
};

const toggleSave = async () => {
  try {
    isSaved.value = !isSaved.value;
    
    await postsApi.toggleSavePost(props.post.id);
  } catch (error) {
    isSaved.value = !isSaved.value;
    console.error("Failed to toggle save:", error);
  }
};

const formatFullDate = (d: string | number | Date) => {
  try {
    return format(new Date(d), "MMMM d, yyyy");
  } catch {
    return "";
  }
};
</script>

<style scoped>
/* --- LAYOUT --- */
.overlay-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.overlay-container {
  display: flex;
  width: 100%;
  max-width: 1100px;
  height: 85vh;
  background: #000;
  border-radius: 4px;
  overflow: hidden;
}

/* --- LEFT SIDE (MEDIA) --- */
.overlay-left {
  flex: 1.5;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative; /* Ensures buttons can be absolute relative to this area */
}

.media-container {
  width: 100%;
  height: 100%;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.media-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.post-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

/* --- CAROUSEL CONTROLS (Matches CreatePostOverlay Design) --- */
.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(26, 26, 26, 0.8);
  color: white;
  border: none;
  width: 32px; /* Was 30px in previous code, restored to 32px */
  height: 32px; /* Was 30px, restored to 32px */
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px; /* Larger arrow */
  z-index: 10;
  transition: background 0.2s;
}

.nav-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.left {
  left: 12px;
} /* Was 10px, restored to 12px */
.right {
  right: 12px;
} /* Was 10px, restored to 12px */

.dots-container {
  position: absolute;
  bottom: 15px;
  display: flex;
  gap: 6px;
  z-index: 10;
}

.dot {
  width: 6px;
  height: 6px;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 50%;
  cursor: pointer;
  transition: background 0.2s;
}

.dot.active {
  background: #fff;
  transform: scale(1.2);
}

/* --- RIGHT SIDE (DETAILS) --- */
.overlay-right {
  flex: 1;
  background: #000;
  border-left: 1px solid #262626;
  display: flex;
  flex-direction: column;
  min-width: 400px;
}

.overlay-header {
  padding: 14px 16px;
  border-bottom: 1px solid #262626;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.author-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
}

.author-info {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 14px;
}

.author-username {
  font-weight: 600;
  color: white;
  margin: 0;
}
.follow-text {
  color: #0095f6;
  font-weight: 600;
  cursor: pointer;
  margin: 0;
}
.following-text {
  color: #a0a0a0;
  font-weight: 600;
  cursor: pointer;
  margin: 0;
}
.close-button {
  background: none;
  border: none;
  color: white;
  font-size: 20px;
  cursor: pointer;
}

.comments-section {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.divider {
  height: 1px;
  background: #262626;
}

.action-section-wrapper {
  padding: 14px 16px;
}
.action-icons {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}
.icons-left {
  display: flex;
  gap: 16px;
}
.icon-button {
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
}
.icon {
  width: 24px;
  height: 24px;
  /* filter: invert(1); */
}
.icon.active {
  filter: none;
}
.likes-text {
  font-weight: 600;
  margin-bottom: 4px;
  font-size: 14px;
}
.date-text {
  color: #a8a8a8;
  font-size: 12px;
  margin-top: 4px;
}

.comment-input-section {
  padding: 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.comment-input {
  flex: 1;
  background: none;
  border: none;
  color: white;
  font-size: 14px;
}
.comment-input:focus {
  outline: none;
}
.post-button {
  background: none;
  border: none;
  color: #0095f6;
  font-weight: 600;
  cursor: pointer;
}
.post-button:disabled {
  opacity: 0.5;
}

.loading-spinner {
  display: flex;
  justify-content: center;
  padding: 20px;
}
.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #262626;
  border-top: 2px solid #fff;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}
@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
.no-comments {
  text-align: center;
  color: #a0a0a0;
  margin-top: 20px;
}
</style>
