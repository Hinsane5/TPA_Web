<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from "vue";
import { format } from "date-fns";
import { postsApi, usersApi } from "@/services/apiService";
import ShareModal from "./ShareModal.vue";
import CommentItem from "./commentItem.vue";

const props = defineProps(["isOpen", "post"]);
const emit = defineEmits(["close", "comment-added", "toggle-like", "post-deleted"]);

const rawComments = ref<any[]>([]);
const userCache = ref(new Map());
const commentText = ref("");
const loadingComments = ref(true);
const isSubmitting = ref(false);
const commentInputRef = ref<HTMLInputElement | null>(null);
const isSaved = ref(false);
const isFollowing = ref(false);

const currentIndex = ref(0);

const savedCollectionId = ref<string | null>(null);
const showPopover = ref(false);
const showCreateInput = ref(false);
const collections = ref<any[]>([]);
const newCollectionName = ref("");
const createInputRef = ref<HTMLInputElement | null>(null);

const currentUsername = localStorage.getItem("username") || "me";
const currentUserPic = localStorage.getItem("profilePicture") || "";
const showShareModal = ref(false);

// Report Logic State
const showMenu = ref(false);
const showReportModal = ref(false);
const reportReason = ref("");

watch(
  () => props.post.id,
  () => {
    currentIndex.value = 0;
    if (props.post.is_saved !== undefined) {
      isSaved.value = props.post.is_saved;
    }
  }
);

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

const handleDelete = async () => {
  if (!confirm("Are you sure you want to delete this post?")) return;

  try {
    await postsApi.deletePost(props.post.id);
    emit("post-deleted", props.post.id);
    emit("close"); 
  } catch (error) {
    console.error("Failed to delete post:", error);
    alert("Failed to delete post.");
  }
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

  if (!isOwnPost.value) {
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

const handleMouseEnter = () => {
  showPopover.value = true;
  fetchCollections();
};

const handleMouseLeave = () => {
  showPopover.value = false;
  showCreateInput.value = false;
  newCollectionName.value = "";
};

const fetchCollections = async () => {
  try {
    const res = await postsApi.getUserCollections();
    collections.value = res.data.collections || res.data || [];
  } catch (error) {
    console.error("Failed to fetch collections", error);
  }
};

const toggleDefaultSave = async () => {
  if (isSaved.value) {
    await toggleSaveApi("");
    isSaved.value = false;
    savedCollectionId.value = null;
  } else {
    await toggleSaveApi("");
    isSaved.value = true;
    savedCollectionId.value = "";
  }
};

const saveToCollection = async (collectionId: string) => {
  if (isSaved.value && savedCollectionId.value === collectionId) {
    await toggleSaveApi(collectionId);
    isSaved.value = false;
    savedCollectionId.value = null;
  } else {
    await toggleSaveApi(collectionId);
    isSaved.value = true;
    savedCollectionId.value = collectionId;
  }
  showPopover.value = false;
};

const toggleSaveApi = async (collectionId: string) => {
  try {
    await postsApi.toggleSavePost(props.post.id, collectionId);
  } catch (error) {
    console.error("Failed to toggle save:", error);
  }
};

const enableCreateMode = () => {
  showCreateInput.value = true;
  nextTick(() => {
    createInputRef.value?.focus();
  });
};

const createNewCollection = async () => {
  if (!newCollectionName.value.trim()) return;
  try {
    const res = await postsApi.createCollection(newCollectionName.value);
    await fetchCollections();
    
    // Auto save to new collection
    const newId = res.data.id || res.data.collection_id;
    if (newId) {
      await saveToCollection(newId);
    }
    
    newCollectionName.value = "";
    showCreateInput.value = false;
  } catch (error) {
    console.error("Failed to create collection", error);
  }
};
const getCollectionCover = (col: any) => {
  // Use array from backend if available
  if (col.cover_images && col.cover_images.length > 0) return col.cover_images[0];
  // Fallback
  if (col.saved_posts && col.saved_posts.length > 0) {
     const p = col.saved_posts[0].post || col.saved_posts[0];
     if (p.media && p.media.length > 0) return p.media[0].media_url;
  }
  return null;
};


const formatFullDate = (d: string | number | Date) => {
  try {
    return format(new Date(d), "MMMM d, yyyy");
  } catch {
    return "";
  }
};

const getUserIdFromToken = (): string | null => {
  const token = localStorage.getItem("accessToken");
  if (!token) return null;

  try {
    const parts = token.split(".");
    if (parts.length < 2) return null;
    
    const payloadPart = parts[1];
    if (!payloadPart) return null;

    const payload = JSON.parse(atob(payloadPart));
    const id = payload.user_id || payload.sub || payload.id;
    return id ? String(id) : null;
  } catch (e) {
    console.error("Error parsing token:", e);
    return null;
  }
};

const currentUserId = getUserIdFromToken();

// --- 3. FIX OWNERSHIP CHECK ---
const isOwnPost = computed(() => {
  // Convert both to string to be safe
  return String(props.post.user_id) === String(currentUserId);
});

// Report Logic
const handleReport = () => {
    showMenu.value = false;
    showReportModal.value = true;
};

const submitReport = async () => {
    try {
        await postsApi.reportPost(props.post.id, reportReason.value);
        alert("Post reported successfully. Our team will review it.");
        showReportModal.value = false;
        reportReason.value = "";
    } catch (e) {
        alert("Failed to report post.");
        console.error(e);
    }
};
</script>

<template>
  <div v-if="isOpen" class="overlay-backdrop" @click.self="$emit('close')">
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
                <img 
                  v-if="post.is_verified" 
                  src="/icons/verified-icon.png" 
                  alt="Verified" 
                  class="verified-badge-small"
                />
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
          
          <div class="header-actions">
            <button class="icon-button menu-btn" @click.stop="showMenu = !showMenu">•••</button>
            <button class="close-button" @click="$emit('close')">✕</button>
            
            <div v-if="showMenu" class="dropdown-menu">
                <button 
                  v-if="isOwnPost" 
                  class="dropdown-item danger" 
                  @click="handleDelete"
                >
                  Remove Post
                </button>

                <button 
                  v-else 
                  class="dropdown-item danger" 
                  @click="handleReport"
                >
                  Report Post
                </button>
            </div>
          </div>
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
              <button class="icon-button" @click="showShareModal = true">
                <img src="/icons/share-icon.png" class="icon" />
              </button>
            </div>

            <div
              class="save-wrapper"
              @mouseenter="handleMouseEnter"
              @mouseleave="handleMouseLeave"
            >
              <button
                class="icon-button"
                :title="isSaved ? 'Unsave' : 'Save'"
                @click="toggleDefaultSave"
              >
                <img
                  :src="
                    isSaved ? '/icons/saved-icon.png' : '/icons/save-icon.png'
                  "
                  class="icon"
                  :class="{ saved: isSaved }"
                />
              </button>

              <transition name="fade">
                <div v-if="showPopover" class="save-popover">
                  <div class="popover-header">Save to...</div>

                  <div class="collections-list">
                    <div class="popover-item" @click="saveToCollection('')">
                      <div class="popover-thumb">
                        <img src="/icons/save-icon.png" class="thumb-icon" />
                      </div>
                      <span class="popover-name">All Posts</span>
                      <span
                        v-if="savedCollectionId === '' && isSaved"
                        class="check-mark"
                        >✓</span
                      >
                    </div>

                    <div
                      v-for="col in collections"
                      :key="col.id"
                      class="popover-item"
                      @click="saveToCollection(col.id)"
                    >
                      <div class="popover-thumb">
                        <img
                          v-if="getCollectionCover(col)"
                          :src="getDisplayUrl(getCollectionCover(col))"
                          class="cover-img"
                        />
                        <div v-else class="empty-cover"></div>
                      </div>
                      <span class="popover-name">{{ col.name }}</span>
                      <span
                        v-if="savedCollectionId === col.id && isSaved"
                        class="check-mark"
                        >✓</span
                      >
                    </div>
                  </div>

                  <div class="popover-footer">
                    <div v-if="showCreateInput" class="create-input-wrapper">
                      <input
                        ref="createInputRef"
                        v-model="newCollectionName"
                        placeholder="Collection Name"
                        class="mini-input"
                        @keyup.enter="createNewCollection"
                      />
                      <button class="mini-btn" @click="createNewCollection">
                        Add
                      </button>
                    </div>

                    <div
                      v-else
                      class="popover-item add-item"
                      @click="enableCreateMode"
                    >
                      <div class="plus-icon-wrapper">+</div>
                      <span class="popover-name">New Collection</span>
                    </div>
                  </div>
                </div>
              </transition>
            </div>
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
            class="post-button"
            :disabled="isSubmitting"
            @click="submitComment"
          >
            Post
          </button>
        </div>
      </div>
    </div>

    <div v-if="showReportModal" class="report-modal-backdrop" @click.self="showReportModal = false">
        <div class="report-modal">
            <h3>Report Post</h3>
            <textarea v-model="reportReason" placeholder="Why are you reporting this post?" rows="4"></textarea>
            <div class="modal-buttons">
                <button class="btn-submit" @click="submitReport">Submit Report</button>
                <button class="btn-cancel" @click="showReportModal = false">Cancel</button>
            </div>
        </div>
    </div>

    <ShareModal 
      v-if="showShareModal"
      :content-id="post.id"
      type="post"
      :thumbnail="currentMedia ? getDisplayUrl(currentMedia.media_url) : ''"
      @close="showShareModal = false"
    />

  </div>
</template>

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
  width: 32px; 
  height: 32px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px; 
  z-index: 10;
  transition: background 0.2s;
}

.nav-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

.left {
  left: 12px;
} 
.right {
  right: 12px;
} 

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

.header-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    position: relative;
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

.username-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
}

.verified-badge-small {
  width: 14px;
  height: 14px;
  object-fit: contain;
  vertical-align: middle;
}

.menu-btn {
    color: white;
    font-size: 16px;
    cursor: pointer;
}

.dropdown-menu {
    position: absolute;
    right: 0;
    top: 40px;
    background: #262626;
    border: 1px solid #363636;
    border-radius: 4px;
    z-index: 100;
    width: 150px;
}

.dropdown-item {
    display: block;
    width: 100%;
    padding: 10px;
    color: white;
    background: none;
    border: none;
    cursor: pointer;
    text-align: left;
    font-size: 14px;
}

.dropdown-item:hover {
    background: #363636;
}

.dropdown-item.danger {
    color: #ed4956;
    font-weight: bold;
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

.save-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.save-popover {
  position: absolute;
  bottom: 40px; /* Above the icon */
  right: 0;
  width: 240px;
  background-color: #262626;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  z-index: 2000; /* Higher index for overlay context */
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid #363636;
}

.popover-header {
  padding: 10px 12px;
  font-size: 14px;
  font-weight: 600;
  border-bottom: 1px solid #363636;
  color: #e0e0e0;
}

.collections-list {
  max-height: 200px;
  overflow-y: auto;
}

.popover-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.2s;
  color: #fff;
}

.popover-item:hover {
  background-color: #3a3a3a;
}

.popover-thumb {
  width: 32px;
  height: 32px;
  background: #121212;
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #363636;
}

.cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.thumb-icon {
  width: 16px;
  height: 16px;
  filter: invert(1);
}

.popover-name {
  font-size: 13px;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.check-mark {
  color: #0095f6;
  font-weight: bold;
}

.popover-footer {
  border-top: 1px solid #363636;
  padding: 4px 0;
}

.add-item {
  color: #0095f6;
}

.plus-icon-wrapper {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  font-weight: 300;
  border: 1px solid #363636;
  border-radius: 4px;
}

.create-input-wrapper {
  padding: 8px 12px;
  display: flex;
  gap: 6px;
}

.mini-input {
  flex: 1;
  background: #121212;
  border: 1px solid #363636;
  border-radius: 4px;
  color: #fff;
  padding: 4px 8px;
  font-size: 13px;
  outline: none;
}

.mini-input:focus {
  border-color: #0095f6;
}

.mini-btn {
  background: #0095f6;
  color: #fff;
  border: none;
  border-radius: 4px;
  padding: 0 10px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.mini-btn:hover {
  background: #007bb5;
}

/* Report Modal Styles */
.report-modal-backdrop {
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.7);
    z-index: 2000;
    display: flex;
    justify-content: center;
    align-items: center;
}

.report-modal {
    background: #262626;
    padding: 20px;
    border-radius: 12px;
    width: 400px;
    color: white;
    text-align: center;
}

.report-modal h3 {
    margin-top: 0;
    margin-bottom: 15px;
    font-weight: 600;
}

.report-modal textarea {
    width: 100%;
    background: #121212;
    border: 1px solid #363636;
    color: white;
    border-radius: 4px;
    padding: 10px;
    resize: none;
    font-family: inherit;
    margin-bottom: 20px;
}

.modal-buttons {
    display: flex;
    justify-content: center;
    gap: 10px;
}

.btn-submit {
    background: #ed4956;
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 4px;
    font-weight: 600;
    cursor: pointer;
}

.btn-cancel {
    background: transparent;
    color: #e0e0e0;
    border: 1px solid #363636;
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(5px);
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