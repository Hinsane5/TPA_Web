import { ref, watch } from "vue";
import { useAuth } from "./useAuth";

export interface Notification {
  ID: number;
  sender_name: string;
  sender_image: string;
  type: "like" | "comment" | "follow" | "mention";
  message: string;
  entity_id: string;
  created_at: string;
  is_read: boolean;
}

const notifications = ref<Notification[]>([]);
const unreadCount = ref(0);
let socket: WebSocket | null = null;

export function useNotifications() {
  const { token, user } = useAuth(); 

  const connect = () => {
    if (socket || !token.value || !user.value?.id) return;

    const wsUrl = import.meta.env.VITE_WS_URL || "ws://localhost:8084/ws";
    socket = new WebSocket(
      `${wsUrl}?token=${token.value}&userId=${user.value.id}`
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
    };
  };

  const handleIncomingNotification = (notif: Notification) => {
    notifications.value.unshift(notif);
    unreadCount.value++;
  };

  const fetchHistory = async () => {
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
