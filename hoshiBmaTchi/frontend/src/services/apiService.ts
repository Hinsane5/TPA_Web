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
    // If the backend returns 401 Unauthorized
    if (error.response && error.response.status === 401) {
      console.log("Session expired. Logging out...");

      // 1. Clear local storage
      localStorage.removeItem("accessToken");
      localStorage.removeItem("userID");

      // 2. Redirect to Login page
      // Note: You might need to reload or push to router
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
  }) => {
    return apiClient.post("/v1/posts", data);
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
    // 1. Try to get ID from LocalStorage
    let userId = localStorage.getItem("userID");

    // 2. Fallback: Decode it from the Access Token if missing
    if (!userId) {
      const token = localStorage.getItem("accessToken");
      if (token) {
        try {
          const parts = token.split(".");
          if (parts.length >= 2) {
            const payloadPart = parts[1]; // FIX: Assign to variable first

            if (payloadPart) {
              // FIX: Explicit check to ensure it's a string
              const payload = JSON.parse(atob(payloadPart));
              // Check common claims for the ID
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
};

export const chatApi = {
  // Find or create a direct conversation with a target user
  getOrCreateConversation: (targetUserId: string) => {
    // Adjust the endpoint to match your backend route for creating/finding a chat
    return apiClient.post(`/chats`, {
      name: "Direct Message", // Optional, backend might ignore for DM
      user_ids: [targetUserId], // The person you want to message
      is_group: false, // Helpful flag if your backend supports it
    });
  },
};