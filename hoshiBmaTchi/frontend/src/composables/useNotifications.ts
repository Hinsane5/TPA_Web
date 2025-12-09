import { ref, watch } from "vue";
import { useAuth } from "./useAuth";

export interface Notification {
  ID: number;
  sender_name: string;
  sender_image: string;
  type: "like" | "comment" | "follow" | "mention";
  message: string;
  entity_id: number;
  created_at: string;
  is_read: boolean;
}

const notifications = ref<Notification[]>([]);
const unreadCount = ref(0);
let socket: WebSocket | null = null;

export function useNotifications() {
  const { token, user } = useAuth(); // Now this works because you fixed useAuth!

  const connect = () => {
    // Guard: Don't connect if no token or no user ID
    if (socket || !token.value || !user.value?.id) return;

    // Connect to Backend WS
    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8084/ws";
    socket = new WebSocket(
      `${wsUrl}?token=${token.value}&user_id=${user.value.id}`
    );

    socket.onopen = () => {
      console.log("Notification WS Connected");
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        handleIncomingNotification(data);
      } catch (e) {
        console.error("WS Parse Error", e);
      }
    };

    socket.onclose = () => {
      console.log("WS Disconnected");
      socket = null;
      // Optional: Add reconnection logic here
    };
  };

  const handleIncomingNotification = (notif: Notification) => {
    notifications.value.unshift(notif);
    unreadCount.value++;
  };

  const fetchHistory = async () => {
    // FIX: Guard clause to prevent crash if user is null
    if (!user.value || !user.value.id) return;

    try {
      const res = await fetch(
        `http://localhost:8084/notifications/${user.value.id}`,
        {
          headers: { Authorization: `Bearer ${token.value}` },
        }
      );
      if (res.ok) {
        const data = await res.json();
        notifications.value = data;
      }
    } catch (e) {
      console.error("Failed to fetch notifications", e);
    }
  };

  // Watch for user changes to connect/fetch automatically
  watch(user, (newUser) => {
    if (newUser) {
      connect();
      fetchHistory();
    }
  });

  return {
    notifications,
    unreadCount,
    connect,
    fetchHistory,
  };
}
