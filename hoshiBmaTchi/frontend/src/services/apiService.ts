import axios from 'axios';
import router from "@/router";

const apiClient = axios.create({
    baseURL: '/api',
    headers: {
        'Content-Type' : 'application/json',
    }
})

export const authApi = {
  register(data: any) {
    return apiClient.post("/auth/register", data);
  },

  sendOtp(email: string) {
    return apiClient.post("/auth/send-otp", { email });
  },

  login(data: any) {
    return apiClient.post("/auth/login", data);
  },

  loginWithGoogle(idToken: string) {
    return apiClient.post("/auth/google-login", { id_token: idToken });
  },

  verify2FA(data: any) {
    return apiClient.post("/auth/verify-2fa", data);
  },

  forgotPassword(email: string) {
    return apiClient.post("/auth/request-password-reset", { email });
  },

  resetPassword(data: any) {
    return apiClient.post("/auth/reset-password", data);
  },
};

export const setAuthHeader = (token: string) => {
  if (token) {
    apiClient.defaults.headers.common["Authorization"] = `Bearer ${token}`;
  } else{
    delete apiClient.defaults.headers.common["Authorization"];
  }
}

apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem("accessToken");
    if (token) {
      config.headers["Authorization"] = `Bearer ${token}`;
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
)

apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    if (error.response && error.response.status === 401) {
      console.log("Session expired. Logging out...");

      localStorage.removeItem("accessToken");
      localStorage.removeItem("userID");

      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);

export const postsApi = {
  generateUploadUrl: (fileName: string, fileType: string) => {
    return apiClient.get("/v1/posts/generate-upload-url", {
      params: {
        file_name: fileName,
        file_type: fileType,
      },
    });
  },

  uploadFileToMinio: (uploadUrl: string, file: File) => {
    return axios.put(uploadUrl, file, {
      headers: {
        "Content-Type": file.type,
      },
    });
  },

  createPost: (data: {
    media_object_name: string;
    media_type: string;
    caption: string;
    location: string;
    is_reel?: boolean;
  }) => {
    return apiClient.post("/v1/posts", data);
  },

  getPost: (postId: string) => {
    return apiClient.get(`/v1/posts/${postId}`);
  },

  getPostByUserID: (userId: string) => {
    return apiClient.get(`/v1/posts/user/${userId}`);
  },

  getHomeFeed: (limit: number, offset: number) => {
    return apiClient.get(`/v1/posts/feed`, {
      params: { limit, offset },
    });
  },

  likePost: (postId: string) => {
    return apiClient.post(`/v1/posts/${postId}/like`);
  },

  unlikePost: (postId: string) => {
    return apiClient.delete(`/v1/posts/${postId}/like`);
  },

  createComment: (postId: string, content: string) => {
    return apiClient.post(`/v1/posts/${postId}/comments`, { content });
  },

  getCommentForPost: (postId: string) => {
    return apiClient.get(`/v1/posts/${postId}/comments`);
  },

  toggleSavePost: (postId: string, collectionId: string = "") => {
    return apiClient.post(`/v1/posts/${postId}/save`, {
      collection_id: collectionId,
    });
  },

  getUserCollections: () => {
    return apiClient.get(`/v1/posts/collections`);
  },

  createCollection: (name: string) => {
    return apiClient.post(`/v1/posts/collections`, { name });
  },

  getUserMentions: (targetUserId: string) => {
    return apiClient.get(`/v1/posts/mentions/${targetUserId}`);
  },

  getExplorePosts: (limit: number, offset: number, hashtag: string = "") => {
    return apiClient.get(`/v1/posts/explore`, {
      params: { limit, offset, hashtag },
    });
  },

  getUserReels: (userId: string) => {
    return apiClient.get(`/v1/posts/user/${userId}/reels`);
  },

  getCollectionPosts: (
    collectionId: string,
    limit: number = 15,
    offset: number = 0
  ) => {
    return apiClient.get(`/v1/posts/collections/${collectionId}/posts`, {
      params: { limit, offset },
    });
  },

  updateCollection: (collectionId: string, name: string) => {
    return apiClient.put(`/v1/posts/collections/${collectionId}`, { name });
  },

  deleteCollection: (collectionId: string) => {
    return apiClient.delete(`/v1/posts/collections/${collectionId}`);
  },

  reportPost: (postId: string, reason: string) => {
    return apiClient.post(`/v1/posts/${postId}/report`, { reason });
  },
};

export const usersApi = {
  getUserProfile: (userId: string) => {
    return apiClient.get(`/v1/users/${userId}`);
  },

  followUser: (userId: string) => {
    return apiClient.post(`/v1/users/${userId}/follow`);
  },

  unfollowUser: (userId: string) => {
    return apiClient.delete(`/v1/users/${userId}/follow`);
  },

  getMe: () => {
    let userId = localStorage.getItem("userID");

    if (!userId) {
      const token = localStorage.getItem("accessToken");
      if (token) {
        try {
          const parts = token.split(".");
          if (parts.length >= 2) {
            const payloadPart = parts[1];

            if (payloadPart) {
              const payload = JSON.parse(atob(payloadPart));
              userId = payload.user_id || payload.sub || payload.id;
            }
          }
        } catch (e) {
          console.error("Failed to decode token for User ID", e);
        }
      }
    }

    if (!userId) return Promise.reject("No user ID found");
    return apiClient.get(`/v1/users/${userId}`);
  },

  searchUsers: (query: string) => {
    return apiClient.get(`/v1/users/search`, { params: { q: query } });
  },

  getSuggestedUsers: () => {
    return apiClient.get("/v1/users/suggested");
  },

  blockUser: (userId: string) => {
    return apiClient.post(`/v1/users/${userId}/block`);
  },

  unblockUser: (userId: string) => {
    return apiClient.delete(`/v1/users/${userId}/block`);
  },

  getBlockedUsers: () => {
    return apiClient.get(`/v1/users/blocked`);
  },

  reportUser: (userId: string, reason: string) => {
    return apiClient.post(`/v1/users/${userId}/report`, { reason });
  },
};

export const chatApi = {
  getOrCreateConversation: (targetUserId: string) => {
    return apiClient.post(`/chats`, {
      name: "Direct Message",
      user_ids: [targetUserId],
      is_group: false,
    });
  },

  shareContent: (data: {
    recipient_id: string;
    content_id: string;
    type: string;
    thumbnail?: string;
  }) => {
    return apiClient.post(`/chats/share`, data);
  },
  
};

export const storiesApi = {
  generateUploadUrl: (fileName: string, fileType: string) => {
    return apiClient.get("/v1/stories/upload-url", {
      params: {
        file_name: fileName,
        file_type: fileType,
      },
    });
  },

  createStory: (data: {
    media_object_name: string;
    media_type: string;
    duration: number;
  }) => {
    return apiClient.post("/stories", data);
  },

  getFollowingStories: () => {
    return apiClient.get("/stories/following");
  },

  getUserStories: (userId: string) => {
    return apiClient.get("/stories/user", {
      params: { user_id: userId },
    });
  },

  viewStory: (storyId: string) => {
    return apiClient.post("/stories/view", { story_id: storyId });
  },

  likeStory: (storyId: string) => {
    return apiClient.post("/stories/like", { story_id: storyId });
  },

  unlikeStory: (storyId: string) => {
    return apiClient.post("/stories/unlike", { story_id: storyId });
  },

  deleteStory: (storyId: string) => {
    return apiClient.delete("/stories", {
      params: { id: storyId },
    });
  },

  replyToStory: (storyId: string, content: string) => {
    return apiClient.post("/stories/reply", {
      story_id: storyId,
      content,
    });
  },

  getStoryReplies: (storyId: string) => {
    return apiClient.get("/stories/replies", {
      params: { story_id: storyId },
    });
  },

  shareStory: (storyId: string, recipientId: string) => {
    return apiClient.post("/stories/share", {
      story_id: storyId,
      recipient_id: recipientId,
    });
  },

  getStoryViewers: (storyId: string) => {
    return apiClient.get("/stories/viewers", {
      params: { story_id: storyId },
    });
  },
};

export const reelsApi = {
  getReelsFeed: (limit: number, offset: number) => {
    return apiClient.get(`/v1/reels/feed`, {
      params: { limit, offset },
    });
  },

  likeReel: (reelId: string) => {
    return apiClient.post(`/v1/reels/${reelId}/like`);
  },

  unlikeReel: (reelId: string) => {
    return apiClient.delete(`/v1/reels/${reelId}/like`);
  },

  saveReel: (reelId: string) => {
    return apiClient.post(`/v1/reels/${reelId}/save`);
  },

  unsaveReel: (reelId: string) => {
    return apiClient.delete(`/v1/reels/${reelId}/save`);
  },

  getShareRecipients: () => {
    return apiClient.get("/v1/users/following");
  },
};

export const markNotificationsRead = async (userId: string) => {
  const response = await fetch(
    `http://localhost:8084/notifications/${userId}/read`,
    {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
    }
  );

  if (!response.ok) {
    throw new Error("Failed to mark notifications as read");
  }
  return response.json();
};

export const settingsApi = {
  updateProfile: (data: {
    name: string;
    bio: string;
    gender: string;
    profile_picture_url?: string;
  }) => {
    return apiClient.put("/v1/settings/profile", data);
  },

  getSettings: () => {
    return apiClient.get("/v1/settings/preferences");
  },

  updateNotifications: (enablePush: boolean, enableEmail: boolean) => {
    return apiClient.put("/v1/settings/notifications", {
      enable_push: enablePush,
      enable_email: enableEmail,
    });
  },

  updatePrivacy: (isPrivate: boolean) => {
    return apiClient.put("/v1/settings/privacy", { is_private: isPrivate });
  },

  getCloseFriends: () => {
    return apiClient.get("/v1/settings/close-friends");
  },
  addCloseFriend: (targetUserId: string) => {
    return apiClient.post(`/v1/settings/close-friends/${targetUserId}`);
  },
  removeCloseFriend: (targetUserId: string) => {
    return apiClient.delete(`/v1/settings/close-friends/${targetUserId}`);
  },

  getBlockedUsers: () => {
    return apiClient.get("/v1/users/blocked");
  },
  unblockUser: (targetUserId: string) => {
    return apiClient.delete(`/v1/users/${targetUserId}/block`);
  },

  getHiddenStoryUsers: () => {
    return apiClient.get("/v1/settings/story-hide");
  },
  hideStoryFromUser: (targetUserId: string) => {
    return apiClient.post(`/v1/settings/story-hide/${targetUserId}`);
  },
  unhideStoryFromUser: (targetUserId: string) => {
    return apiClient.delete(`/v1/settings/story-hide/${targetUserId}`);
  },

  requestVerification: (data: {
    national_id: string;
    reason: string;
    selfie_url: string;
  }) => {
    return apiClient.post("/v1/settings/verification-request", data);
  },
};

export const adminApi = {
  getAllUsers: () => apiClient.get("/admin/users"),
  banUser: (userId: string, ban: boolean) =>
    apiClient.post(`/admin/users/${userId}/ban`, { is_banned: ban }),

  sendNewsletter: (subject: string, body: string) =>
    apiClient.post("/admin/newsletter", { subject, body }),

  getVerificationRequests: () => apiClient.get("/admin/verifications"),
  reviewVerification: (reqId: string, action: "ACCEPTED" | "REJECTED") =>
    apiClient.post(`/admin/verifications/${reqId}/review`, { action }),

  getReports: (type: "post" | "user") =>
    apiClient.get(`/admin/reports`, { params: { type } }),
  reviewReport: (reportId: string, type: "post" | "user", action: string) =>
    apiClient.post(`/admin/reports/${reportId}/review`, { type, action }),
};