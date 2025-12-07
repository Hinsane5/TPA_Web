import { ref, computed, watch, onUnmounted } from "vue";
import { storiesApi } from "../services/apiService";
import type { Story, User } from "../types/stories";

const stories = ref<Story[]>([]);
const currentStoryIndex = ref(0);
const selectedUsers = ref<Set<string>>(new Set());

export interface StoryGroup {
  userId: string;
  username: string;
  userAvatar: string;
  startIndex: number;
  hasUnseen: boolean;
}

export function useStories() {
  const isPlaying = ref(true);
  const progress = ref(0);
  const storyReplyText = ref("");
  let timer: number | null = null;
  const STORY_DURATION = 5000;
  const TICK_RATE = 100;

  const suggestedUsers = ref<User[]>([
    { id: "s1", username: "sher", fullName: "Sher", userAvatar: "" },
    { id: "s2", username: "perry", fullName: "Perry", userAvatar: "" },
  ]);

  const storyGroups = computed<StoryGroup[]>(() => {
    const groups: StoryGroup[] = [];
    const seenUsers = new Set<string>();

    stories.value.forEach((story, index) => {
      if (!seenUsers.has(story.userId)) {
        seenUsers.add(story.userId);
        
        const userStories = stories.value.filter(s => s.userId === story.userId);
        const hasUnseen = userStories.some(s => !s.isViewed);

        groups.push({
          userId: story.userId,
          username: story.user?.username || story.username,
          userAvatar: story.user?.userAvatar || story.userAvatar,
          startIndex: index, 
          hasUnseen
        });
      }
    });
    return groups;
  });
  
  

  const currentStory = computed(
    () => stories.value[currentStoryIndex.value] || null
  );

  const fetchStories = async () => {
    try {
      const response = await storiesApi.getFollowingStories();
      const rawStories = response.data.user_stories || [];

      // 1. Map raw data to your Story interface
      const mappedStories = rawStories.map((s: any) => {
        const userObj = s.user || {};
        return {
          id: s.id,
          mediaType: s.media_type ? s.media_type.toLowerCase() : "image",
          mediaUrl: s.media_url,
          isViewed: s.is_viewed || false,
          isLiked: s.is_liked || false,
          likes: s.likes_count || 0,
          timestamp: s.created_at ? new Date(s.created_at) : new Date(),
          replies: [],
          userId: s.user_id,
          username: userObj.username || "Unknown",
          userAvatar: userObj.userAvatar || "",
          isVerified: userObj.isVerified || false,
          user: {
            id: s.user_id,
            username: userObj.username || "Unknown",
            fullName: userObj.fullName || "",
            userAvatar: userObj.userAvatar || "",
          },
        };
      });

      // 2. CRITICAL FIX: Sort by User ID first, then by Date
      // This ensures [UserA_1, UserB_1, UserA_2] becomes [UserA_1, UserA_2, UserB_1]
      mappedStories.sort((a: Story, b: Story) => {
        // If users are different, group them together
        if (a.userId !== b.userId) {
          // Tip: If you have the current user's ID, check it here to put "Me" first
          return a.userId.localeCompare(b.userId);
        }
        // If same user, sort by time (oldest first)
        return a.timestamp.getTime() - b.timestamp.getTime();
      });

      stories.value = mappedStories;
    } catch (error) {
      console.error("Failed to fetch stories", error);
    }
  };

  const addReply = async () => {
    if (!storyReplyText.value.trim() || !currentStory.value) return;
    try {
      await storiesApi.replyToStory(
        currentStory.value.id,
        storyReplyText.value
      );
      storyReplyText.value = "";
      isPlaying.value = true;
    } catch (error) {
      console.error("Failed to reply", error);
    }
  };

  const toggleUserSelection = (userId: string) => {
    if (selectedUsers.value.has(userId)) {
      selectedUsers.value.delete(userId);
    } else {
      selectedUsers.value.add(userId);
    }
  };

  const sendStory = async () => {
    const recipients = Array.from(selectedUsers.value);
    if (recipients.length === 0 || !currentStory.value) return;
    try {
      console.log(`Sharing story ${currentStory.value.id} with`, recipients);
      selectedUsers.value.clear();
      return true;
    } catch (error) {
      console.error("Failed to share story", error);
    }
  };

  const startProgress = () => {
    stopProgress();
    if (!currentStory.value) return;
    if (currentStory.value.mediaType !== "video") {
      timer = window.setInterval(() => {
        if (!isPlaying.value) return;
        progress.value += (TICK_RATE / STORY_DURATION) * 100;
        if (progress.value >= 100) nextStory();
      }, TICK_RATE);
    }
  };

  const stopProgress = () => {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
    progress.value = 0;
  };

  const nextStory = () => {
    stopProgress();
    if (currentStoryIndex.value < stories.value.length - 1) {
      currentStoryIndex.value++;
    } else {
      isPlaying.value = false;
    }
  };

  const previousStory = () => {
    stopProgress();
    if (currentStoryIndex.value > 0) {
      currentStoryIndex.value--;
    }
  };

  const toggleLike = async () => {
    if (!currentStory.value) return;
    const story = currentStory.value;
    story.isLiked = !story.isLiked;
    story.likes += story.isLiked ? 1 : -1;

    try {
      if (story.isLiked) await storiesApi.likeStory(story.id);
      else await storiesApi.unlikeStory(story.id);
    } catch (error) {
      story.isLiked = !story.isLiked;
      story.likes += story.isLiked ? 1 : -1;
    }
  };

  watch(currentStory, (newStory) => {
    if (newStory) {
      startProgress();
      if (!newStory.isViewed) {
        storiesApi.viewStory(newStory.id).catch(console.error);
        newStory.isViewed = true;
      }
    }
  });

  onUnmounted(() => stopProgress());

  return {
    stories,
    storyGroups,
    currentStoryIndex,
    currentStory,
    progress,
    isPlaying,
    storyReplyText,
    suggestedUsers,
    selectedUsers,
    fetchStories,
    addReply,
    toggleUserSelection,
    sendStory,
    startProgress,
    stopProgress,
    nextStory,
    previousStory,
    toggleLike,
  };
}
