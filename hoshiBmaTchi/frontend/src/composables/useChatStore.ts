import { ref, computed, shallowRef, markRaw } from "vue";
import type { Conversation, Message, User } from "../types/chat";
import AgoraRTC, {
  type IAgoraRTCClient,
  type ICameraVideoTrack,
  type IMicrophoneAudioTrack,
} from "agora-rtc-sdk-ng";
import { usersApi } from "../services/apiService";
import axios from "axios";

const currentUser = ref<User | null>(null);
const conversations = ref<Conversation[]>([]);
const messages = ref<Message[]>([]);
const selectedConversationId = ref<string | null>(null);
const isConnected = ref(false);
let socket: WebSocket | null = null;

const callState = ref<"idle" | "dialing" | "incoming" | "connected">("idle");
const activeCallType = ref<"audio" | "video">("video"); // Defined here
const incomingCaller = ref<{ id: string; name: string; avatar: string } | null>(
  null
);

const outgoingCallInfo = ref<{ name: string; avatar: string } | null>(null);

const agoraClient = shallowRef<IAgoraRTCClient | null>(null);
const localTracks = shallowRef<{
  video?: ICameraVideoTrack;
  audio?: IMicrophoneAudioTrack;
}>({});

const remoteUsers = ref<any[]>([]);
const isAudioEnabled = ref(true);
const isVideoEnabled = ref(true);

const API_URL = "/api";
const WS_URL = "ws://localhost:8081/ws";

export function useChatStore() {
  const getToken = () => localStorage.getItem("accessToken");

  const resolveSenderInfo = (senderId: string, conversationId: string) => {
    if (currentUser.value && senderId === currentUser.value.id) {
      return {
        name: currentUser.value.fullName || currentUser.value.username || "Me",
        avatar: currentUser.value.avatar || "/placeholder.svg",
      };
    }
    const conversation = conversations.value.find(
      (c) => c.id === conversationId
    );
    if (conversation) {
      const participant = conversation.participants.find(
        (p) => p.id === senderId
      );
      if (participant && participant.fullName) {
        return {
          name: participant.fullName,
          avatar: participant.avatar,
        };
      }
    }
    return { name: "Unknown", avatar: "/placeholder.svg" };
  };

  const initialize = async (user: User) => {
    currentUser.value = user;
    await fetchConversations();
    connectWebSocket();
  };

  const fetchConversations = async () => {
    try {
      const token = getToken();
      if (!token) return;

      const res = await fetch(`${API_URL}/chats`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (res.ok) {
        const data = await res.json();
        const mappedConversations = data.map((c: any) => ({
          ...c,
          isGroup: c.is_group,
          participants: c.participants || [],
          updatedAt:
            c.last_message?.created_at ||
            c.updated_at ||
            c.created_at ||
            new Date().toISOString(),
        }));

        const enrichedConversations = await Promise.all(
          mappedConversations.map(async (conv: any) => {
            const myId = currentUser.value?.id;
            conv.participants = await Promise.all(
              conv.participants.map(async (p: any) => {
                if (myId && p.id === myId && currentUser.value) {
                  return {
                    ...p,
                    fullName:
                      currentUser.value.fullName || currentUser.value.username,
                    username: currentUser.value.username,
                    avatar: currentUser.value.avatar || "/placeholder.svg",
                  };
                }
                try {
                  const userRes = await usersApi.getUserProfile(p.id);
                  const userData = userRes.data;
                  return {
                    ...p,
                    fullName: userData.name || userData.username,
                    username: userData.username,
                    avatar: userData.profile_picture_url || "/placeholder.svg",
                  };
                } catch (err) {
                  return {
                    ...p,
                    fullName: "Unknown User",
                    avatar: "/placeholder.svg",
                  };
                }
              })
            );

            if (!conv.isGroup) {
              const otherUser = conv.participants.find(
                (p: any) => p.id !== myId
              );
              if (otherUser) {
                conv.name = otherUser.fullName; 
                conv.avatar = otherUser.avatar;
              }
            }

            return conv;
          })
        );
        conversations.value = enrichedConversations;
      }
    } catch (error) {
      console.error("Failed to fetch chats:", error);
    }
  };

  const selectConversation = async (conversationId: string) => {
    selectedConversationId.value = conversationId;
    messages.value = [];

    try {
      const token = getToken();
      if (!token) return;

      const res = await fetch(
        `${API_URL}/chats/${conversationId}/messages?limit=50`,
        {
          headers: { Authorization: `Bearer ${token}` },
        }
      );

      if (res.ok) {
        const rawData = await res.json();
        const mappedMessages = rawData.map((msg: any) => {
          const { name, avatar } = resolveSenderInfo(
            msg.sender_id,
            conversationId
          );
          const isMedia = !!msg.media_url;
          return {
            id: msg.id,
            conversationId: msg.conversation_id,
            senderId: msg.sender_id,
            senderName: name,
            senderAvatar: avatar,
            content: isMedia ? msg.media_url : msg.content,
            mediaUrl: msg.media_url,
            messageType: isMedia ? "image" : "text",
            createdAt: msg.created_at,
            timestamp: msg.created_at,
            isUnsent: msg.is_unsent,
          };
        });
        messages.value = mappedMessages.reverse();
      }
    } catch (error) {
      console.error("Failed to fetch history:", error);
    }
  };

  const connectWebSocket = () => {
    if (socket) return;
    const token = getToken();
    if (!token) return;

    socket = new WebSocket(`${WS_URL}?token=${token}`);

    socket.onopen = () => {
      console.log("WS Connected");
      isConnected.value = true;
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);

        if (data.type === "signal") {
          handleSignal(data);
        } else if (data.type === "group_created") {
          fetchConversations();
        } else if (data.type === "new_message") {
          handleIncomingMessage(data);
        }
      } catch (e) {
        console.error("WS Parse Error", e);
      }
    };

    socket.onclose = () => {
      isConnected.value = false;
      socket = null;
      setTimeout(connectWebSocket, 3000);
    };
  };

  const handleIncomingMessage = async (wsMsg: any) => {
    let conversation = conversations.value.find(
      (c) => c.id === wsMsg.conversation_id
    );

    if (!conversation) {
      await fetchConversations();
      conversation = conversations.value.find(
        (c) => c.id === wsMsg.conversation_id
      );
    }

    const { name, avatar } = resolveSenderInfo(
      wsMsg.sender_id,
      wsMsg.conversation_id
    );
    const isMedia = !!wsMsg.media_url;

    const newMessage: Message = {
      id: wsMsg.id || `msg-${Date.now()}`,
      conversationId: wsMsg.conversation_id,
      senderId: wsMsg.sender_id,
      senderName: name,
      senderAvatar: avatar,
      content: isMedia ? wsMsg.media_url : wsMsg.content,
      messageType: isMedia ? "image" : "text",
      mediaUrl: wsMsg.media_url,
      createdAt: new Date().toISOString(),
      timestamp: new Date().toISOString(),
      status: "sent",
      isEdited: false,
      canUnsend: false,
    };

    if (selectedConversationId.value === wsMsg.conversation_id) {
      messages.value.push(newMessage);
    }

    if (conversation) {
      conversation.lastMessage = newMessage;
      if (selectedConversationId.value !== wsMsg.conversation_id) {
        conversation.unreadCount = (conversation.unreadCount || 0) + 1;
      }
      conversations.value = [
        conversation,
        ...conversations.value.filter((c) => c.id !== conversation!.id),
      ];
    }
  };

  const handleSignal = (data: any) => {
    if (data.sender_id === currentUser.value?.id) return;

    switch (data.signal_type) {
      case "incoming":
        if (callState.value === "idle") {
          const { name, avatar } = resolveSenderInfo(
            data.sender_id,
            data.conversation_id
          );
          incomingCaller.value = { id: data.sender_id, name, avatar };
          activeCallType.value = data.call_type;
          selectedConversationId.value = data.conversation_id;
          callState.value = "incoming";
        }
        break;

      case "end":
        if (
          callState.value !== "idle" &&
          selectedConversationId.value === data.conversation_id
        ) {
          leaveCall();
          alert("Call ended");
        }
        break;
    }
  };

  const initAgora = async (
    channel: string,
    token: string,
    appId: string,
    uid: number
  ) => {
    const client = AgoraRTC.createClient({ mode: "rtc", codec: "vp8" });

    client.on("user-published", async (user, mediaType) => {
      await client.subscribe(user, mediaType);

      if (mediaType === "video") {
        user.videoTrack?.play(`remote-player-${user.uid}`);
      }
      if (mediaType === "audio") {
        user.audioTrack?.play();
      }

      if (!remoteUsers.value.find((u) => u.uid === user.uid)) {
        remoteUsers.value.push(user);
      }
    });

    client.on("user-unpublished", (user) => {
      remoteUsers.value = remoteUsers.value.filter((u) => u.uid !== user.uid);
    });

    await client.join(appId, channel, token, uid);
    agoraClient.value = markRaw(client);

    if (activeCallType.value === "audio") {
      const audioTrack = await AgoraRTC.createMicrophoneAudioTrack();
      localTracks.value.audio = markRaw(audioTrack);
      await client.publish([localTracks.value.audio]);
    } else {
      const [mic, cam] = await AgoraRTC.createMicrophoneAndCameraTracks();
      localTracks.value.audio = markRaw(mic);
      localTracks.value.video = markRaw(cam);
      cam.play("local-player");
      await client.publish([mic, cam]);
    }
  };

  const startCall = async (type: "audio" | "video") => {
    if (!selectedConversationId.value || !currentUser.value) return;

    activeCallType.value = type;
    callState.value = "dialing";

    const conversation = conversations.value.find(
      (c) => c.id === selectedConversationId.value
    );
    if (conversation) {
      const other = conversation.participants.find(
        (p) => p.id !== currentUser.value!.id
      );

      if (other) {
        outgoingCallInfo.value = {
          name: other.fullName || other.username || "Unknown User",
          avatar: other.avatar || "/placeholder.svg",
        };
      } else {
        outgoingCallInfo.value = {
          name: conversation.name || "Unknown Group",
          avatar: conversation.avatar || "/placeholder.svg",
        };
      }
    }

    try {
      const token = getToken();
      const res = await axios.get(
        `${API_URL}/chats/${selectedConversationId.value}/call-token`,
        { headers: { Authorization: `Bearer ${token}` } }
      );

      const { token: agoraToken, app_id, channel_name } = res.data;

      const signalPayload = {
        type: "signal",
        signal_type: "incoming",
        call_type: type,
        conversation_id: selectedConversationId.value,
        sender_id: currentUser.value.id,
      };
      socket?.send(JSON.stringify(signalPayload));

      callState.value = "connected";
      const uid = Math.floor(Math.random() * 10000);
      await initAgora(channel_name, agoraToken, app_id, uid);
    } catch (e) {
      console.error("Call failed", e);
      callState.value = "idle";
      outgoingCallInfo.value = null; 
    }
  };

  const acceptCall = async () => {
    if (!selectedConversationId.value || !currentUser.value) return;

    try {
      const token = getToken();
      const res = await axios.get(
        `${API_URL}/chats/${selectedConversationId.value}/call-token`,
        { headers: { Authorization: `Bearer ${token}` } }
      );

      const { token: agoraToken, app_id, channel_name } = res.data;

      callState.value = "connected";
      const uid = Math.floor(Math.random() * 10000);
      await initAgora(channel_name, agoraToken, app_id, uid);
    } catch (e) {
      console.error("Accept failed", e);
      leaveCall();
    }
  };

  const leaveCall = async () => {
    localTracks.value.audio?.close();
    localTracks.value.video?.close();

    if (agoraClient.value) {
      await agoraClient.value.leave();
    }

    if (callState.value === "connected" && selectedConversationId.value) {
      socket?.send(
        JSON.stringify({
          type: "signal",
          signal_type: "end",
          conversation_id: selectedConversationId.value,
          sender_id: currentUser.value?.id,
        })
      );
    }

    callState.value = "idle";
    remoteUsers.value = [];
    incomingCaller.value = null;
    localTracks.value = {};
    agoraClient.value = null;
  };

  const toggleAudio = () => {
    if (localTracks.value.audio) {
      isAudioEnabled.value = !isAudioEnabled.value;
      localTracks.value.audio.setEnabled(isAudioEnabled.value);
    }
  };

  const toggleVideo = () => {
    if (localTracks.value.video) {
      isVideoEnabled.value = !isVideoEnabled.value;
      localTracks.value.video.setEnabled(isVideoEnabled.value);
    }
  };

  const sendMessage = (
    content: string,
    type: "text" | "gif" | "image" = "text",
    mediaUrl?: string
  ) => {
    if (!socket || !selectedConversationId.value || !currentUser.value) return;

    const payload = {
      type: "chat",
      conversation_id: selectedConversationId.value,
      sender_id: currentUser.value.id,
      content: content,
      media_type: type,
      media_url: mediaUrl,
    };

    socket.send(JSON.stringify(payload));
  };

  const deleteConversation = async (conversationId: string) => {
    const token = getToken();
    await fetch(`${API_URL}/chats/${conversationId}`, {
      method: "DELETE",
      headers: { Authorization: `Bearer ${token}` },
    });
    conversations.value = conversations.value.filter(
      (c) => c.id !== conversationId
    );
    if (selectedConversationId.value === conversationId) {
      selectedConversationId.value = null;
    }
  };

  const unsendMessage = async (messageId: string) => {
    const msg = messages.value.find((m) => m.id === messageId);
    if (msg) {
      msg.content = "This message was unsent";
      msg.isUnsent = true;
      msg.mediaUrl = undefined;
      msg.messageType = "text";
    }
    try {
      const token = getToken();
      await fetch(`${API_URL}/chats/messages/${messageId}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });
    } catch (err) {
      console.error("Failed to unsend:", err);
    }
  };

  const selectedConversation = computed(
    () =>
      conversations.value.find((c) => c.id === selectedConversationId.value) ||
      null
  );

  return {
    currentUser,
    conversations,
    messages,
    selectedConversationId,
    selectedConversation,
    isConnected,
    initialize,
    selectConversation,
    sendMessage,
    deleteConversation,
    unsendMessage,
    fetchConversations,

    callState,
    activeCallType,
    incomingCaller,
    outgoingCallInfo,
    remoteUsers,
    isAudioEnabled,
    isVideoEnabled,
    startCall,
    acceptCall,
    leaveCall,
    toggleAudio,
    toggleVideo,
  };
}
