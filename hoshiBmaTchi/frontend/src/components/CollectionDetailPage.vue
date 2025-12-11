<template>
  <div class="collection-detail-page">
    <div class="collection-header">
      <button class="back-btn" @click="$router.go(-1)">❮</button>
      <div class="header-info">
        <h1 class="collection-title">{{ collectionName }}</h1>
      </div>
      <button class="menu-btn" @click="showActionsModal = true">•••</button>
    </div>

    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
    </div>
    
    <div v-else-if="posts.length === 0" class="empty-state">
      <p>No posts in this collection yet.</p>
    </div>

    <div v-else class="posts-grid">
      <div 
        v-for="post in posts" 
        :key="post.id" 
        class="grid-item"
        @click="openPostDetail(post)"
      >
        <div class="media-wrapper">
           <video 
             v-if="post.media[0]?.media_type.startsWith('video/')"
             :src="getDisplayUrl(post.media[0].media_url)"
             class="grid-media"
             muted
           />
           <img 
             v-else
             :src="getDisplayUrl(post.media[0]?.media_url)" 
             class="grid-media"
           />
        </div>
        
        <div class="hover-overlay">
          <div class="overlay-stat">
            <img src="/icons/liked-icon.png" class="overlay-icon" />
            <span>{{ post.likes_count }}</span>
          </div>
          <div class="overlay-stat">
            <img src="/icons/comment-icon.png" class="overlay-icon" />
            <span>{{ post.comments_count }}</span>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showActionsModal" class="modal-backdrop" @click.self="showActionsModal = false">
      <div class="modal-content">
        <div v-if="modalView === 'options'" class="options-list">
           <button class="modal-btn destructive" @click="modalView = 'delete_confirm'">
             Delete Collection
           </button>
           <button class="modal-btn" @click="startRename">
             Rename Collection
           </button>
           <button class="modal-btn cancel" @click="showActionsModal = false">
             Cancel
           </button>
        </div>

        <div v-if="modalView === 'rename'" class="rename-box">
          <h3>Rename Collection</h3>
          <input 
            v-model="renameInput" 
            placeholder="Collection Name" 
            class="rename-input"
            @keyup.enter="handleRename"
          />
          <div class="modal-actions">
             <button class="text-btn cancel" @click="modalView = 'options'">Cancel</button>
             <button class="text-btn save" @click="handleRename">Save</button>
          </div>
        </div>

        <div v-if="modalView === 'delete_confirm'" class="confirm-box">
           <h3>Delete Collection?</h3>
           <p class="confirm-text">This will delete the collection. The posts will remain in your saved list.</p> 
           <div class="options-list">
             <button class="modal-btn destructive" @click="handleDelete">Delete</button>
             <button class="modal-btn" @click="modalView = 'options'">Cancel</button>
           </div>
        </div>
      </div>
    </div>

    <PostDetailOverlay 
      v-if="selectedPost" 
      :isOpen="!!selectedPost" 
      :post="selectedPost"
      @close="selectedPost = null"
      @toggle-like="handleLikeToggle"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { postsApi } from '../services/apiService';
import PostDetailOverlay from './PostDetailOverlay.vue';

const route = useRoute();
const router = useRouter();
const collectionID = route.params.collectionID as string;

// State
const collectionName = ref("Collection"); // Ideally fetch name or pass as query param
const posts = ref<any[]>([]);
const loading = ref(true);
const showActionsModal = ref(false);
const modalView = ref<'options' | 'rename' | 'delete_confirm'>('options');
const renameInput = ref("");
const selectedPost = ref<any>(null);

// Fetch Data
const fetchPosts = async () => {
  try {
    loading.value = true;
    const res = await postsApi.getCollectionPosts(collectionID);
    posts.value = res.data.data || [];
    
    // Attempt to get name from query if passed, or could fetch specific collection details API if exists
    if (route.query.name) {
        collectionName.value = route.query.name as string;
    }
  } catch (err) {
    console.error("Failed to load collection posts", err);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
    fetchPosts();
});

// Rename Logic
const startRename = () => {
    renameInput.value = collectionName.value;
    modalView.value = 'rename';
};

const handleRename = async () => {
    if(!renameInput.value.trim()) return;
    try {
        await postsApi.updateCollection(collectionID, renameInput.value);
        collectionName.value = renameInput.value;
        showActionsModal.value = false;
        modalView.value = 'options';
    } catch(err) {
        alert("Failed to rename");
    }
};

// Delete Logic
const handleDelete = async () => {
    try {
        await postsApi.deleteCollection(collectionID);
        router.push({ name: 'profile', params: { id: localStorage.getItem('userID') } });
    } catch(err) {
        alert("Failed to delete");
    }
};

// Helpers
const getDisplayUrl = (url: string) => {
  if (!url) return "/placeholder.png";
  return url.replace("http://minio:9000", "http://localhost:9000");
};

const openPostDetail = (post: any) => {
    selectedPost.value = post;
};

const handleLikeToggle = (post: any) => {
    // Basic optimistic update for the grid item
    post.is_liked = !post.is_liked;
    post.likes_count += post.is_liked ? 1 : -1;
};
</script>

<style scoped>
.collection-detail-page {
    width: 100%;
    color: white;
    padding-bottom: 50px;
}

.collection-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 15px 20px;
    border-bottom: 1px solid #262626;
    margin-bottom: 15px;
}

.back-btn, .menu-btn {
    background: none;
    border: none;
    color: white;
    font-size: 24px;
    cursor: pointer;
    padding: 5px;
}

.collection-title {
    font-size: 16px;
    font-weight: 700;
    margin: 0;
    text-transform: uppercase;
    letter-spacing: 1px;
}

/* Grid Layout */
.posts-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 4px; /* Instagram-like tight spacing */
    width: 100%;
}

.grid-item {
    position: relative;
    aspect-ratio: 1 / 1;
    cursor: pointer;
    background: #121212;
}

.media-wrapper {
    width: 100%;
    height: 100%;
}

.grid-media {
    width: 100%;
    height: 100%;
    object-fit: cover;
}

/* Hover Overlay */
.hover-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.3);
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 20px;
    opacity: 0;
    transition: opacity 0.2s;
}

.grid-item:hover .hover-overlay {
    opacity: 1;
}

.overlay-stat {
    display: flex;
    align-items: center;
    gap: 5px;
    font-weight: bold;
    font-size: 16px;
}

.overlay-icon {
    width: 20px;
    height: 20px;
    filter: invert(1);
}

/* Modal Styles */
.modal-backdrop {
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0,0,0,0.65);
    z-index: 2000;
    display: flex;
    align-items: center;
    justify-content: center;
}

.modal-content {
    background: #262626;
    width: 400px;
    max-width: 90%;
    border-radius: 12px;
    overflow: hidden;
    text-align: center;
    animation: zoomIn 0.1s ease-out;
}

.options-list {
    display: flex;
    flex-direction: column;
}

.modal-btn {
    background: transparent;
    border: none;
    border-bottom: 1px solid #363636;
    color: white;
    padding: 14px;
    font-size: 14px;
    cursor: pointer;
    font-weight: 400;
}

.modal-btn:last-child {
    border-bottom: none;
}

.modal-btn.destructive {
    color: #ed4956;
    font-weight: 700;
}

.modal-btn.cancel {
    
}

/* Rename Box */
.rename-box, .confirm-box {
    padding: 20px;
}

.rename-box h3, .confirm-box h3 {
    margin: 0 0 15px 0;
    font-size: 18px;
    font-weight: 600;
}

.rename-input {
    width: 100%;
    padding: 10px;
    background: #121212;
    border: 1px solid #363636;
    color: white;
    border-radius: 6px;
    margin-bottom: 15px;
}

.modal-actions {
    display: flex;
    justify-content: flex-end;
    gap: 15px;
}

.text-btn {
    background: none;
    border: none;
    cursor: pointer;
    font-weight: 600;
}

.text-btn.cancel { color: #a0a0a0; }
.text-btn.save { color: #0095f6; }

.confirm-text {
    color: #a0a0a0;
    font-size: 14px;
    margin-bottom: 20px;
}

/* Loading */
.loading-state {
  display: flex;
  justify-content: center;
  padding: 50px;
}

.spinner {
  width: 30px;
  height: 30px;
  border: 3px solid #333;
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin { to { transform: rotate(360deg); } }
@keyframes zoomIn { from { transform: scale(1.1); opacity: 0; } to { transform: scale(1); opacity: 1; } }
</style>