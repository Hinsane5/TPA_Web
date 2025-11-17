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

