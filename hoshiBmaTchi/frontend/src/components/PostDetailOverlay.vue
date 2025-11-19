<template>
  <div v-if="isOpen" class="overlay-backdrop" @click="$emit('close')">
    <div class="overlay-container" @click.stop>
      
      <div class="overlay-left">
        <div class="media-container">
          <video 
            v-if="post.media_type?.startsWith('video')" 
            :src="post.media_url" 
            class="post-media" 
            controls autoplay muted loop
          ></video>
          <img v-else :src="post.media_url" class="post-media" />
        </div>
      </div>

      <div class="overlay-right">
        
        <div class="overlay-header">
          <div class="header-content">
            <img :src="post.profile_picture || '/default-avatar.png'" class="author-avatar" />
            <div class="author-info">
              <p class="author-username">{{ post.username }}</p>
              <p v-if="!isOwnPost && !isFollowing" class="follow-text" @click="handleFollow">
                 • Follow
              </p>
              <p v-if="!isOwnPost && isFollowing" class="following-text" @click="handleUnfollow">
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
          
          <div v-if="!loadingComments && enrichedComments.length === 0 && !post.caption" class="no-comments">
            <p>No comments yet.</p>
          </div>
        </div>

        <div class="divider"></div>

        <div class="action-section-wrapper">
           <div class="action-icons">
             <div class="icons-left">
               <button class="icon-button" @click="$emit('toggle-like', post)">
                  <img :src="post.is_liked ? '/icons/liked-icon.png' : '/icons/notifications-icon.png'" class="icon"  :class="{ active: post.is_liked }"/>
               </button>
               <button class="icon-button" @click="focusInput">
                  <img src="/icons/comment-icon.png" class="icon" />
               </button>
               <button class="icon-button">
                  <img src="/icons/share-icon.png" class="icon" />
               </button>
             </div>
             <button class="icon-button" @click="toggleSave">
                <img :src="isSaved ? '/icons/saved-icon.png' : '/icons/save-icon.png'" class="icon" />
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
import { ref, computed, onMounted } from 'vue';
import { format } from 'date-fns';
import { postsApi, usersApi } from '@/services/apiService';
import CommentItem from './commentItem.vue';

const props = defineProps(['isOpen', 'post']);
const emit = defineEmits(['close', 'comment-added', 'toggle-like']);

// State
const rawComments = ref<any[]>([]);
const userCache = ref(new Map()); // Cache for user profiles [userId -> UserData]
const commentText = ref('');
const loadingComments = ref(true);
const isSubmitting = ref(false);
const commentInputRef = ref<HTMLInputElement | null>(null);

const isSaved = ref(false);
const isFollowing = ref(false);

// User Info (from localStorage)
const currentUserId = localStorage.getItem('userID');
const currentUsername = localStorage.getItem('username') || 'me';
const currentUserPic = localStorage.getItem('profilePicture') || '';
const isOwnPost = props.post.user_id === currentUserId;

// --- Computed: Merge Comments with User Data ---
const enrichedComments = computed(() => {
  return rawComments.value.map(comment => {
    const user = userCache.value.get(comment.user_id);
    return {
      ...comment,
      username: user ? user.username : 'Loading...', // Show Loading or ID
      profile_picture: user ? user.profile_picture_url : null
    };
  });
});

// --- Fetch Logic ---
onMounted(async () => {
  document.body.style.overflow = 'hidden';
  
  // 1. Check Follow Status
  if (!isOwnPost) {
    try {
      const profileRes = await usersApi.getUserProfile(props.post.user_id);
      isFollowing.value = profileRes.data.is_following;
    } catch (e) { console.error("Follow check failed", e); }
  }

  // 2. Fetch Comments
  try {
    const res = await postsApi.getCommentForPost(props.post.id);
    rawComments.value = res.data || [];
    
    // 3. Extract unique User IDs from comments
    const userIdsToFetch = new Set(rawComments.value.map(c => c.user_id));
    
    // 4. Fetch Profiles (Parallel)
    const fetchPromises = Array.from(userIdsToFetch).map(async (uid: string) => {
      if (userCache.value.has(uid)) return; // Skip if cached
      try {
        const userRes = await usersApi.getUserProfile(uid);
        userCache.value.set(uid, userRes.data);
      } catch (e) {
        // Fallback if user not found
        userCache.value.set(uid, { username: 'Unknown User', profile_picture_url: '' });
      }
    });

    await Promise.all(fetchPromises);
    // Force reactivity update for Map
    userCache.value = new Map(userCache.value);

  } catch (e) {
    console.error("Fetch comments failed", e);
  } finally {
    loadingComments.value = false;
  }
});

// --- Actions ---
const submitComment = async () => {
  if (!commentText.value.trim()) return;
  
  const tempText = commentText.value;
  commentText.value = '';
  isSubmitting.value = true;

  // Optimistic Update
  const fakeId = Date.now().toString();
  const optimisticComment = {
    id: fakeId,
    user_id: currentUserId, // Our ID
    content: tempText,
    created_at: new Date().toISOString()
  };

  // Pre-fill cache with our own data so our comment looks correct instantly
  if (currentUserId && !userCache.value.has(currentUserId)) {
    userCache.value.set(currentUserId, { 
      username: currentUsername, 
      profile_picture_url: currentUserPic 
    });
  }
  
  rawComments.value.push(optimisticComment);

  try {
    await postsApi.createComment(props.post.id, tempText);
    emit('comment-added');
    // In production, you would replace fakeId with real ID from response
  } catch (e) {
    rawComments.value = rawComments.value.filter(c => c.id !== fakeId);
    commentText.value = tempText;
    alert("Failed to post comment");
  } finally {
    isSubmitting.value = false;
    // Scroll to bottom logic can be added here
  }
};

const handleReply = (username: string) => {
  commentText.value = `@${username} `;
  focusInput();
};

const focusInput = () => {
  commentInputRef.value?.focus();
};

const handleFollow = async () => {
  isFollowing.value = true;
  try { await usersApi.followUser(props.post.user_id); } 
  catch { isFollowing.value = false; }
};

const handleUnfollow = async () => {
  isFollowing.value = false;
  try { await usersApi.unfollowUser(props.post.user_id); } 
  catch { isFollowing.value = true; }
};

const toggleSave = () => isSaved.value = !isSaved.value;

const formatFullDate = (d: string | number | Date) => {
  try { return format(new Date(d), 'MMMM d, yyyy'); } catch { return ''; }
};
</script>

<style scoped>
.overlay-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.overlay-container {
  display: flex;
  width: 100%;
  max-width: 1000px;
  height: 600px;
  background-color: #000;
  --bg-primary: #000000;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 32px rgba(0, 0, 0, 0.3);
}

/* Left Side - Image */
.overlay-left {
  flex: 1;
  background-color: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.post-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* Right Side - Details */
.overlay-right {
  width: 360px;
  display: flex;
  flex-direction: column;
  background-color: #202327;
  border-left: 1px solid #e0e0e0;
}

/* Header */
.overlay-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.author-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  object-fit: cover;
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.author-username {
  font-size: 14px;
  font-weight: 600;
  color: #000000;
  margin: 0;
}

.follow-status {
  font-size: 12px;
  color: #65676b;
  margin: 0;
}

.close-button {
  background: none;
  border: none;
  font-size: 24px;
  color: #fff;
  cursor: pointer;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background-color 0.2s ease;
}

.close-button:hover {
  background-color: #f0f0f0;
}

/* Divider */
.divider {
  height: 1px;
  /* background-color: #e0e0e0; */
}

/* Comments Section */
.comments-section {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 12px 0;
}

.comments-section::-webkit-scrollbar {
  width: 8px;
}

.comments-section::-webkit-scrollbar-track {
  background: transparent;
}

.comments-section::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 4px;
}

/* Stats Section */
.stats-section {
  padding: 12px 16px;
}

.likes-text {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin: 0 0 4px 10px;
}

.date-text {
  font-size: 12px;
  color: #65676b;
  margin-left: 9px;
}

/* Action Icons */
.action-icons {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
}

.icons-left {
  display: flex;
  gap: 12px;
}

.icon-button {
  background: none;
  border: none;
  padding: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: background-color 0.2s ease;
}

.icon-button:hover {
  background-color: #f0f0f0;
}

.icon {
  width: 24px;
  height: 24px;
  opacity: 0.6;
  transition: opacity 0.2s ease;
}

.icon.active {
  opacity: 1;
}

/* Comment Input Section */
.comment-input-section {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  align-items: flex-start;
}

.input-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  flex-shrink: 0;
  margin-top: 4px;
}

.input-wrapper {
  flex: 1;
  display: flex;
  gap: 8px;
  align-items: center;
}

.comment-input {
  flex: 1;
  background: #202327;
  /* border: 1px solid #e0e0e0; */
  border: none;
  /* border-radius: 20px; */
  padding: 8px 16px;
  font-size: 14px;
  color: #fff;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s ease;
}

.comment-input:focus {
  border-color: #cccccc;
}

.comment-input::placeholder {
  color: #65676b;
}

.post-button {
  background: none;
  border: none;
  color: #5b5bff;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: color 0.2s ease;
  white-space: nowrap;
}

.post-button:hover {
  color: #4949ff;
}

.post-button:disabled {
  color: #cccccc;
  cursor: not-allowed;
}

/* Responsive */
@media (max-width: 768px) {
  .overlay-container {
    flex-direction: column;
    height: 100vh;
    max-width: 100%;
    border-radius: 0;
  }

  .overlay-left {
    height: 300px;
  }

  .overlay-right {
    width: 100%;
    border-left: none;
    border-top: 1px solid #e0e0e0;
  }
}
</style>
