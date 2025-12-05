export interface Story {
  id: string;
  userId: string;
  username: string;
  userAvatar: string;
  isVerified: boolean;
  imageUrl: string;
  timestamp: Date;
  viewedBy: string[];
  replies: StoryReply[];
  likes: number;
  isLiked: boolean;
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
