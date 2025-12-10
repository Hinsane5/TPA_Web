<template>
  <div class="profile-container">
    <div class="profile-header">
      <div class="profile-info">
        <div
          class="profile-picture-wrapper"
          @click="showProfileImageModal = true"
        >
          <img
            :src="
              profileUser?.profileImage ||
              '/placeholder.svg?height=150&width=150'
            "
            :alt="profileUser?.fullName || 'Profile'"
            class="profile-picture"
          />
        </div>

        <div class="profile-details">
          <div class="profile-top">
            <div class="user-info">
              <h1 class="full-name">
                {{ profileUser?.fullName || "Loading..." }}
              </h1>
              <p class="username">@{{ profileUser?.username || "username" }}</p>
            </div>

            <div class="profile-actions">
              <template v-if="isOwnProfile">
                <button class="action-btn">Edit profile</button>
                <button class="action-btn">View archive</button>
                <button class="action-btn settings-btn" title="Settings">
                  <img
                    src="/icons/setting-icon.png"
                    alt="Settings"
                    class="settings-icon"
                  />
                </button>
              </template>

              <template v-else>
                <button
                  class="action-btn follow-btn"
                  :class="{ following: isFollowing }"
                  @click="toggleFollow"
                >
                  {{ isFollowing ? "Following" : "Follow" }}
                </button>
                <button class="action-btn" @click="handleMessageClick">
                  Message
                </button>
              </template>
            </div>
          </div>

          <p class="bio">{{ profileUser?.bio || "No bio yet." }}</p>

          <div class="stats">
            <div class="stat">
              <span class="stat-number">{{
                profileUser?.postsCount || 0
              }}</span>
              <span class="stat-label">posts</span>
            </div>
            <div class="stat">
              <span class="stat-number">{{ profileUser?.followers || 0 }}</span>
              <span class="stat-label">followers</span>
            </div>
            <div class="stat">
              <span class="stat-number">{{ profileUser?.following || 0 }}</span>
              <span class="stat-label">following</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="profile-tabs">
      <button
        v-for="tab in tabs"
        :key="tab"
        :class="['tab', { active: activeTab === tab }]"
        @click="activeTab = tab"
      >
        <img :src="getTabIconPath(tab)" :alt="tab" class="tab-icon" />
        {{ tab.charAt(0).toUpperCase() + tab.slice(1) }}
      </button>
    </div>

    <div class="tab-content">
      <div v-if="activeTab === 'posts'" class="posts-grid">
        <div
          class="grid-item"
          v-for="post in posts"
          :key="post.id"
          @click="openPostDetail(post)"
        >
          <img
            :src="getThumbnail(post)"
            class="post-image"
            loading="lazy"
            alt="Post thumbnail"
          />

          <div class="grid-hover-overlay">
            <div class="hover-stat">
              <img
                src="/icons/notifications-icon.png"
                class="hover-icon"
                alt="Likes"
              />
              {{ post.likes_count || 0 }}
            </div>
            <div class="hover-stat">
              <img
                src="/icons/comment-icon.png"
                class="hover-icon"
                alt="Comments"
              />
              {{ post.comments_count || 0 }}
            </div>
          </div>
        </div>

        <div v-if="posts.length === 0" class="empty-state">
          <p>No posts yet.</p>
        </div>
      </div>

      <div v-if="activeTab === 'reels'">
        <ProfileReelsTab 
          :userId="profileUser.id" 
          @open-post="openPostDetail" 
        />
      </div>

      <div v-if="activeTab === 'saved'" class="saved-grid">
        <div v-if="isOwnProfile">
          <ProfileSavedTab />
        </div>
        <div v-else class="empty-state">
          <p>Saved posts are private.</p>
        </div>
      </div>

      <div v-if="activeTab === 'mentions'" class="mentions-grid">
        <div v-if="!hasContent" class="empty-state">
          <p>No mentions yet.</p>
        </div>
      </div>
    </div>

    <div
      v-if="showProfileImageModal"
      class="modal-overlay"
      @click="showProfileImageModal = false"
    >
      <div class="modal-content" @click.stop>
        <button class="close-btn" @click="showProfileImageModal = false">
          ✕
        </button>
        <img
          :src="
            profileUser?.profileImage || '/placeholder.svg?height=400&width=400'
          "
          :alt="profileUser?.fullName || 'Profile'"
          class="modal-image"
        />
      </div>
    </div>

    <PostDetailOverlay
      v-if="showPostOverlay"
      :is-open="showPostOverlay"
      :post="selectedPost"
      @close="closePostDetail"
      @toggle-like="handleLikeUpdate"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { useRoute } from "vue-router";
import { usersApi, postsApi, chatApi } from "../services/apiService";
import router from "@/router";
import PostDetailOverlay from "./PostDetailOverlay.vue"; 
import ProfileReelsTab from "./ProfileReelsTab.vue";
import ProfileSavedTab from "./ProfileSavedTab.vue";

const route = useRoute();
const posts = ref<any[]>([]);
const mentionsPosts = ref<any[]>([]);
const activeTab = ref("posts");
const showProfileImageModal = ref(false);
const hasContent = ref(false);
const tabs = ["posts", "reels", "saved", "mentions"] as const;
const isFollowing = ref(false);

const showPostOverlay = ref(false);
const selectedPost = ref<any>(null);

const profileUser = ref({
  id: "",
  fullName: "Loading...",
  username: "loading...",
  bio: "Loading...",
  postsCount: 0,
  followers: 0,
  following: 0,
  profileImage: "",
});

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
    return typeof id === "string" ? id : null;
  } catch (e) {
    return null;
  }
};

const currentUserId = getUserIdFromToken();

const getRouteId = (): string | undefined => {
  const param = route.params.id;
  return Array.isArray(param) ? param[0] : param;
};

const getDisplayUrl = (url: string) => {
  if (!url) return "/placeholder.png";
  return url.replace("http://minio:9000", "http://localhost:9000");
};

const getThumbnail = (post: any) => {
  if (post.media && Array.isArray(post.media) && post.media.length > 0) {
    return getDisplayUrl(post.media[0].media_url);
  }
  if (post.media_url) {
    return getDisplayUrl(post.media_url);
  }
  return "/placeholder.png";
};

const isOwnProfile = computed(() => {
  const paramId = getRouteId();
  if (!paramId) return true;
  return paramId === currentUserId;
});

const loadProfileData = async () => {
  const routeId = getRouteId();
  const rawId = routeId || currentUserId;

  if (!rawId || typeof rawId !== "string") {
    console.warn("Skipping profile load: No valid User ID found.");
    return;
  }

  const targetUserId: string = rawId;

  try {
    const userRes = await usersApi.getUserProfile(targetUserId);
    const data = userRes.data;

    profileUser.value = {
      id: data.id,
      fullName: data.name,
      username: data.username,
      bio: data.bio || "No bio yet.",
      profileImage: data.profile_picture_url,
      followers: data.followers_count,
      following: data.following_count,
      postsCount: 0,
    };

    if (data.is_following !== undefined) {
      isFollowing.value = data.is_following;
    }

    const postsRes = await postsApi.getPostByUserID(targetUserId);
    posts.value = postsRes.data || [];
    profileUser.value.postsCount = posts.value.length;
    hasContent.value = posts.value.length > 0;
  } catch (error) {
    console.error("Failed to load profile:", error);
  }
};

const handleMessageClick = async () => {
  const targetUserId = profileUser.value.id;
  if (!targetUserId) return;

  try {
    const res = await chatApi.getOrCreateConversation(targetUserId);
    const conversationId = res.data.conversation_id;

    router.push({
      name: "messages",
      query: { conversationId: conversationId },
    });
  } catch (error) {
    console.error("Failed to start conversation:", error);
    router.push({ name: "messages" });
  }
};

const toggleFollow = async () => {
  const targetId = profileUser.value.id;
  if (!targetId) return;

  try {
    if (isFollowing.value) {
      await usersApi.unfollowUser(targetId);
      isFollowing.value = false;
      profileUser.value.followers--;
    } else {
      await usersApi.followUser(targetId);
      isFollowing.value = true;
      profileUser.value.followers++;
    }
  } catch (error) {
    console.error("Follow action failed:", error);
  }
};

const getTabIconPath = (tab: string): string => {
  const icons: Record<string, string> = {
    posts: "/icons/post-icon.png",
    reels: "/icons/reels-icon.png",
    saved: "/icons/save-icon.png",
    mentions: "/icons/mention-icon.png",
  };
  return icons[tab] || "";
};

const openPostDetail = (post: any) => {
  selectedPost.value = post;
  showPostOverlay.value = true;
};

const closePostDetail = () => {
  showPostOverlay.value = false;
  selectedPost.value = null;
};

const handleLikeUpdate = (post: any) => {
  const target = posts.value.find((p) => p.id === post.id);
  if (target) {
    target.is_liked = !target.is_liked;
    target.likes_count += target.is_liked ? 1 : -1;
  }
};

const loadMentions = async () => {
  try {
    const res = await postsApi.getUserMentions(profileUser.value.id);
    mentionsPosts.value = res.data.posts || [];
    hasContent.value = mentionsPosts.value.length > 0;
  } catch (error) {
    console.error("Failed to load mentions", error);
  }
};

onMounted(() => {
  loadProfileData();
});

watch(
  () => route.params.id,
  () => {
    posts.value = [];
    loadProfileData();
  }
);

watch(activeTab, (newTab) => {
  if (newTab === 'mentions') {
    loadMentions();
  }
});
</script>

<style scoped>
.follow-btn {
  background-color: #0095f6;
  border: none;
  font-weight: 600;
}
.follow-btn:hover {
  background-color: #007bd2;
}
.follow-btn.following {
  background-color: transparent;
  border: 1px solid #404040;
  color: #fff;
}

.profile-container {
  width: 100%;
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
  color: #fff;
}

.profile-header {
  border-bottom: 1px solid #262626;
  padding-bottom: 40px;
  margin-bottom: 30px;
}

.profile-info {
  display: flex;
  gap: 40px;
  align-items: flex-start;
}

.profile-picture-wrapper {
  cursor: pointer;
  flex-shrink: 0;
}

.profile-picture {
  width: 150px;
  height: 150px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid #262626;
  transition: transform 0.2s ease;
}

.profile-picture-wrapper:hover .profile-picture {
  transform: scale(1.05);
}

.profile-details {
  flex: 1;
  min-width: 0;
}

.profile-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  gap: 20px;
}

.user-info h1 {
  font-size: 28px;
  font-weight: 300;
  margin: 0 0 5px 0;
}

.username {
  color: #a0a0a0;
  font-size: 16px;
  margin: 0;
}

.profile-actions {
  display: flex;
  gap: 10px;
}

.action-btn {
  padding: 8px 16px;
  background: #262626;
  border: 1px solid #404040;
  color: #fff;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: #383838;
}

.settings-btn {
  padding: 8px 12px;
  min-width: auto;
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
  filter: brightness(0.9);
}

.bio {
  color: #e0e0e0;
  margin: 0 0 20px 0;
  font-size: 14px;
  line-height: 1.4;
}

.stats {
  display: flex;
  gap: 40px;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-number {
  font-size: 20px;
  font-weight: 600;
}

.stat-label {
  color: #a0a0a0;
  font-size: 12px;
  margin-top: 5px;
}

.profile-tabs {
  display: flex;
  gap: 0;
  border-bottom: 1px solid #262626;
  margin-bottom: 20px;
  overflow-x: auto;
}

.tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 15px 20px;
  background: none;
  border: none;
  color: #a0a0a0;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  border-bottom: 2px solid transparent;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.tab:hover {
  color: #fff;
}

.tab.active {
  color: #fff;
  border-bottom-color: #fff;
}

.tab-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
  filter: brightness(0.9);
}

.tab.active .tab-icon {
  filter: brightness(1.2);
}

.tab-content {
  min-height: 400px;
}

.posts-grid,
.reels-grid,
.saved-grid,
.mentions-grid {
  display: grid;
  gap: 20px;
  width: 100%;
}

.posts-grid,
.mentions-grid {
  grid-template-columns: repeat(3, 1fr);
}

.reels-grid {
  grid-template-columns: repeat(4, 1fr);
}

.saved-grid {
  grid-template-columns: repeat(3, 1fr);
}

.grid-item {
  aspect-ratio: 1;
  background: #262626;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.grid-item:hover {
  transform: scale(1.02);
}

.post-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* Grid Hover Overlay Effect */
.grid-hover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  opacity: 0;
  transition: opacity 0.2s;
}

.grid-item:hover .grid-hover-overlay {
  opacity: 1;
}

.hover-stat {
  color: white;
  font-weight: bold;
  font-size: 16px;
  display: flex;
  align-items: center;
  gap: 5px;
}

.hover-icon {
  width: 20px;
  height: 20px;
  object-fit: contain;
  filter: invert(1); /* This turns the black icons into white */
  display: block;
}

/* Update .hover-stat to align items properly */
.hover-stat {
  color: white;
  font-weight: bold;
  font-size: 16px;
  display: flex;
  align-items: center;
  gap: 8px; /* Add some space between icon and number */
}

.empty-state {
  grid-column: 1 / -1;
  text-align: center;
  padding: 60px 20px;
  color: #a0a0a0;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  position: relative;
  max-width: 90%;
  max-height: 90vh;
}

.close-btn {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(0, 0, 0, 0.6);
  border: none;
  color: #fff;
  font-size: 24px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease;
  z-index: 10;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.8);
}

.modal-image {
  width: 100%;
  max-width: 500px;
  height: auto;
  border-radius: 8px;
  object-fit: contain;
}

/* Responsive Design */
@media (max-width: 768px) {
  .profile-info {
    flex-direction: column;
    align-items: center;
    gap: 20px;
    text-align: center;
  }

  .profile-picture {
    width: 120px;
    height: 120px;
  }

  .profile-top {
    flex-direction: column;
    align-items: center;
  }

  .profile-actions {
    width: 100%;
    justify-content: center;
  }

  .action-btn {
    flex: 1;
  }

  .stats {
    justify-content: center;
  }

  .posts-grid,
  .mentions-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .reels-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .profile-container {
    padding: 15px;
  }

  .profile-picture {
    width: 100px;
    height: 100px;
  }

  .user-info h1 {
    font-size: 22px;
  }

  .stats {
    gap: 20px;
  }

  .stat-number {
    font-size: 16px;
  }

  .posts-grid,
  .mentions-grid,
  .reels-grid,
  .saved-grid {
    grid-template-columns: 1fr;
  }
}
</style>
