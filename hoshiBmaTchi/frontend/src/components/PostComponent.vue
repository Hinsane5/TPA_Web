<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from "vue";
import { formatDistanceToNow } from "date-fns";
import { postsApi, usersApi } from "../services/apiService";
import { useRouter } from 'vue-router';
import ShareModal from "./ShareModal.vue";
import { aiApi } from "../services/apiService";
import { useAuth } from "../composables/useAuth";

const router = useRouter();

const props = defineProps({
  post: { type: Object, required: true },
});
const emit = defineEmits(["open-detail", "toggle-like", "post-deleted"]);

const isSaved = ref(false);
const savedCollectionId = ref<string | null>(null); 
const showPopover = ref(false);
const showCreateInput = ref(false);
const collections = ref<any[]>([]);
const newCollectionName = ref("");
const createInputRef = ref<HTMLInputElement | null>(null);
const currentIndex = ref(0);
const showShareModal = ref(false);

const isSummarized = ref(false);
const summaryText = ref("");
const isLoadingSummary = ref(false);

const { user } = useAuth(); 
const showMenu = ref(false);
const showReportModal = ref(false);
const reportReason = ref("");

onMounted(() => {
  if (props.post.is_saved !== undefined) {
    isSaved.value = props.post.is_saved;
  }
});

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

const isOwner = computed(() => {
  return user.value?.id === props.post.user_id;
});

const closeMenu = () => {
  showMenu.value = false;
};

const handleDelete = async () => {
  if (!confirm("Are you sure you want to delete this post?")) return;
  
  try {
    await postsApi.deletePost(props.post.id);
    emit("post-deleted", props.post.id);
    showMenu.value = false;
  } catch (error) {
    console.error("Failed to delete post:", error);
    alert("Failed to delete post.");
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

const openReportModal = () => {
  showMenu.value = false;
  showReportModal.value = true;
};

const submitReport = async () => {
  if (!reportReason.value.trim()) return;
  
  try {
    await postsApi.reportPost(props.post.id, reportReason.value);
    alert("Post reported. Thank you.");
    showReportModal.value = false;
    reportReason.value = "";
  } catch (error) {
    console.error("Report failed", error);
    alert("Failed to report post");
  }
};

const vClickOutside = {
  mounted(el: any, binding: any) {
    el.clickOutsideEvent = (event: Event) => {
      if (!(el === event.target || el.contains(event.target))) {
        binding.value(event);
      }
    };
    document.body.addEventListener('click', el.clickOutsideEvent);
  },
  unmounted(el: any) {
    document.body.removeEventListener('click', el.clickOutsideEvent);
  },
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
    console.error("Failed to toggle save", error);
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

const mediaList = computed(() => {
  if (props.post.media?.length > 0) return props.post.media;
  if (props.post.media_url) {
    return [{ media_url: props.post.media_url, media_type: props.post.media_type || "image/jpeg" }];
  }
  return [];
});

const currentMedia = computed(() => mediaList.value[currentIndex.value] || null);
const hasMultiple = computed(() => mediaList.value.length > 1);

const getDisplayUrl = (url: string) => {
  if (!url) return "/placeholder.png";
  return url
    .replace("http://minio:9000", "http://localhost:9000")
    .replace("http://host.docker.internal:9000", "http://localhost:9000")
    .replace("http://backend:9000", "http://localhost:9000");
};

const getCollectionCover = (col: any) => {
  if (col.cover_images && col.cover_images.length > 0) return col.cover_images[0];
  if (col.saved_posts && col.saved_posts.length > 0) {
    const p = col.saved_posts[0].post || col.saved_posts[0];
    if (p.media?.length > 0) return p.media[0].media_url;
  }
  return null;
};

const getInitials = (username: string) => username ? username.substring(0, 2).toUpperCase() : "UN";
const formatTime = (dateStr: string) => {
  try { return dateStr ? formatDistanceToNow(new Date(dateStr), { addSuffix: true }) : ""; } 
  catch { return ""; }
};

const handleCaptionClick = async (event: MouseEvent) => {
  const target = event.target as HTMLElement;

  // Handle Mentions
  if (target.classList.contains("mention-link")) {
    const rawUsername = target.dataset.username;
    if (rawUsername) {
      const username = rawUsername.substring(1); 
      try {
        const response = await usersApi.searchUsers(username);
        const user = response.data.users?.find((u: any) => u.username === username);
        if (user?.user_id) router.push(`/dashboard/profile/${user.user_id}`);
      } catch (e) { console.error(e); }
    }
  }

  // Handle Hashtags -> ADD THIS
  if (target.classList.contains("hashtag-link")) {
    const rawTag = target.dataset.tag; // e.g. "#test"
    if (rawTag) {
      // Navigate to explore page with the tag as a query
      router.push({ path: '/dashboard/explore', query: { q: rawTag } });
    }
  }
};

const parseCaption = (text: string) => {
  if (!text) return "";
  
  // Basic HTML escape to prevent XSS before we add our own tags
  let content = text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");

  content = content.replace(
    /#([a-zA-Z0-9_]+)/g,
    '<span class="hashtag-link" data-tag="$1" style="color: #00376b; cursor: pointer; font-weight: 400;">#$1</span>'
  );

  content = content.replace(
    /(@[a-zA-Z0-9._]+)/g,
    '<span class="mention-link" data-username="$1" style="color: #0095f6; cursor: pointer; font-weight: 600;">$1</span>'
  );

  return content;
};

const handleSummarize = async () => {
  // If we already have the summary, just toggle it ON
  if (summaryText.value) {
    isSummarized.value = true;
    return;
  }

  // Otherwise, fetch it from API
  try {
    isLoadingSummary.value = true;
    const response = await aiApi.summarizeText(props.post.caption);
    summaryText.value = response.data.summary;
    isSummarized.value = true;
  } catch (error) {
    console.error("Failed to summarize caption:", error);
    alert("Could not generate summary at this time.");
  } finally {
    isLoadingSummary.value = false;
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

const isOwnPost = computed(() => {
  return props.post.user_id === currentUserId; 
});

const toggleOriginal = () => {
  isSummarized.value = false;
};
</script>

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
          <div class="username-row">
            <p class="username">{{ post.username }}</p>
            <img 
              v-if="post.is_verified" 
              src="/icons/verified-icon.png" 
              alt="Verified" 
              class="verified-badge-small"
            />
          </div>
          <p class="timestamp">{{ formatTime(post.created_at) }}</p>
        </div>
      </div>

      <div class="more-options-wrapper">
        <button class="more-button" @click.stop="showMenu = !showMenu">
          <svg class="dots-icon" fill="currentColor" viewBox="0 0 24 24">
            <circle cx="6" cy="12" r="2" />
            <circle cx="12" cy="12" r="2" />
            <circle cx="18" cy="12" r="2" />
          </svg>
        </button>

        <div v-if="showMenu" v-click-outside="closeMenu" class="options-menu">
          <button 
            v-if="isOwnPost" 
            class="menu-item delete" 
            @click="handleDelete"
          >
            Remove Post
          </button>
          
          <button 
            v-else 
            class="menu-item report" 
            @click="openReportModal"
          >
            Report Post
          </button>
          
          <button class="menu-item cancel" @click="showMenu = false">
            Cancel
          </button>
        </div>
      </div>
    </div>

    <div class="post-media" @dblclick="$emit('toggle-like', post)">
      <div v-if="currentMedia" class="media-item-wrapper">
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
          class="action-button"
          :title="post.is_liked ? 'Unlike' : 'Like'"
          @click="$emit('toggle-like', post)"
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
        <button
          class="action-button"
          title="Share"
          @click="showShareModal = true"  >
          <img src="/icons/share-icon.png" alt="Share" class="action-icon" />
        </button>
      </div>

      <div 
        class="save-wrapper" 
        @mouseenter="handleMouseEnter" 
        @mouseleave="handleMouseLeave"
      >
        <button
          class="action-button"
          :title="isSaved ? 'Unsave' : 'Save'"
          @click="toggleDefaultSave"
        >
          <img
            :src="isSaved ? '/icons/saved-icon.png' : '/icons/save-icon.png'"
            alt="Save"
            class="action-icon"
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
                <span v-if="savedCollectionId === '' && isSaved" class="check-mark">✓</span>
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
                <span v-if="savedCollectionId === col.id && isSaved" class="check-mark">✓</span>
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
                <button class="mini-btn" @click="createNewCollection">Add</button>
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

    <div class="post-footer">
      <p class="likes-count">
        {{ post.likes_count ? post.likes_count.toLocaleString() : 0 }} likes
      </p>

      <div class="caption-section">
        <p class="caption-text">
          <span class="caption-username">{{ post.username }}</span>
          
          <span v-if="isSummarized" class="ai-summary-text">
            ✨ {{ summaryText }}
            <button class="toggle-btn" @click="toggleOriginal">
              (Show Original)
            </button>
          </span>

          <span v-else>
            <span
@click="handleCaptionClick"
            v-html="parseCaption(post.caption)"></span>
            
            <button 
              v-if="post.caption.length > 100" 
              class="summarize-btn" 
              :disabled="isLoadingSummary"
              @click="handleSummarize"
            >
              <span v-if="isLoadingSummary">✨ Summarizing...</span>
              <span v-else>✨ Summarize with AI</span>
            </button>
          </span>
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

    <div v-if="showReportModal" class="report-modal-backdrop" @click.self="showReportModal = false">
        <div class="report-modal">
            <h3>Report Post</h3>
            <textarea 
              v-model="reportReason" 
              placeholder="Why are you reporting this post?" 
              rows="4"
            ></textarea>
            <div class="modal-buttons">
                <button class="btn-submit" @click="submitReport">Submit</button>
                <button class="btn-cancel" @click="showReportModal = false">Cancel</button>
            </div>
        </div>
    </div>

    <ShareModal 
      v-if="showShareModal" 
      :content-id="post.id"
      type="post"
      :thumbnail="currentMedia?.media_url"
      @close="showShareModal = false"
    />
  </div>
</template>

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

.username-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

.verified-badge-small {
  width: 12px;       
  height: 12px;
  object-fit: contain;
  vertical-align: -1px; 
  margin-left: 4px;
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

.save-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.save-popover {
  position: absolute;
  bottom: 40px; /* Positions it above the icon */
  right: 0;
  width: 240px;
  background-color: #262626;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
  z-index: 50;
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

.summarize-btn {
  background: none;
  border: none;
  color: #a78bfa; /* Light purple for AI feel */
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  margin-left: 8px;
  padding: 0;
}

.summarize-btn:hover {
  text-decoration: underline;
}

.toggle-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 12px;
  cursor: pointer;
  margin-left: 5px;
}

.ai-summary-text {
  color: #e0e0e0;
  font-style: italic;
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

.more-options-wrapper {
  position: relative;
}

.options-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background-color: #262626;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.5);
  z-index: 20;
  width: 120px;
  overflow: hidden;
  border: 1px solid #363636;
}

.menu-item {
  padding: 10px 12px;
  font-size: 14px;
  cursor: pointer;
  text-align: center;
  border-bottom: 1px solid #363636;
  color: #fff;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-item:hover {
  background-color: #3a3a3a;
}

.menu-item.delete {
  color: #ed4956;
  font-weight: 600;
}

.more-options-wrapper {
  position: relative;
}

.options-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background-color: #262626;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.5);
  z-index: 20;
  width: 140px;
  overflow: hidden;
  border: 1px solid #363636;
  display: flex;
  flex-direction: column;
}

.menu-item {
  padding: 12px;
  font-size: 14px;
  cursor: pointer;
  background: none;
  border: none;
  border-bottom: 1px solid #363636;
  color: #fff;
  text-align: center;
  width: 100%;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-item:hover {
  background-color: #3a3a3a;
}

.menu-item.delete, .menu-item.report {
  color: #ed4956;
  font-weight: 600;
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
  width: 90%;
  max-width: 400px;
  color: white;
  text-align: center;
  border: 1px solid #363636;
}

.report-modal textarea {
  width: 100%;
  background: #121212;
  border: 1px solid #363636;
  color: white;
  border-radius: 4px;
  padding: 10px;
  resize: none;
  margin: 15px 0;
}

.modal-buttons {
  display: flex;
  gap: 10px;
  justify-content: center;
}

.btn-submit {
  background: #ed4956;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
}

.btn-cancel {
  background: transparent;
  color: #fff;
  border: 1px solid #363636;
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
}

/* Transition */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(5px);
}

:deep(.hashtag-link) {
  color: #00376b;
}
:deep(.hashtag-link:hover) {
  text-decoration: underline;
}
</style>
