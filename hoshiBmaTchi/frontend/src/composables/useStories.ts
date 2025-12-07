import { ref, computed, watch, onUnmounted } from "vue";
import { storiesApi } from "../services/apiService";
import type { Story, StoryGroup, User } from "../types/stories";

// Global state to persist across components
const storyGroups = ref<StoryGroup[]>([]);
const currentGroupIndex = ref(0);
const currentStoryIndex = ref(0); // Index relative to the current group
const selectedUsers = ref<Set<string>>(new Set());

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

  // --- Computed Properties ---

  // The specific group of stories the user is currently watching
  const currentGroup = computed(() => {
    // FIX: Optional chaining safety
    return storyGroups.value[currentGroupIndex.value] || null;
  });

  const stories = computed(() => currentGroup.value?.stories || []);

  const currentStory = computed(
    () => stories.value[currentStoryIndex.value] || null
  );

  // Helper: Just the stories for the current user (for Progress Bars)
  const currentGroupStories = computed(() => currentGroup.value?.stories || []);

  // --- Fetch & Grouping Logic ---

  const fetchStories = async () => {
    try {
      const response = await storiesApi.getFollowingStories();
      const rawStories = response.data.user_stories || [];

      // 1. Map Data
      const mappedStories: Story[] = rawStories.map((s: any) => {
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

      // 2. Group by User ID
      const groupsMap = new Map<string, StoryGroup>();
      mappedStories.forEach((story) => {
        if (!groupsMap.has(story.userId)) {
          groupsMap.set(story.userId, {
            userId: story.userId,
            username: story.username,
            userAvatar: story.userAvatar,
            isVerified: story.isVerified,
            stories: [],
            hasUnseen: false,
          });
        }
        const group = groupsMap.get(story.userId)!;
        group.stories.push(story);
        if (!story.isViewed) group.hasUnseen = true;
      });

      const newGroups = Array.from(groupsMap.values());

      newGroups.forEach((g) =>
        g.stories.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime())
      );

      newGroups.sort((a, b) => (b.hasUnseen ? 1 : 0) - (a.hasUnseen ? 1 : 0));

      storyGroups.value = newGroups;
    } catch (error) {
      console.error("Failed to fetch stories", error);
    }
  };

  const selectGroup = (index: number) => {
    if (storyGroups.value[index]) {
      currentGroupIndex.value = index;

      // Find first unseen story in this user's list
      const firstUnseen = storyGroups.value[index].stories.findIndex(
        (s) => !s.isViewed
      );
      currentStoryIndex.value = firstUnseen !== -1 ? firstUnseen : 0;

      isPlaying.value = true;
    }
  };

  // --- Navigation Logic ---

  const nextStory = () => {
    stopProgress();

    // Case 1: More stories in THIS group?
    if (currentStoryIndex.value < stories.value.length - 1) {
      currentStoryIndex.value++;
    }
    // Case 2: Jump to NEXT User Group?
    else if (currentGroupIndex.value < storyGroups.value.length - 1) {
      currentGroupIndex.value++;
      currentStoryIndex.value = 0; // Start from their first story
    }
    // Case 3: End of everything
    else {
      isPlaying.value = false;
    }
  };

  const previousStory = () => {
    stopProgress();

    // Case 1: Back in this group
    if (currentStoryIndex.value > 0) {
      currentStoryIndex.value--;
    }
    // Case 2: Back to PREVIOUS User Group
    else if (currentGroupIndex.value > 0) {
      currentGroupIndex.value--;
      // Go to the LAST story of that user
      currentStoryIndex.value =
        (storyGroups.value[currentGroupIndex.value]?.stories.length || 1) - 1;
    }
  };

  // --- Utility Functions ---

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

  // Watch for story changes to handle viewed status and progress reset
  watch(currentStory, (newStory) => {
    if (newStory) {
      startProgress();
      if (!newStory.isViewed) {
        storiesApi.viewStory(newStory.id).catch(console.error);
        newStory.isViewed = true;

        // FIX: Ensure currentGroup.value exists before accessing
        if (currentGroup.value) {
          currentGroup.value.hasUnseen = currentGroup.value.stories.some(
            (s) => !s.isViewed
          );
        }
      }
    }
  });

  onUnmounted(() => stopProgress());

  return {
    storyGroups,
    currentGroupIndex,
    currentStoryIndex,
    currentStory,
    currentGroupStories,
    progress,
    isPlaying,
    storyReplyText,
    suggestedUsers,
    selectedUsers,
    fetchStories,
    selectGroup,
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
