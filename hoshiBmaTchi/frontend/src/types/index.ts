export type AuthPage =
  | "login"
  | "register"
  | "forgot-password"
  | "2fa-verification"

export type DashboardPage =
  | "home"
  | "explore"
  | "reels"
  | "messages"
  | "profile"
  | "dashboard"
  | "side-bar";

export interface User {
  id: string;
  username: string;
  fullName: string;
  email: string;
  bio: string;
  profileImage: string;
  followers: number;
  following: number;
  postsCount: number;
  verified: boolean;
}

export interface Notification {
  ID: number;
  CreatedAt: string;
  UpdatedAt: string;
  recipient_id: string;
  sender_id: string;
  sender_name: string;
  sender_image: string;
  type: "like" | "comment" | "follow" | "mention";
  entity_id: string;
  message: string;
  is_read: boolean;
}

