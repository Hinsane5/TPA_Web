export interface User {
  id: string;
  username: string;
  fullName: string;
  avatar: string;
  isOnline: boolean;
}

export interface Message {
  id: string;
  senderId: string;
  senderName: string;
  senderAvatar: string;
  content: string;
  messageType: "text" | "image" | "gif" | "video";


  timestamp: Date | string;
  status: "sending" | "sent" | "seen";
  isEdited: boolean;
  canUnsend: boolean;


  createdAt: Date | string;


  conversationId?: string;
  isUnsent?: boolean;
  mediaUrl?: string;
}

export interface Conversation {
  id: string;
  name?: string;
  isGroup: boolean;
  participants: User[];
  lastMessage?: Message;
  unreadCount: number;
  updatedAt: string;
}

export interface GroupParticipant extends User {
  role: "admin" | "member";
}
