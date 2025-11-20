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
  timestamp: Date;
  status: "sending" | "sent" | "seen";
  isEdited: boolean;
  canUnsend: boolean;
  createdAt: Date;

  // Added: link back to conversation for local filtering
  conversationId?: string;
}

export interface Conversation {
  id: string;
  type: "direct" | "group";
  participants: User[];
  lastMessage: Message | null;
  unreadCount: number;
  updatedAt: Date;
  name?: string;
  avatar?: string;
}

export interface GroupParticipant extends User {
  role: "admin" | "member";
}
