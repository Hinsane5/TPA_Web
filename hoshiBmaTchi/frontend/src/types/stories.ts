export interface Story {
  id: string;
  user: User;
  userId: string;
  username: string;
  userAvatar: string;
  isVerified: boolean;
  mediaUrl: string;
  mediaType: "image" | "video";
  timestamp: Date;
  viewedBy: string[];
  replies: StoryReply[];
  likes: number;
  isLiked: boolean;
  isViewed: boolean;
}

export interface StoryReply {
  id: string;
  userId: string;
  username: string;
  userAvatar: string;
  message: string;
  timestamp: Date;
}

export interface User {
  id: string;
  username: string;
  fullName: string;
  userAvatar: string;
  isVerified?: boolean;
}
