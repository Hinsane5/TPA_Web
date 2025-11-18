import axios from 'axios';

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
    return apiClient.get(`/v1/posts/user/${userId}`)
  },
  
  likePost: (postId: string) => {
    return apiClient.post(`/v1/posts/${postId}/like`);
  },

  unlikePost: (postId: string) => {
    return apiClient.delete(`/v1/posts/${postId}/like`);
  },

  createComment: (postId: string, content: string) => {
    return apiClient.post(`/v1/posts/${postId}/comments`, {content});
  },

  getCommentForPost: (postId: string) => {
    return apiClient.get(`/v1/posts/${postId}/comments`)
  }


}