import { defineStore } from "pinia";
import { ref, computed } from "vue";
import type { Notification } from "@/types";
import { markNotificationsRead } from "@/services/apiService";

export const useNotificationStore = defineStore("notification", () => {
  const notifications = ref<Notification[]>([]);
  const socket = ref<WebSocket | null>(null);
  const toastMessage = ref<Notification | null>(null);

  const unreadCount = computed(
    () => notifications.value.filter((n) => !n.is_read).length
  );

  const connectWebSocket = (userId: string) => {
    if (socket.value) return;

    const wsUrl = `ws://localhost:8084/ws?userId=${userId}`;
    console.log("Connecting to WS:", wsUrl);

    socket.value = new WebSocket(wsUrl);

    socket.value.onopen = () => {
      console.log("WebSocket Connected");
    };

    socket.value.onmessage = (event) => {
      try {
        const newNotification: Notification = JSON.parse(event.data);
        console.log("New Notification:", newNotification);

        notifications.value.unshift(newNotification);

        toastMessage.value = newNotification;
        setTimeout(() => {
          toastMessage.value = null;
        }, 5000);
      } catch (e) {
        console.error("Error parsing WS message:", e);
      }
    };

    socket.value.onclose = () => {
      console.log("WS Disconnected, retrying in 3s...");
      socket.value = null;
      setTimeout(() => connectWebSocket(userId), 3000);
    };
  };

  const fetchNotifications = async (userId: string) => {
    try {
      const res = await fetch(`http://localhost:8084/notifications/${userId}`);
      if (!res.ok) throw new Error("Failed to fetch");
      const data = await res.json();
      notifications.value = data;
    } catch (error) {
      console.error("Failed to fetch notifications history", error);
    }
  };

  const markNotificationsAsRead = async (userId: string) => {
    notifications.value.forEach((n) => {
      n.is_read = true;
    });

    try {
      await markNotificationsRead(userId);
      console.log("Backend updated: Notifications marked as read");
    } catch (error) {
      console.error("Failed to update backend:", error);
    }
  };

  const markAsRead = (notificationId: number) => {
    const notif = notifications.value.find((n) => n.ID === notificationId);
    if (notif) notif.is_read = true;
  };

  return {
    notifications,
    unreadCount,
    toastMessage,
    connectWebSocket,
    fetchNotifications,
    markNotificationsAsRead,
    markAsRead,
  };
});
