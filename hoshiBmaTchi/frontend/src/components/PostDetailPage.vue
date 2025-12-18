<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { postsApi, usersApi } from '@/services/apiService';
import { format } from 'date-fns';
import CommentItem from './commentItem.vue';
import ShareModal from './ShareModal.vue';

const route = useRoute();
const router = useRouter();

const post = ref<any>(null);
const comments = ref<any[]>([]);
const morePosts = ref<any[]>([]);
const loading = ref(true);
const currentIndex = ref(0);
const newComment = ref("");
const commentInput = ref<HTMLInputElement | null>(null);
const showShareModal = ref(false);

const loadData = async () => {
  const postId = route.params.id as string;
  loading.value = true;
  currentIndex.value = 0;
  
  try {
    // 1. Fetch Post Details
    const res = await postsApi.getPost(postId);
    post.value = res.data;

    // 2. Fetch Comments
    const commentsRes = await postsApi.getCommentForPost(postId);
    const rawComments = commentsRes.data || [];
    
    // Enrich comments with user data (simplified for brevity, ideally cache users)
    const enriched = await Promise.all(rawComments.map(async (c: any) => {
      try {
        const u = await usersApi.getUserProfile(c.user_id);
        return {
          ...c,
          username: u.data.username,
          profile_picture: u.data.profile_picture_url
        };
      } catch {
        return { ...c, username: 'Unknown', profile_picture: '' };
      }
    }));
    comments.value = enriched;

    // 3. Fetch "More Posts" from user
    if (post.value && post.value.user_id) {
      const userPostsRes = await postsApi.getPostByUserID(post.value.user_id);
      // Filter out current post and take top 6
      morePosts.value = (userPostsRes.data || [])
        .filter((p: any) => p.id !== postId)
        .slice(0, 6);
    }

  } catch (error) {
    console.error("Failed to load post page data", error);
  } finally {
    loading.value = false;
  }
};

onMounted(loadData);

// Reload when route ID changes (e.g. clicking a post in "More Posts")
watch(() => route.params.id, loadData);

const mediaList = computed(() => {
  if (!post.value) return [];
  if (post.value.media && post.value.media.length > 0) return post.value.media;
  if (post.value.media_url) return [{ media_url: post.value.media_url, media_type: post.value.media_type || 'image/jpeg' }];
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

const getThumbnail = (p: any) => {
  if (p.media && p.media.length > 0) return getDisplayUrl(p.media[0].media_url);
  return p.media_url ? getDisplayUrl(p.media_url) : '/placeholder.png';
};

const handleLike = async () => {
  if (!post.value) return;
  try {
    if (post.value.is_liked) {
      await postsApi.unlikePost(post.value.id);
      post.value.is_liked = false;
      post.value.likes_count--;
    } else {
      await postsApi.likePost(post.value.id);
      post.value.is_liked = true;
      post.value.likes_count++;
    }
  } catch (e) {
    console.error(e);
  }
};

const toggleSave = async () => {
   if (!post.value) return;
   try {
     await postsApi.toggleSavePost(post.value.id);
     post.value.is_saved = !post.value.is_saved;
   } catch (e) { console.error(e); }
};

const submitComment = async () => {
  if (!newComment.value.trim() || !post.value) return;
  const text = newComment.value;
  newComment.value = "";
  
  try {
    await postsApi.createComment(post.value.id, text);
    // Optimistic update or reload comments
    const me = await usersApi.getMe();
    comments.value.push({
      id: Date.now().toString(), // temp id
      content: text,
      created_at: new Date().toISOString(),
      username: me.data.username,
      profile_picture: me.data.profile_picture_url,
      user_id: me.data.id
    });
  } catch (e) {
    console.error(e);
    alert("Failed to comment");
  }
};

const focusInput = () => commentInput.value?.focus();
const handleReply = (username: string) => {
  newComment.value = `@${username} `;
  focusInput();
};
const goToProfile = (userId: string) => router.push({ name: 'profile', params: { id: userId }});
const goToPost = (postId: string) => router.push({ name: 'post-detail', params: { id: postId }});
const formatFullDate = (d: string) => {
  try { return format(new Date(d), "MMMM d, yyyy"); } catch { return ""; }
};
</script>

<template>
  <div class="page-container">
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
    </div>

    <div v-else-if="post" class="content-wrapper">
      <div class="post-main-container">
        <div class="media-section">
          <div class="media-wrapper" @dblclick="handleLike">
            <template v-if="currentMedia">
              <video
                v-if="currentMedia.media_type && currentMedia.media_type.startsWith('video/')"
                :src="getDisplayUrl(currentMedia.media_url)"
                controls
                autoplay
                class="post-content-media"
              ></video>
              <img
                v-else
                :src="getDisplayUrl(currentMedia.media_url)"
                class="post-content-media"
                alt="Post content"
              />
            </template>
          </div>

          <button
            v-if="hasMultiple && currentIndex > 0"
            class="nav-btn left"
            @click="currentIndex--"
          >
            ❮
          </button>
          <button
            v-if="hasMultiple && currentIndex < mediaList.length - 1"
            class="nav-btn right"
            @click="currentIndex++"
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

        <div class="details-section">
          <div class="header">
            <div class="user-info" @click="goToProfile(post.user_id)">
              <img :src="post.profile_picture || '/placeholder.png'" class="avatar" />
              <span class="username">{{ post.username }}</span>
            </div>
            </div>

          <div class="comments-list">
            <CommentItem
              v-if="post.caption"
              :username="post.username"
              :profile-image="post.profile_picture"
              :comment-text="post.caption"
              :timestamp="post.created_at"
              :likes="post.likes_count"
            />
            
            <div v-if="post.caption" class="divider"></div>

            <CommentItem
              v-for="comment in comments"
              :key="comment.id"
              :username="comment.username"
              :profile-image="comment.profile_picture"
              :comment-text="comment.content"
              :timestamp="comment.created_at"
              @reply="handleReply(comment.username)"
            />
            
            <p v-if="comments.length === 0 && !post.caption" class="no-comments">
              No comments yet.
            </p>
          </div>

          <div class="actions-container">
            <div class="icons-row">
              <div class="left-icons">
                <button class="icon-btn" @click="handleLike">
                  <img
                    :src="post.is_liked ? '/icons/liked-icon.png' : '/icons/notifications-icon.png'"
                    class="icon"
                    :class="{ active: post.is_liked }"
                  />
                </button>
                <button class="icon-btn" @click="focusInput">
                  <img src="/icons/comment-icon.png" class="icon" />
                </button>
                <button class="icon-btn" @click="showShareModal = true">
                  <img src="/icons/share-icon.png" class="icon" />
                </button>
              </div>
              <button class="icon-btn" @click="toggleSave">
                 <img
                    :src="post.is_saved ? '/icons/saved-icon.png' : '/icons/save-icon.png'"
                    class="icon"
                  />
              </button>
            </div>
            <p class="likes-count">{{ post.likes_count }} likes</p>
            <p class="timestamp">{{ formatFullDate(post.created_at) }}</p>
          </div>

          <div class="input-section">
            <input 
              ref="commentInput"
              v-model="newComment"
              placeholder="Add a comment..."
              @keyup.enter="submitComment"
            />
            <button v-if="newComment.trim()" @click="submitComment">Post</button>
          </div>
        </div>
      </div>

      <div class="more-posts-section">
        <h3>More posts from <span class="highlight" @click="goToProfile(post.user_id)">{{ post.username }}</span></h3>
        <div class="posts-grid">
          <div 
            v-for="p in morePosts" 
            :key="p.id" 
            class="grid-item"
            @click="goToPost(p.id)"
          >
            <img :src="getThumbnail(p)" class="grid-img" />
            <div class="hover-overlay">
              <div class="stat">
                 <img src="/icons/liked-icon.png" class="small-icon" /> {{ p.likes_count }}
              </div>
              <div class="stat">
                 <img src="/icons/comment-icon.png" class="small-icon" /> {{ p.comments_count }}
              </div>
            </div>
          </div>
        </div>
        <div v-if="morePosts.length === 0" class="no-more-posts">
          No other posts to show.
        </div>
      </div>
    </div>

    <ShareModal 
      v-if="showShareModal && post"
      :content-id="post.id"
      type="post"
      :thumbnail="currentMedia ? getDisplayUrl(currentMedia.media_url) : ''"
      @close="showShareModal = false"
    />
  </div>
</template>

<style scoped>
.page-container {
  max-width: 935px;
  margin: 0 auto;
  padding: 30px 20px;
  width: 100%;
  color: white;
}

.loading-state {
  display: flex;
  justify-content: center;
  padding: 50px;
}
.spinner {
  width: 30px;
  height: 30px;
  border: 3px solid #262626;
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 1s infinite linear;
}

.post-main-container {
  display: flex;
  border: 1px solid #262626;
  background: #000;
  border-radius: 4px;
  /* Fixed height for the detail view style, optional but matches overlay feel */
  height: 600px; 
  margin-bottom: 50px;
}

/* Left Media */
.media-section {
  flex: 1.5;
  background: #000;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  border-right: 1px solid #262626;
}
.media-wrapper {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
.post-content-media {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.nav-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(26,26,26,0.8);
  color: #fff;
  border: none;
  width: 30px; height: 30px;
  border-radius: 50%;
  cursor: pointer;
  z-index: 2;
}
.left { left: 10px; }
.right { right: 10px; }

/* Right Details */
.details-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 350px;
}
.header {
  padding: 15px;
  border-bottom: 1px solid #262626;
  display: flex;
  align-items: center;
}
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
}
.avatar {
  width: 32px; height: 32px;
  border-radius: 50%;
}
.username {
  font-weight: 600;
  font-size: 14px;
}

.comments-list {
  flex: 1;
  overflow-y: auto;
  padding: 15px;
}
.no-comments {
  color: #8e8e8e;
  text-align: center;
  margin-top: 20px;
}
.divider {
  border-bottom: 1px solid #262626;
  margin: 10px 0;
}

.actions-container {
  padding: 15px;
  border-top: 1px solid #262626;
}
.icons-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}
.left-icons { display: flex; gap: 15px; }
.icon-btn { background: none; border: none; cursor: pointer; padding: 0; }
.icon { width: 24px; height: 24px; }
.likes-count { font-weight: 600; font-size: 14px; margin-bottom: 5px; }
.timestamp { font-size: 12px; color: #8e8e8e; }

.input-section {
  padding: 15px;
  border-top: 1px solid #262626;
  display: flex;
}
.input-section input {
  flex: 1;
  background: none;
  border: none;
  color: white;
  outline: none;
}
.input-section button {
  background: none;
  border: none;
  color: #0095f6;
  font-weight: 600;
  cursor: pointer;
}

/* More Posts Section */
.more-posts-section {
  padding-top: 20px;
  border-top: 1px solid #262626;
}
.more-posts-section h3 {
  color: #8e8e8e;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 20px;
}
.highlight {
  color: #fff;
  cursor: pointer;
}
.posts-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}
.grid-item {
  aspect-ratio: 1;
  position: relative;
  cursor: pointer;
  background: #262626;
}
.grid-img {
  width: 100%; height: 100%;
  object-fit: cover;
}
.hover-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0,0,0,0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 15px;
  opacity: 0;
  transition: opacity 0.2s;
}
.grid-item:hover .hover-overlay { opacity: 1; }
.stat { font-weight: bold; display: flex; align-items: center; gap: 5px; }
.small-icon { width: 16px; height: 16px; filter: invert(1); }

@media (max-width: 768px) {
  .post-main-container {
    flex-direction: column;
    height: auto;
  }
  .media-section { border-right: none; height: 400px; }
  .details-section { min-width: 100%; border-top: 1px solid #262626; }
  .posts-grid { gap: 3px; }
}

@keyframes spin { 100% { transform: rotate(360deg); } }
</style>