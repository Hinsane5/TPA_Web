import { ref, computed } from "vue";
import type { Story, User } from "../types/stories";

export function useStories() {
  // Mock stories data
  const stories = ref<Story[]>([
    {
      id: "1",
      userId: "user1",
      username: "honkastarrarl",
      userAvatar: "👤",
      isVerified: true,
      imageUrl: "https://via.placeholder.com/400x600?text=Story+1",
      timestamp: new Date(Date.now() - 3600000),
      viewedBy: [],
      replies: [],
      likes: 234,
      isLiked: false,
    },
    {
      id: "2",
      userId: "user2",
      username: "user_two",
      userAvatar: "👤",
      isVerified: false,
      imageUrl: "https://via.placeholder.com/400x600?text=Story+2",
      timestamp: new Date(Date.now() - 7200000),
      viewedBy: [],
      replies: [],
      likes: 456,
      isLiked: false,
    },
    {
      id: "3",
      userId: "user3",
      username: "user_three",
      userAvatar: "👤",
      isVerified: false,
      imageUrl: "https://via.placeholder.com/400x600?text=Story+3",
      timestamp: new Date(Date.now() - 10800000),
      viewedBy: [],
      replies: [],
      likes: 789,
      isLiked: false,
    },
    {
      id: "4",
      userId: "user4",
      username: "user_four",
      userAvatar: "👤",
      isVerified: false,
      imageUrl: "https://via.placeholder.com/400x600?text=Story+4",
      timestamp: new Date(Date.now() - 14400000),
      viewedBy: [],
      replies: [],
      likes: 321,
      isLiked: false,
    },
    {
      id: "5",
      userId: "user5",
      username: "user_five",
      userAvatar: "👤",
      isVerified: false,
      imageUrl: "https://via.placeholder.com/400x600?text=Story+5",
      timestamp: new Date(Date.now() - 18000000),
      viewedBy: [],
      replies: [],
      likes: 654,
      isLiked: false,
    },
  ]);

  // Mock suggested users for share modal
  const suggestedUsers = ref<User[]>([
    {
      id: "s1",
      username: "sher",
      fullName: "sher 夏妙玉",
      userAvatar: "👤",
      isVerified: false,
    },
    {
      id: "s2",
      username: "perry",
      fullName: "perry",
      userAvatar: "👤",
      isVerified: false,
    },
    {
      id: "s3",
      username: "SigmaLigmaBalls",
      fullName: "SigmaLigmaBalls",
      userAvatar: "👤",
      isVerified: false,
    },
    {
      id: "s4",
      username: "Vincent Lee",
      fullName: "Vincent Lee",
      userAvatar: "👤",
      isVerified: false,
    },
    {
      id: "s5",
      username: "Leon",
      fullName: "Leon",
      userAvatar: "👤",
      isVerified: false,
    },
    {
      id: "s6",
      username: "alexander christian",
      fullName: "alexander christian",
      userAvatar: "👤",
      isVerified: false,
    },
    {
      id: "s7",
      username: "gershwin.lee",
      fullName: "gershwin.lee",
      userAvatar: "👤",
      isVerified: false,
    },
    {
      id: "s8",
      username: "wendyajaah",
      fullName: "wendyajaah",
      userAvatar: "👤",
      isVerified: false,
    },
  ]);

  const currentStoryIndex = ref(0);
  const selectedUsers = ref<Set<string>>(new Set());
  const storyReplyText = ref("");

  const currentStory = computed(
    () => stories.value[currentStoryIndex.value] || null
  );

  const nextStory = () => {
    if (currentStoryIndex.value < stories.value.length - 1) {
      currentStoryIndex.value++;
    }
  };

  const previousStory = () => {
    if (currentStoryIndex.value > 0) {
      currentStoryIndex.value--;
    }
  };

  const toggleLike = () => {
    if (currentStory.value) {
      currentStory.value.isLiked = !currentStory.value.isLiked;
      currentStory.value.likes += currentStory.value.isLiked ? 1 : -1;
    }
  };

  const toggleUserSelection = (userId: string) => {
    if (selectedUsers.value.has(userId)) {
      selectedUsers.value.delete(userId);
    } else {
      selectedUsers.value.add(userId);
    }
  };

  const sendStory = () => {
    const recipients = Array.from(selectedUsers.value);
    console.log("Sending story to:", recipients);
    selectedUsers.value.clear();
  };

  const addReply = () => {
    if (storyReplyText.value.trim() && currentStory.value) {
      const newReply = {
        id: Date.now().toString(),
        userId: "currentUser",
        username: "You",
        userAvatar: "👤",
        message: storyReplyText.value,
        timestamp: new Date(),
      };
      currentStory.value.replies.push(newReply);
      storyReplyText.value = "";
    }
  };

  return {
    stories,
    currentStoryIndex,
    currentStory,
    storyReplyText,
    selectedUsers,
    suggestedUsers,
    nextStory,
    previousStory,
    toggleLike,
    toggleUserSelection,
    sendStory,
    addReply,
  };
}
